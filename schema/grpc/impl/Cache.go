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
	"github.com/th2-net/th2-common-go/schema/grpc/config"
	"google.golang.org/grpc"
)

type ConnectionCacheMapData[T comparable] map[T]grpc.ClientConnInterface

func initConnectionCacheMapData[T comparable]() map[T]grpc.ClientConnInterface {
	return make(map[T]grpc.ClientConnInterface)
}

func (cc *ConnectionCacheMapData[T]) put(key T, conn grpc.ClientConnInterface) {
	(*cc)[key] = conn
}

func (cc *ConnectionCacheMapData[T]) get(key T) (grpc.ClientConnInterface, bool) {
	conn, exists := (*cc)[key]

	return conn, exists
}

//the following caching design offers the flexibility of having an internal cache of different type than Address,
//and importantly, insert additional logic in put and get

type ConnectionCache interface {
	put(key config.Address, conn grpc.ClientConnInterface)
	get(key config.Address) (grpc.ClientConnInterface, bool)
}

type ConnectionAddressKeyCache struct {
	internalCache ConnectionCacheMapData[config.Address]
}

func InitConnectionAddressKeyCache() ConnectionCache {
	return &ConnectionAddressKeyCache{internalCache: initConnectionCacheMapData[config.Address]()}
}

func (sc *ConnectionAddressKeyCache) put(key config.Address, conn grpc.ClientConnInterface) {
	sc.internalCache.put(key, conn)
}

func (sc *ConnectionAddressKeyCache) get(key config.Address) (grpc.ClientConnInterface, bool) {
	return sc.internalCache.get(key)
}
