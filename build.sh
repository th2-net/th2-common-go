#! /bin/bash


# Generating proto code from .proto files
./proto/gen_proto.sh

# Running the test scripts
go test ./...

# Running the bench test
go test -bench=/bench -benchmem bench/*.go

# Building the code
# go build .