/*
 * Copyright 2022 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package message

import (
	p_buff "github.com/th2-net/th2-common-go/proto"
	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"google.golang.org/protobuf/proto"
	"log"
)

type CommonMessageSender struct {
	ConnManager  *MQcommon.ConnectionManager
	exchangeName string
	sendQueue    string
	th2Pin       string
}

func (sender *CommonMessageSender) Send(batch *p_buff.MessageGroupBatch) error {

	if batch == nil {
		log.Fatalln("Value for send cann't be null")
	}
	body, err := proto.Marshal(batch)
	if err != nil {
		if err != nil {
			log.Panicf("Error during marshaling message into proto message. : %s", err)
		}
		return err
	}

	fail := sender.ConnManager.Publisher.Publish(body, sender.sendQueue, sender.exchangeName)
	if fail != nil {
		return fail
	}
	// th2Pin will be used for Metrics
	return nil
}
