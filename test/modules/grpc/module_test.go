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

package grpc

import (
	"github.com/th2-net/th2-common-go/pkg/modules/grpc"
	"github.com/th2-net/th2-common-go/test/modules/internal"
	"testing"
	"testing/fstest"
)

func TestCanRegisterGrpc(t *testing.T) {
	cfg := `
    {
		"server": {
		},
		"services": {
		}
	}
    `
	factory := internal.CreateTestFactory(fstest.MapFS{
		"grpc": &fstest.MapFile{
			Data: []byte(cfg),
		},
	})

	err := factory.Register(grpc.NewModule)
	if err != nil {
		t.Fatal(err)
	}
	var mod grpc.Module
	mod, err = grpc.ModuleID.GetModule(factory)
	if err != nil {
		t.Fatal(err)
	}
	if mod == nil {
		t.Fatal("module is nil")
	}
}
