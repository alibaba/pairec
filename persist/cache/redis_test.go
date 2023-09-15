package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestRedisInit(t *testing.T) {
	config := "{\"host\":\"127.0.0.1\", \"port\":6379, \"maxIdle\":3, \"password\":\"\"}"

	cache, err := NewCache("redis", config)
	if err != nil {
		t.Error(err)
	}

	if cache == nil {
		t.Errorf("cache is nil")
	}
}
func getCache() (Cache, error) {
	config := "{\"host\":\"127.0.0.1\", \"port\":6379, \"maxIdle\":3, \"password\":\"\"}"

	cache, err := NewCache("redis", config)
	return cache, err
}
func TestRedisPut(t *testing.T) {
	cache, err := getCache()
	if err != nil {
		t.Error(err)
	}

	err = cache.Put("foo", "bar", 10*time.Second)
	if err != nil {
		t.Error(err)
	}

	val := cache.Get("foo")

	fmt.Println(string(val.([]uint8)))

}
