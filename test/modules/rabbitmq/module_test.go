/*
 * Copyright 2023 Exactpro (Exactpro Systems Limited)
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

package rabbitmq

import (
	"encoding/json"
	"github.com/th2-net/th2-common-go/pkg/modules/queue"
	"github.com/th2-net/th2-common-go/test/modules/internal"
	rabbitInternal "github.com/th2-net/th2-common-go/test/modules/rabbitmq/internal"
	"testing"
	"testing/fstest"
)

func TestCanRegisterMqModule(t *testing.T) {
	if testing.Short() {
		t.Skip("do not run containers in short run")
		return
	}
	cfg := rabbitInternal.StartMq(t, "amq.direct")
	connectionCfg, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	mqCfg := `{
	  "queues": {
		"pin":  {
		  "attributes": [
			"group",
			"publish"
		  ],
		  "exchange": "",
		  "filters": [],
		  "name": "",
		  "queue": ""
		}
	  }
	}`

	factory := internal.CreateTestFactory(fstest.MapFS{
		"rabbitMQ": &fstest.MapFile{
			Data: connectionCfg,
		},
		"mq": &fstest.MapFile{
			Data: []byte(mqCfg),
		},
	})

	err = factory.Register(queue.NewRabbitMqModule)
	if err != nil {
		t.Fatal(err)
	}
	var mod queue.Module
	mod, err = queue.ModuleID.GetModule(factory)
	if err != nil {
		t.Fatal(err)
	}
	defer mod.Close()
	if mod == nil {
		t.Fatal("module is nil")
	}
}
