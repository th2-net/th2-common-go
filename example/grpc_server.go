package main

import (
	"context"
	commongrpc "github.com/th2-net/th2-common-go/grpc"
	ac "github.com/th2-net/th2-common-go/grpc/proto"
	cg "github.com/th2-net/th2-common-go/proto" //common proto
	"google.golang.org/grpc"
	"log"
	"time"
)

const (
	grpcJsonFileName = "grpc.json"
)

type server struct {
	ac.UnimplementedActServer
}

func (s *server) PlaceOrderFIX(_ context.Context, in *ac.PlaceMessageRequest) (*ac.PlaceMessageResponse, error) {
	return &ac.PlaceMessageResponse{Status: &cg.RequestStatus{
		Message: "the order has been received",
	},
		ResponseMessage: in.Message}, nil //return the request content as the response
}

func registerService(registrar grpc.ServiceRegistrar) {
	ac.RegisterActServer(registrar, &server{})
}

const lifetimeSec = 30

func main() {
	grpcRouter := commongrpc.CommonGrpcRouter{Config: commongrpc.GrpcConfig{}}
	cp := &commongrpc.ConfigProviderFromFile{DirectoryPath: "../resources"}
	cfgErr := cp.GetConfig(grpcJsonFileName, &grpcRouter.Config)
	if cfgErr != nil {
		log.Fatalf(cfgErr.Error())
	}
	stopServerFunc, err := grpcRouter.StartServerAsync(registerService)
	if err != nil {
		log.Fatalf(err.Error())
	}

	time.Sleep(time.Second * lifetimeSec)

	stopServerFunc()

}
