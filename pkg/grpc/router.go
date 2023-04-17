/*
 * Copyright 2022-2023 Exactpro (Exactpro Systems Limited)
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

package grpc

import (
	"errors"
	"fmt"
	"net"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewRouter(config Config, logger zerolog.Logger) Router {
	return &commonGrpcRouter{
		Config:    config,
		connCache: newConnectionCache(),
		ZLogger:   logger,
	}
}

/*
checks whether there is just 1 endpoint in all the pins in the config.
mainly intended for first release
*/
type commonGrpcRouter struct {
	Config    Config
	connCache connectionCache
	ZLogger   zerolog.Logger
}

func (gr *commonGrpcRouter) createListener() (net.Listener, error) {
	address := gr.Config.getServerAddress()

	listener, netErr := net.Listen("tcp", address)
	if netErr != nil {
		gr.ZLogger.Error().Err(netErr).Msgf("could not start listening to the network: %s", netErr.Error())
		return nil, netErr
	}
	gr.ZLogger.Debug().Msgf("listening on address: %v", address)

	return listener, nil
}

func (gr *commonGrpcRouter) createServerWithRegisteredService(registrar func(grpc.ServiceRegistrar)) *grpc.Server {
	s := grpc.NewServer()
	registrar(s)
	gr.ZLogger.Debug().Msgf("Created server")
	return s
}

func (gr *commonGrpcRouter) Close() error {
	for name, service := range gr.Config.ServicesMap {
		for endpointName, endpoint := range service.Endpoints {
			con, exists := gr.connCache.Get(endpoint.Address)
			if exists {
				if err := con.Close(); err != nil {
					gr.ZLogger.Error().
						Err(err).
						Str("serviceName", name).
						Str("endpointName", endpointName).
						Msg("cannot close connection")

				}
			}
		}
		gr.ZLogger.Debug().Str("serviceName", name).
			Msg("connections for service closed")
	}
	gr.ZLogger.Info().Msg("grpc router closed")
	return nil
}

func (gr *commonGrpcRouter) StartServer(registrar func(grpc.ServiceRegistrar)) error {
	listener, netErr := gr.createListener()
	if netErr != nil {
		return netErr
	}

	s := gr.createServerWithRegisteredService(registrar)
	if serveErr := s.Serve(listener); serveErr != nil {
		gr.ZLogger.Error().Err(serveErr).Msgf("error reading gRPC requests: %s", serveErr.Error())
		return serveErr
	}
	gr.ZLogger.Debug().Msg("server started")
	return nil
}

func (gr *commonGrpcRouter) StartServerAsync(registrar func(grpc.ServiceRegistrar)) (StopServer, error) {
	listener, netErr := gr.createListener()
	if netErr != nil {
		return nil, netErr
	}

	s := gr.createServerWithRegisteredService(registrar)

	go func() {
		if serveErr := s.Serve(listener); serveErr != nil {
			gr.ZLogger.Panic().Err(serveErr).Msgf("error reading gRPC requests: %s", serveErr.Error())
		}
	}()

	return s.GracefulStop, nil
}

type connError struct {
	specificErr error
}

func (ce connError) make() error {
	return errors.New(
		fmt.Sprint("could not create a connection to the given target: ", ce.specificErr.Error()))
}

func (gr *commonGrpcRouter) GetConnection(ServiceName string) (grpc.ClientConnInterface, error) {
	if gr.connCache == nil {
		gr.connCache = newConnectionCache()
	}
	addr, findErr := gr.Config.findEndpointAddrViaServiceName(ServiceName)
	if findErr != nil {
		return nil, connError{specificErr: findErr}.make()
	}
	if conn, exists := gr.findConnection(addr); exists {
		gr.ZLogger.Debug().Msgf("connection for address %v was found", addr)
		return conn, nil
	}
	gr.ZLogger.Error().Msgf("couldn't found connection for address %v ", addr)

	return gr.newConnection(addr)
}

func (gr *commonGrpcRouter) findConnection(addr Address) (grpc.ClientConnInterface, bool) {
	return gr.connCache.Get(addr)
}

func (gr *commonGrpcRouter) newConnection(addr Address) (grpc.ClientConnInterface, error) {
	validationErr := gr.Config.ValidatePins()
	if validationErr != nil {
		return nil, connError{specificErr: validationErr}.make()
	}

	conn, dialErr := grpc.Dial(addr.AsColonSeparatedString(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if dialErr != nil {
		return nil, connError{specificErr: dialErr}.make()
	}

	gr.connCache.Put(addr, conn)
	gr.ZLogger.Debug().Msgf("Created new connection for address : %v", addr)

	return conn, nil

}
