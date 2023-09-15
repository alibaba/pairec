package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

type Redis struct {
	pool     *redis.Pool
	Host     string
	Port     int
	Password string
	MaxIdle  int
}

func NewRedis() Cache {
	return &Redis{}
}

func init() {
	Register("redis", NewRedis)
}

// config format like {"host":"127.0.0.1", "port":6379, "maxIdle":3, "password":""}
func (r *Redis) StartAndGC(config string) error {

	err := json.Unmarshal([]byte(config), r)
	if err != nil {
		return err
	}

	dialFunc := func() (redis.Conn, error) {
		addr := fmt.Sprintf("%s:%d", r.Host, r.Port)
		conn, err := redis.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}

		if len(r.Password) != 0 {
			_, err = conn.Do("AUTH", r.Password)
			if err != nil {
				return nil, err
			}
		}
		return conn, nil
	}

	r.pool = &redis.Pool{
		MaxIdle:     r.MaxIdle,
		IdleTimeout: 180 * time.Second,
		Dial:        dialFunc,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	c := r.pool.Get()
	defer c.Close()

	return c.Err()
}
func (r *Redis) Put(key string, val interface{}, duration time.Duration) error {
	_, err := r.do("SETEX", key, int64(duration/time.Second), val)
	return err
}
func (r *Redis) do(commandName string, args ...interface{}) (reply interface{}, err error) {
	if len(args) == 0 {
		err = errors.New("missing require argument")
		return
	}
	conn := r.pool.Get()
	defer conn.Close()
	reply, err = conn.Do(commandName, args...)

	return
}
func (r *Redis) Get(key string) interface{} {
	reply, err := r.do("GET", key)
	if err != nil {
		return nil
	}

	return reply
}
func (r *Redis) DefaultGet(key string, defaultValue interface{}) interface{} {
	reply, err := r.do("GET", key)
	if err != nil {
		return defaultValue
	}
	return reply
}
func (r *Redis) Delete(key string) error {
	_, err := r.do("DEL", key)
	return err
}
