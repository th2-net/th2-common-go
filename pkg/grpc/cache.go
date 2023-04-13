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

package grpc

import (
	"google.golang.org/grpc"
)

type connectionCacheMapData[T comparable] map[T]*grpc.ClientConn

func initConnectionCacheMapData[T comparable]() map[T]*grpc.ClientConn {
	return make(map[T]*grpc.ClientConn)
}

func (cc *connectionCacheMapData[T]) Put(key T, conn *grpc.ClientConn) {
	(*cc)[key] = conn
}

func (cc *connectionCacheMapData[T]) Get(key T) (*grpc.ClientConn, bool) {
	conn, exists := (*cc)[key]

	return conn, exists
}

//the following caching design offers the flexibility of having an internal cache of different type than Address,
//and importantly, insert additional logic in put and get

type connectionCache interface {
	Put(key Address, conn *grpc.ClientConn)
	Get(key Address) (*grpc.ClientConn, bool)
}

type connectionAddressKeyCache struct {
	internalCache connectionCacheMapData[Address]
}

func newConnectionCache() connectionCache {
	return &connectionAddressKeyCache{internalCache: initConnectionCacheMapData[Address]()}
}

func (sc *connectionAddressKeyCache) Put(key Address, conn *grpc.ClientConn) {
	sc.internalCache.Put(key, conn)
}

func (sc *connectionAddressKeyCache) Get(key Address) (*grpc.ClientConn, bool) {
	return sc.internalCache.Get(key)
}
