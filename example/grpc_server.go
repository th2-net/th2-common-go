package main

import (
	"context"
	ac "exactpro/th2/example/proto"
	"fmt"
	cg "github.com/th2-net/th2-common-go/proto" //common proto
	"github.com/th2-net/th2-common-go/schema/factory"
	"github.com/th2-net/th2-common-go/schema/modules/grpcModule"
	"google.golang.org/grpc"
	"log"
	"os"
	"reflect"
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
	newFactory := factory.NewFactory(os.Args)
	if err := newFactory.Register(grpcModule.NewGrpcModule); err != nil {
		panic(err)
	}

	module, err := grpcModule.ModuleID.GetModule(newFactory)
	if err != nil {
		panic("no module")
	} else {
		fmt.Println("module found", reflect.TypeOf(module))
	}

	grpcRouter := module.GrpcRouter
	stopServerFunc, err := grpcRouter.StartServerAsync(registerService)
	if err != nil {
		log.Fatalf(err.Error())
	}

	time.Sleep(time.Second * lifetimeSec)

	stopServerFunc()

}
