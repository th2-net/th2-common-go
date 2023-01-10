TARGET_DIR?=$(shell pwd)

clean-deps:
	-rm go.work

deps: clean-deps
	go get -u -t github.com/th2-net/th2-grpc-common@makefile
	@cd $(shell go list -m -f '{{.Dir}}' github.com/th2-net/th2-grpc-common@makefile) && make generate-module TARGET_DIR=$(TARGET_DIR)

test:
	go test ./...