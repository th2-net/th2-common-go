package main

import (
	"google.golang.org/grpc"
	"strings"
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

type ConnectionCache interface {
	put(attributes []string, conn grpc.ClientConnInterface)
	get(attributes []string) (grpc.ClientConnInterface, bool)
}

type ConnectionStringKeyCache struct {
	internalCache ConnectionCacheMapData[string]
}

func initConnectionStringKeyCache() ConnectionCache {
	return &ConnectionStringKeyCache{internalCache: initConnectionCacheMapData[string]()}
}

func (sc *ConnectionStringKeyCache) put(attributes []string, conn grpc.ClientConnInterface) {
	attrsCacheKey := sc.formatForCacheKey(attributes)
	sc.internalCache.put(attrsCacheKey, conn)
}

func (sc *ConnectionStringKeyCache) get(attributes []string) (grpc.ClientConnInterface, bool) {
	attrsCacheKey := sc.formatForCacheKey(attributes)
	return sc.internalCache.get(attrsCacheKey)
}

func (sc *ConnectionStringKeyCache) formatForCacheKey(attributes []string) string {
	return strings.Join(attributes, ", ")
}
