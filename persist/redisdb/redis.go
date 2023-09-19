package redisdb

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils/netutil"
)

type Redis struct {
	Pool           *redis.Pool
	Host           string
	Port           int
	Password       string
	MaxIdle        int
	DbNum          int
	connectTimeout time.Duration
	readTimeout    time.Duration
	writeTimeout   time.Duration
}

var redisPools = make(map[string]*Redis)
var redisConfs = make(map[string]*recconf.RedisConfig)

func GetRedisConf(name string) (*recconf.RedisConfig, error) {
	if conf, ok := redisConfs[name]; !ok {
		return nil, fmt.Errorf("RedisConf:not found, name:%s", name)
	} else {
		return conf, nil
	}
}

func GetRedis(name string) (*Redis, error) {
	if _, ok := redisPools[name]; !ok {
		return nil, fmt.Errorf("Redis:not found, name:%s", name)
	}

	return redisPools[name], nil
}

func (r *Redis) Init() error {

	dialFunc := func() (redis.Conn, error) {
		var (
			ip  string
			err error
		)

		ip, err = netutil.GetAddrByHost(r.Host)
		if err != nil {
			ip = r.Host
		}
		addr := fmt.Sprintf("%s:%d", ip, r.Port)
		conn, err := redis.DialTimeout("tcp", addr, r.connectTimeout, r.readTimeout, r.writeTimeout)
		if err != nil {
			// use Host:Port try again
			addr = fmt.Sprintf("%s:%d", r.Host, r.Port)
			conn, err = redis.DialTimeout("tcp", addr, r.connectTimeout, r.readTimeout, r.writeTimeout)
			if err != nil {
				return nil, err
			}
		}

		if len(r.Password) != 0 {
			_, err = conn.Do("AUTH", r.Password)
			if err != nil {
				conn.Close()
				return nil, err
			}
		}
		_, selecterr := conn.Do("SELECT", r.DbNum)
		if selecterr != nil {
			conn.Close()
			return nil, selecterr
		}
		return conn, nil
	}

	r.Pool = &redis.Pool{
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

	c := r.Pool.Get()
	defer c.Close()

	return c.Err()
}
func (r *Redis) Get() redis.Conn {
	return r.Pool.Get()
}
func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.RedisConfs {
		redisConfs[name] = &conf
		if _, ok := redisPools[name]; ok {
			continue
		}

		r := &Redis{
			Host:     conf.Host,
			Password: conf.Password,
			Port:     conf.Port,
			MaxIdle:  conf.MaxIdle,
			DbNum:    conf.DbNum,
		}

		r.connectTimeout = time.Millisecond * time.Duration(50)
		if conf.ConnectTimeout != 0 {
			r.connectTimeout = time.Millisecond * time.Duration(conf.ConnectTimeout)
		}

		r.readTimeout = time.Millisecond * time.Duration(100)
		if conf.ReadTimeout != 0 {
			r.readTimeout = time.Millisecond * time.Duration(conf.ReadTimeout)
		}

		r.writeTimeout = time.Millisecond * time.Duration(100)
		if conf.WriteTimeout != 0 {
			r.writeTimeout = time.Millisecond * time.Duration(conf.WriteTimeout)
		}

		err := r.Init()
		if err != nil {
			panic(fmt.Sprintf("name=%s, err=%v", name, err))
		}
		redisPools[name] = r
	}
}
