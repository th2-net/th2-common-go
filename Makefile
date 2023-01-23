TARGET_DIR?=$(shell pwd)
GITHUB_GROUP=github.com/th2-net

TH2_GRPC_COMMON=th2-grpc-common
TH2_GRPC_COMMON_URL=$(GITHUB_GROUP)/$(TH2_GRPC_COMMON)@makefile

clean-deps:
	-rm go.work

deps: clean-deps 
	go work init .

	go get -u -t $(TH2_GRPC_COMMON_URL)
	sleep 1
	@cd $(shell go list -m -f '{{.Dir}}' $(TH2_GRPC_COMMON_URL)) && make generate-module TARGET_DIR=$(TARGET_DIR)

test:
	go test ./...