package main

import (
	"context"
	ac "exactpro/th2/th2-common-go/grpc/proto"
	cg "exactpro/th2/th2-common-go/proto" //common proto
	"google.golang.org/grpc"
	"log"
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

func main() {

	grpcRouter := CommonGrpcRouter{Config: GrpcConfig{}}
	cp := &ConfigProviderFromFile{DirectoryPath: "../resources"}
	cfgErr := cp.GetConfig(grpcJsonFileName, &grpcRouter.Config)
	if cfgErr != nil {
		log.Fatalf(cfgErr.Error())
	}
	stopServerFunc, err := grpcRouter.StartServer(registerService)
	if err != nil {
		log.Fatalf(err.Error())
	}

	stopServerFunc() //seems to have no effect?..

	//time.Sleep(time.Second * 3)
}
