package grpc

import (
	"google.golang.org/grpc"
)

type StopServer func()

type GrpcRouter interface {
	StartServer(registrar func(grpc.ServiceRegistrar)) error
	StartServerAsync(registrar func(grpc.ServiceRegistrar)) (StopServer, error)
	GetConnection(attributes ...string) (grpc.ClientConnInterface, error)
}
