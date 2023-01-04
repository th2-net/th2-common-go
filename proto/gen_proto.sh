#   Copyright 2020-2022 Exactpro (Exactpro Systems Limited)
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

#! /bin/bash

# Changing GOPATH
TEMP_PATH=$GOPATH
export GOPATH=$PWD/dependencies

# Downloading common.proto from th2-grpc-common
go get github.com/th2-net/th2-grpc-common@go-package

# Generating go code from common.proto
protoc --go_out=. dependencies/pkg/mod/github.com/th2-net/**/src/main/proto/**/*.proto

# Changing the GOPATH back
export GOPATH=TEMP_PATH