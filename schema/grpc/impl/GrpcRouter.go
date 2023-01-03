/*
 * Copyright 2022 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package impl

import (
	"errors"
	"fmt"
	routerInterface "github.com/th2-net/th2-common-go/schema/grpc"
	"github.com/th2-net/th2-common-go/schema/grpc/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
)

/*
checks whether there is just 1 endpoint in all the pins in the config.
mainly intended for first release
*/

type CommonGrpcRouter struct {
	Config    config.GrpcConfig
	connCache ConnectionCache
}

func (gr *CommonGrpcRouter) createListener() (net.Listener, error) {
	address := gr.Config.GetServerAddress()

	listener, netErr := net.Listen("tcp", address)
	if netErr != nil {
		return nil, errors.New(
			fmt.Sprintf("could not start listening to the network: %s", netErr.Error()))
	}

	return listener, nil
}

func (gr *CommonGrpcRouter) createServerWithRegisteredService(registrar func(grpc.ServiceRegistrar)) *grpc.Server {
	s := grpc.NewServer()
	registrar(s)
	return s
}

func (gr *CommonGrpcRouter) StartServer(registrar func(grpc.ServiceRegistrar)) error {
	listener, netErr := gr.createListener()
	if netErr != nil {
		return netErr
	}

	s := gr.createServerWithRegisteredService(registrar)
	if serveErr := s.Serve(listener); serveErr != nil {
		return errors.New(fmt.Sprintf("error reading gRPC requests: %s", serveErr.Error()))
	}

	return nil
}

func (gr *CommonGrpcRouter) StartServerAsync(registrar func(grpc.ServiceRegistrar)) (routerInterface.StopServer, error) {
	listener, netErr := gr.createListener()
	if netErr != nil {
		return nil, netErr
	}

	s := gr.createServerWithRegisteredService(registrar)

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
		gr.connCache = InitConnectionAddressKeyCache()
	}

	addr, findErr := gr.Config.FindEndpointAddrViaAttributes(attributes)
	if findErr != nil {
		return nil, connError{specificErr: findErr}.make()
	}

	if conn, exists := gr.findConnection(addr); exists {
		return conn, nil
	}

	return gr.newConnection(addr)
}

func (gr *CommonGrpcRouter) findConnection(addr config.Address) (grpc.ClientConnInterface, bool) {
	return gr.connCache.get(addr)
}

func (gr *CommonGrpcRouter) newConnection(addr config.Address) (grpc.ClientConnInterface, error) {
	validationErr := gr.Config.ValidatePins()
	if validationErr != nil {
		return nil, connError{specificErr: validationErr}.make()
	}

	conn, dialErr := grpc.Dial(addr.AsColonSeparatedString(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if dialErr != nil {
		return nil, connError{specificErr: dialErr}.make()
	}

	gr.connCache.put(addr, conn)

	return conn, nil

}
