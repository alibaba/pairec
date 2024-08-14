package redisdb

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	Address  string
	Password string
	DbNum    int
	client   *redis.Client
}

var redisPools = make(map[string]*Redis)

func GetRedis(name string) (*Redis, error) {
	if _, ok := redisPools[name]; !ok {
		return nil, fmt.Errorf("redis:not found, name:%s", name)
	}

	return redisPools[name], nil
}

func (r *Redis) Init() error {

	rdb := redis.NewClient(&redis.Options{
		Addr:        r.Address,
		Password:    r.Password, // no password set
		DB:          r.DbNum,    // use default DB
		MaxRetries:  1,
		DialTimeout: time.Second,
		ReadTimeout: time.Second,
		PoolSize:    1000,
		MaxConnAge:  30 * time.Minute,
	})

	r.client = rdb
	_, err := rdb.Ping(context.Background()).Result()
	return err
}
func (r *Redis) GetClient() *redis.Client {
	return r.client
}
func RegisterRedis(name, address, password string, database int) {
	if _, ok := redisPools[name]; !ok {
		m := &Redis{
			Address:  address,
			Password: password,
			DbNum:    database,
		}
		err := m.Init()
		if err != nil {
			panic(err)
		}
		redisPools[name] = m
	}

}
