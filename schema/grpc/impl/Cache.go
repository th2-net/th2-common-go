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
