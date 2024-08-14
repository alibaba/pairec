package cache

import (
	"encoding/json"
	"time"

	"github.com/patrickmn/go-cache"
)

type LocalCache struct {
	localCache *cache.Cache
}

func NewLocalCache() Cache {
	return &LocalCache{}
}

func (c *LocalCache) Put(key string, val interface{}, d time.Duration) error {
	c.localCache.Set(key, val, d)
	return nil
}
func (c *LocalCache) Get(key string) interface{} {
	if value, found := c.localCache.Get(key); found {
		return value
	}
	return nil

}
func (c *LocalCache) DefaultGet(key string, defaultValue interface{}) interface{} {

	if value, found := c.localCache.Get(key); found {
		return value
	} else {
		return defaultValue
	}
}
func (c *LocalCache) Delete(key string) error {
	c.localCache.Delete(key)
	return nil
}

// config like {"defaultExpiration":1800, "cleanupInterval":1800}
func (c *LocalCache) StartAndGC(config string) error {
	var data struct {
		DefaultExpiration int64
		CleanupInterval   int64
	}

	data.DefaultExpiration = 0
	data.CleanupInterval = 0

	err := json.Unmarshal([]byte(config), &data)
	if err != nil {
		return err
	}

	c.localCache = cache.New(time.Duration(data.DefaultExpiration)*time.Second, time.Duration(data.CleanupInterval)*time.Second)
	return nil
}
func init() {
	Register("localCache", NewLocalCache)
}
