package module

import (
	"errors"

	"github.com/gomodule/redigo/redis"
	"github.com/alibaba/pairec/v2/persist/redisdb"
	"github.com/alibaba/pairec/v2/recconf"
)

type VectorRedisDao struct {
	redis      *redisdb.Redis
	prefix     string
	defaultKey string
}

func NewVectorRedisDao(config recconf.RecallConfig) *VectorRedisDao {
	dao := &VectorRedisDao{
		prefix:     config.DaoConf.RedisPrefix,
		defaultKey: config.DaoConf.RedisDefaultKey,
	}
	redis, err := redisdb.GetRedis(config.DaoConf.RedisName)
	if err != nil {
		panic(err)
	}
	dao.redis = redis
	return dao
}
func (d *VectorRedisDao) VectorString(id string) (string, error) {
	conn := d.redis.Get()
	defer conn.Close()
	// key := fmt.Sprintf("UI2V_%s", user.Id)
	key := d.prefix + string(id)
	value, err := redis.String(conn.Do("GET", key))
	if err != nil {
		if !errors.Is(err, redis.ErrNil) {
			return "", err
		} else if d.defaultKey != "" {
			value, err = redis.String(conn.Do("GET", d.defaultKey))
			if err != nil && !errors.Is(err, redis.ErrNil) {
				return "", err
			}
		} else {
			return "", VectoryEmptyError
		}
	}

	return value, nil
}
