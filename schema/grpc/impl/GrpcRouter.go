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
	"github.com/rs/zerolog"
	routerInterface "github.com/th2-net/th2-common-go/schema/grpc"
	"github.com/th2-net/th2-common-go/schema/grpc/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
)

/*
 checks whether there is just 1 endpoint in all the pins in the config.
 mainly intended for first release
*/

type CommonGrpcRouter struct {
	Config    config.GrpcConfig
	connCache ConnectionCache
	ZLogger   zerolog.Logger
}

func (gr *CommonGrpcRouter) createListener() (net.Listener, error) {
	address := gr.Config.GetServerAddress()

	listener, netErr := net.Listen("tcp", address)
	if netErr != nil {
		gr.ZLogger.Error().Err(netErr).Msgf("could not start listening to the network: %s", netErr.Error())
		return nil, netErr
	}
	gr.ZLogger.Debug().Msgf("listening on address: %v", address)

	return listener, nil
}

func (gr *CommonGrpcRouter) createServerWithRegisteredService(registrar func(grpc.ServiceRegistrar)) *grpc.Server {
	s := grpc.NewServer()
	registrar(s)
	gr.ZLogger.Debug().Msgf("Created server")
	return s
}

func (gr *CommonGrpcRouter) Close() {
	//gr.connCache.get(gr.Config.GetServerAddress()).Close()
	log.Println("closing grpc router")
}

func (gr *CommonGrpcRouter) StartServer(registrar func(grpc.ServiceRegistrar)) error {
	listener, netErr := gr.createListener()
	if netErr != nil {
		return netErr
	}

	s := gr.createServerWithRegisteredService(registrar)
	if serveErr := s.Serve(listener); serveErr != nil {
		gr.ZLogger.Error().Err(serveErr).Msgf("error reading gRPC requests: %s", serveErr.Error())
		return serveErr
	}
	gr.ZLogger.Debug().Msg("server started")
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
			gr.ZLogger.Panic().Err(serveErr).Msgf("error reading gRPC requests: %s", serveErr.Error())
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

func (gr *CommonGrpcRouter) GetConnection(ServiceName string) (grpc.ClientConnInterface, error) {
	if gr.connCache == nil {
		gr.connCache = InitConnectionAddressKeyCache()
	}
	addr, findErr := gr.Config.FindEndpointAddrViaServiceName(ServiceName)
	if findErr != nil {
		return nil, connError{specificErr: findErr}.make()
	}
	if conn, exists := gr.findConnection(addr); exists {
		gr.ZLogger.Debug().Msgf("connection for address %v was found", addr)
		return conn, nil
	}
	gr.ZLogger.Error().Msgf("couldn't found connection for address %v ", addr)

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
	gr.ZLogger.Debug().Msgf("created nec connection for address : %v", addr)

	return conn, nil

}
