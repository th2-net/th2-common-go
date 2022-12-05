package main

import (
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
)

type Address struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func (addr *Address) getColonSeparatedAddr() string {
	return fmt.Sprint(addr.Host, ":", addr.Port)
}

type Endpoint struct {
	Attributes []string `json:"attributes"`
	Address
}

type Service struct {
	Endpoints map[string]Endpoint `json:"endpoints"`
	//others if needed..
}

type Services map[string]Service

type GrpcConfig struct {
	ServerConfig Address  `json:"server"`
	ServicesMap  Services `json:"services"`
}

/*
checks whether there is just 1 endpoint in all the pins in the config.
mainly intended for first release
*/

func (gc *GrpcConfig) validatePins() error {
	for pinName, service := range gc.ServicesMap {
		if len(service.Endpoints) > 1 {
			return errors.New(fmt.Sprintf(
				`config is invalid. pin "%s" has more than 1 endpoint`, pinName))
		}
	}
	return nil
}

func (gc *GrpcConfig) findEndpointAddrViaAttributes(targetAttributes []string) (string, error) {
	for _, service := range gc.ServicesMap {
		for _, endpoint := range service.Endpoints {
			if cmp.Equal(endpoint.Attributes, targetAttributes) {
				return endpoint.Address.getColonSeparatedAddr(), nil
			}
		}
	}
	return "", errors.New("endpoint with provided attributes does not exist")
}

func (gc *GrpcConfig) getServerAddress() string {
	return gc.ServerConfig.getColonSeparatedAddr()
}

type StopServer func()

type GrpcRouter interface {
	StartServer(registrar func(grpc.ServiceRegistrar)) (StopServer, error)
	GetConnection(attributes ...string) (grpc.ClientConnInterface, error)
}

type CommonGrpcRouter struct {
	Config    GrpcConfig
	connCache ConnectionCache
}

func (gr *CommonGrpcRouter) StartServer(registrar func(grpc.ServiceRegistrar)) (StopServer, error) {

	address := gr.Config.getServerAddress()

	listener, netErr := net.Listen("tcp", address)
	if netErr != nil {
		return nil, errors.New(
			fmt.Sprintf("could not start listening to the network: %s", netErr.Error()))
	}

	s := grpc.NewServer()
	registrar(s)
	if err := s.Serve(listener); err != nil {
		return nil, errors.New(fmt.Sprintf("error reading gRPC requests: %s", err.Error()))
	}

	return s.Stop, nil
}

func (gr *CommonGrpcRouter) GetConnection(attributes ...string) (grpc.ClientConnInterface, error) {
	if gr.connCache == nil {
		gr.connCache = initConnectionStringKeyCache()
	}
	if conn, exists := gr.findConnection(attributes); exists {
		return conn, nil
	}

	return gr.newConnection(attributes)
}

func (gr *CommonGrpcRouter) findConnection(attributes []string) (grpc.ClientConnInterface, bool) {
	return gr.connCache.get(attributes)
}

func (gr *CommonGrpcRouter) newConnection(attributes []string) (grpc.ClientConnInterface, error) {
	makeErr := func(specificErr error) error {
		return errors.New(
			fmt.Sprint("could not create a connection to the given target: ", specificErr.Error()))
	}

	validationErr := gr.Config.validatePins()
	if validationErr != nil {
		return nil, makeErr(validationErr)
	}

	addr, findErr := gr.Config.findEndpointAddrViaAttributes(attributes)
	if findErr != nil {
		return nil, makeErr(findErr)
	}
	conn, dialErr := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if dialErr != nil {
		return nil, makeErr(dialErr)
	}

	gr.connCache.put(attributes, conn)

	return conn, nil

}
