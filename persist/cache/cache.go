package cache

import (
	"fmt"
	"time"
)

type Cache interface {
	Put(key string, val interface{}, duration time.Duration) error
	Get(key string) interface{}
	DefaultGet(key string, defaultValue interface{}) interface{}
	Delete(key string) error
	StartAndGC(config string) error
}

type Instance func() Cache

var adapters = make(map[string]Instance)

func Register(adapterName string, instance Instance) {
	if instance == nil {
		panic("Cache:instance is nil,name:" + adapterName)
	}
	adapters[adapterName] = instance
}

func NewCache(adapterName, config string) (Cache, error) {
	instance, ok := adapters[adapterName]
	if !ok {
		return nil, fmt.Errorf("Cache:not found instance, name:%s", adapterName)
	}

	cache := instance()

	err := cache.StartAndGC(config)
	if err != nil {
		return nil, err
	}

	return cache, nil
}
