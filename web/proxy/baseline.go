package proxy

import (
	"errors"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"github.com/alibaba/pairec/persist/redisdb"
)

func GetBaselineData(uid string, redisdb *redisdb.Redis, prefix string, size, retry int) (ret []string) {
	if retry <= 0 {
		return
	}

	conn := redisdb.Get()
	defer conn.Close()

	uidKey := "base_" + uid
	value, err := redis.Int(conn.Do("GET", uidKey))

	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			value = 0
		} else {
			return
		}
	}

	var keys []interface{}
	for i := value; i < value+size; i++ {
		keys = append(keys, prefix+strconv.Itoa(i))
	}
	strs, err := redis.Strings(conn.Do("MGET", keys...))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			conn.Do("SET", uidKey, 0, "EX", 21600)
			retry--
			return GetBaselineData(uid, redisdb, prefix, size, retry)
		} else {
			return
		}
	}

	for _, s := range strs {
		if s != "" {
			ret = append(ret, s)
		}
	}

	if len(ret) < size {
		conn.Do("SET", uidKey, 0, "EX", 21600)
		retry--
		return GetBaselineData(uid, redisdb, prefix, size, retry)
	}

	conn.Do("SET", uidKey, value+size, "EX", 21600)
	return
}
