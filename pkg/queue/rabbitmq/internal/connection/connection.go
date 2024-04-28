/*
 * Copyright 2024 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package connection

import (
	"errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/connection"
	"sync"
	"time"
)

const (
	defaultMinRecoveryTimeout = 1 * time.Second
	defaultMaxRecoveryTimeout = 60 * time.Second
	// defaultMaxRecoveryAttempts used in case an error with status NOT_FOUND is returned from channel
	defaultMaxRecoveryAttempts = 5
)

type connectionHolder struct {
	connMutex             sync.RWMutex
	conn                  *amqp.Connection
	channels              map[string]*amqp.Channel
	done                  chan struct{}
	reconnectToMq         func() (*amqp.Connection, error)
	onConnectionRecovered func()
	onChannelRecovered    func(channelKey string)
	logger                zerolog.Logger
	notifyMutex           sync.Mutex
	notifyRecovered       []chan struct{}
	minRecoveryTimeout    time.Duration
	maxRecoveryTimeout    time.Duration
}

func newConnection(url string, name string, logger zerolog.Logger,
	configuration connection.Config,
	onConnectionRecovered func(), onChannelRecovered func(channelKey string)) (*connectionHolder, error) {
	if configuration.MinConnectionRecoveryTimeout > configuration.MaxConnectionRecoveryTimeout {
		return nil, errors.New("min connection recovery timeout is greater than max connection recovery timeout")
	}
	var minRecoveryTimeout time.Duration
	var maxRecoveryTimeout time.Duration
	if configuration.MinConnectionRecoveryTimeout > 0 {
		minRecoveryTimeout = time.Duration(configuration.MinConnectionRecoveryTimeout) * time.Millisecond
	} else {
		minRecoveryTimeout = defaultMinRecoveryTimeout
	}
	if configuration.MaxConnectionRecoveryTimeout > 0 {
		maxRecoveryTimeout = time.Duration(configuration.MaxConnectionRecoveryTimeout) * time.Millisecond
	} else {
		maxRecoveryTimeout = defaultMaxRecoveryTimeout
	}
	logger.Info().
		Dur("minRecoveryTimeout", minRecoveryTimeout).
		Dur("maxRecoveryTimeout", maxRecoveryTimeout).
		Msg("recovery timeouts configured")
	conn, err := dial(url, name)
	if err != nil {
		return nil, err
	}
	return &connectionHolder{
		connMutex: sync.RWMutex{},
		conn:      conn,
		channels:  make(map[string]*amqp.Channel),
		done:      make(chan struct{}),
		reconnectToMq: func() (*amqp.Connection, error) {
			return dial(url, name)
		},
		onConnectionRecovered: onConnectionRecovered,
		onChannelRecovered:    onChannelRecovered,
		logger:                logger,
		notifyMutex:           sync.Mutex{},
		notifyRecovered:       make([]chan struct{}, 0),
		minRecoveryTimeout:    minRecoveryTimeout,
		maxRecoveryTimeout:    maxRecoveryTimeout,
	}, nil
}

func (c *connectionHolder) runConnectionRoutine() {
	run := true
	connectionClosed := true
	var connectionErrors chan *amqp.Error
	for run {
		if connectionClosed {
			connectionClosed = false
			c.connMutex.RLock()
			connectionErrors = c.conn.NotifyClose(make(chan *amqp.Error))
			c.connMutex.RUnlock()
		}
		select {
		case <-c.done:
			c.logger.Info().
				Msg("stopping connection routine")
			run = false
			break
		case connErr, ok := <-connectionErrors:
			if !ok {
				// normal close
				run = false
				break
			}
			connectionClosed = true
			c.logger.Error().
				Err(connErr).
				Msg("received connection error. reconnecting")
			c.tryToReconnect()
			if c.onConnectionRecovered != nil {
				c.onConnectionRecovered()
			}
			c.notifyMutex.Lock()
			for _, ch := range c.notifyRecovered {
				close(ch)
			}
			c.notifyRecovered = c.notifyRecovered[:0]
			c.notifyMutex.Unlock()
		}
	}
}

func (c *connectionHolder) tryToReconnect() {
	var delay = c.minRecoveryTimeout
	for {
		err := c.reconnect()
		if err == nil {
			c.logger.Info().
				Msg("connection to rabbitmq restored")
			break
		}
		c.logger.Error().
			Err(err).
			Dur("timeout", delay).
			Msg("reconnect failed. retrying after timeout")
		time.Sleep(delay)
		delay *= 2
		if delay > c.maxRecoveryTimeout {
			delay = c.maxRecoveryTimeout
		}
	}
}

func (c *connectionHolder) reconnect() (err error) {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()
	conn := c.conn
	if conn != nil {
		_ = conn.Close()
		// clear map with channels
		c.channels = make(map[string]*amqp.Channel)
	}
	conn, err = c.reconnectToMq()
	if err == nil {
		c.conn = conn
	}
	return
}

func (c *connectionHolder) registerBlockingListener(blocking chan amqp.Blocking) <-chan amqp.Blocking {
	c.connMutex.RLock()
	defer c.connMutex.RUnlock()
	return c.conn.NotifyBlocked(blocking)
}

func dial(url string, name string) (*amqp.Connection, error) {
	properties := amqp.NewConnectionProperties()
	properties.SetClientConnectionName(name)
	conn, err := amqp.DialConfig(url, amqp.Config{
		Heartbeat:  30 * time.Second,
		Locale:     "en_US",
		Properties: properties,
	})
	return conn, err
}

func (c *connectionHolder) Close() error {
	close(c.done)
	c.connMutex.RLock()
	defer c.connMutex.RUnlock()
	return c.conn.Close()
}

func (c *connectionHolder) waitRecovered(ch chan struct{}) <-chan struct{} {
	c.connMutex.RLock()
	if !c.conn.IsClosed() {
		close(ch)
		c.connMutex.RUnlock()
		return ch
	}
	c.connMutex.RUnlock()

	c.notifyMutex.Lock()
	c.notifyRecovered = append(c.notifyRecovered, ch)
	c.notifyMutex.Unlock()
	return ch
}

func (c *connectionHolder) getChannel(key string) (*amqp.Channel, error) {
	var ch *amqp.Channel
	var err error
	var exists bool
	<-c.waitRecovered(make(chan struct{}))
	c.connMutex.RLock()
	ch, exists = c.channels[key]
	c.connMutex.RUnlock()
	if !exists {
		ch, err = c.getOrCreateChannel(key)
	}

	return ch, err
}

func (c *connectionHolder) getOrCreateChannel(key string) (*amqp.Channel, error) {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()
	var ch *amqp.Channel
	var err error
	var exists bool
	ch, exists = c.channels[key]
	if !exists {
		ch, err = c.conn.Channel()
		if err != nil {
			return nil, err
		}
		c.channels[key] = ch
		go func(ch *amqp.Channel) {
			select {
			case err, ok := <-ch.NotifyClose(make(chan *amqp.Error)):
				if !ok {
					break
				}
				c.connMutex.Lock()
				c.logger.Warn().
					Err(err).
					Str("channelKey", key).
					Msg("removing cached channel")
				delete(c.channels, key)
				c.connMutex.Unlock()
				if c.onChannelRecovered != nil {
					c.onChannelRecovered(key)
				}
			case <-c.done:
				// closed. do nothing
			}

		}(ch)
	}
	return ch, nil
}
