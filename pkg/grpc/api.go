package grpc

import (
	"google.golang.org/grpc"
	"io"
)

type StopServer func()

type Router interface {
	StartServer(registrar func(grpc.ServiceRegistrar)) error
	StartServerAsync(registrar func(grpc.ServiceRegistrar)) (StopServer, error)
	GetConnection(serviceName string) (grpc.ClientConnInterface, error)
	io.Closer
}
