#! /bin/bash


# Generating code from proto in /proto
cd proto

# Removing common.proto from /proto if it exists there
if [ -f common.proto ]; then
   rm common.proto
fi
# Removing common.pb.go from /proto if it exists there
if [ -f common.pb.go ]; then
   rm common.pb.go
fi

# Getting common.proto from grpc-common and generating go code from it
./gen_proto.sh

cd ..

# Running the test scripts
go test ./schema ./grpc

# Running the bench test
go test -bench=/bench -benchmem bench/bench_test.go

# Building the code
# go build .