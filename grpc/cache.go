package main

import (
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
	put(key Address, conn grpc.ClientConnInterface)
	get(key Address) (grpc.ClientConnInterface, bool)
}

type ConnectionAddressKeyCache struct {
	internalCache ConnectionCacheMapData[Address]
}

func initConnectionAddressKeyCache() ConnectionCache {
	return &ConnectionAddressKeyCache{internalCache: initConnectionCacheMapData[Address]()}
}

func (sc *ConnectionAddressKeyCache) put(key Address, conn grpc.ClientConnInterface) {
	sc.internalCache.put(key, conn)
}

func (sc *ConnectionAddressKeyCache) get(key Address) (grpc.ClientConnInterface, bool) {
	return sc.internalCache.get(key)
}
