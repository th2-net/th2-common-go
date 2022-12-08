package grpc

import (
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"sort"
	"strings"
)

type Address struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func (addr *Address) asColonSeparatedString() string {
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

// Checks for the inclusion of target attributes
func (gc *GrpcConfig) findEndpointAddrViaAttributes(targetAttributes []string) (Address, error) {
	targetCopy := make([]string, len(targetAttributes))
	copy(targetCopy, targetAttributes)
	sort.Strings(targetCopy)
	separator := ", "
	sortedTargetAttrsStr := strings.Join(targetCopy, separator)

	for _, service := range gc.ServicesMap {
		for _, endpoint := range service.Endpoints {
			endpointAttrs := endpoint.Attributes
			endpointAttrsCopy := make([]string, len(endpointAttrs))
			copy(endpointAttrsCopy, endpointAttrs)
			sort.Strings(endpointAttrsCopy)
			sortedEndpointAttrsStr := strings.Join(endpointAttrsCopy, separator)
			if strings.Contains(sortedEndpointAttrsStr, sortedTargetAttrsStr) {
				return endpoint.Address, nil
			}
		}
	}
	return Address{}, errors.New("endpoint with provided attributes does not exist")
}

func (gc *GrpcConfig) getServerAddress() string {
	return gc.ServerConfig.asColonSeparatedString()
}

type StopServer func()

type GrpcRouter interface {
	StartServer(registrar func(grpc.ServiceRegistrar)) error
	StartServerAsync(registrar func(grpc.ServiceRegistrar)) (StopServer, error)
	GetConnection(attributes ...string) (grpc.ClientConnInterface, error)
}

type CommonGrpcRouter struct {
	Config    GrpcConfig
	connCache ConnectionCache
}

func (gr *CommonGrpcRouter) StartServer(registrar func(grpc.ServiceRegistrar)) error {

	address := gr.Config.getServerAddress()

	listener, netErr := net.Listen("tcp", address)
	if netErr != nil {
		return errors.New(
			fmt.Sprintf("could not start listening to the network: %s", netErr.Error()))
	}

	s := grpc.NewServer()
	registrar(s)
	if serveErr := s.Serve(listener); serveErr != nil {
		return errors.New(fmt.Sprintf("error reading gRPC requests: %s", serveErr.Error()))
	}

	return nil
}

func (gr *CommonGrpcRouter) StartServerAsync(registrar func(grpc.ServiceRegistrar)) (StopServer, error) {

	address := gr.Config.getServerAddress()

	listener, netErr := net.Listen("tcp", address)
	if netErr != nil {
		return nil, errors.New(
			fmt.Sprintf("could not start listening to the network: %s", netErr.Error()))
	}

	s := grpc.NewServer()
	registrar(s)

	go func() {
		if serveErr := s.Serve(listener); serveErr != nil {
			panic(fmt.Sprintf("error reading gRPC requests: %s", serveErr.Error()))
		}
	}()

	return s.GracefulStop, nil
}

type connError struct {
	specificErr error
}

func (ce connError) make() error {
	return errors.New(
		fmt.Sprint("could not create a connection to the given target: ", ce.specificErr.Error()))
}

func (gr *CommonGrpcRouter) GetConnection(attributes ...string) (grpc.ClientConnInterface, error) {
	if gr.connCache == nil {
		gr.connCache = initConnectionAddressKeyCache()
	}

	addr, findErr := gr.Config.findEndpointAddrViaAttributes(attributes)
	if findErr != nil {
		return nil, connError{specificErr: findErr}.make()
	}

	if conn, exists := gr.findConnection(addr); exists {
		return conn, nil
	}

	return gr.newConnection(addr)
}

func (gr *CommonGrpcRouter) findConnection(addr Address) (grpc.ClientConnInterface, bool) {
	return gr.connCache.get(addr)
}

func (gr *CommonGrpcRouter) newConnection(addr Address) (grpc.ClientConnInterface, error) {
	validationErr := gr.Config.validatePins()
	if validationErr != nil {
		return nil, connError{specificErr: validationErr}.make()
	}

	conn, dialErr := grpc.Dial(addr.asColonSeparatedString(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if dialErr != nil {
		return nil, connError{specificErr: dialErr}.make()
	}

	gr.connCache.put(addr, conn)

	return conn, nil

}
