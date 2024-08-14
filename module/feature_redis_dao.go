package module

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/gomodule/redigo/redis"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/redisdb"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

const (
	REDIS_DATA_TYPE_STRING = "string"
	REDIS_DATA_TYPE_HASH   = "hash"

	REDIS_FIELD_TYPE_CSV  = "csv"
	REDIS_FIELD_TYPE_JSON = "json"
)

type FeatureRedisDao struct {
	*FeatureBaseDao
	redis            *redisdb.Redis
	redisPrefix      string
	redisDelimeter   string
	redisDataType    string
	redisFieldType   string
	userSelectFields []interface{}
	itemSelectFields []interface{}
}

func NewFeatureRedisDao(config recconf.FeatureDaoConfig) *FeatureRedisDao {
	dao := &FeatureRedisDao{
		FeatureBaseDao: NewFeatureBaseDao(&config),
		redisPrefix:    config.RedisPrefix,
		redisDelimeter: config.RedisValueDelimeter,
		redisDataType:  REDIS_DATA_TYPE_STRING,
		redisFieldType: REDIS_FIELD_TYPE_CSV,
	}
	redis, err := redisdb.GetRedis(config.RedisName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}
	dao.redis = redis
	if dao.redisDelimeter == "" {
		dao.redisDelimeter = ","
	}
	if config.RedisDataType != "" {
		dao.redisDataType = config.RedisDataType
	}
	if config.RedisFieldType != "" {
		dao.redisFieldType = config.RedisFieldType
	}

	if config.UserSelectFields != "" && config.UserSelectFields != "*" {
		fields := strings.Split(config.UserSelectFields, ",")
		for _, f := range fields {
			dao.userSelectFields = append(dao.userSelectFields, f)
		}
	}
	if config.ItemSelectFields != "" && config.ItemSelectFields != "*" {
		fields := strings.Split(config.ItemSelectFields, ",")
		for _, f := range fields {
			dao.itemSelectFields = append(dao.itemSelectFields, f)
		}
	}
	return dao
}

func (d *FeatureRedisDao) FeatureFetch(user *User, items []*Item, context *context.RecommendContext) {
	if d.featureStore == Feature_Store_User {
		d.userFeatureFetch(user, context)
	} else {
		d.itemsFeatureFetch(items, context)
	}
}
func (d *FeatureRedisDao) userFeatureFetch(user *User, context *context.RecommendContext) {
	comms := strings.Split(d.featureKey, ":")
	if len(comms) < 2 {
		log.Error(fmt.Sprintf("requestId=%s\tuid=%s\terror=featureKey error(%s)", context.RecommendId, user.Id, d.featureKey))
		return
	}

	key := user.StringProperty(comms[1])
	if key == "" {
		log.Error(fmt.Sprintf("requestId=%s\tuid=%s\terror=property not found(%s)", context.RecommendId, user.Id, comms[1]))
		return
	}

	key = d.redisPrefix + key

	conn := d.redis.Get()
	defer conn.Close()
	if d.redisDataType == REDIS_DATA_TYPE_STRING {
		d.userFeatureFetchByString(user, context, conn, key)
	} else if d.redisDataType == REDIS_DATA_TYPE_HASH {
		d.userFeatureFetchByHash(user, context, conn, key)
	}
}
func (d *FeatureRedisDao) userFeatureFetchByString(user *User, context *context.RecommendContext, conn redis.Conn, key string) error {
	str, err := redis.String(conn.Do("GET", key))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tmsg=user feature empty", context.RecommendId, user.Id))
		} else {
			log.Error(fmt.Sprintf("requestId=%s\tuid=%s\terror=get user feature error(%v)", context.RecommendId, user.Id, err))
		}
		return err
	}
	properties := make(map[string]interface{})
	if d.redisFieldType == REDIS_FIELD_TYPE_CSV {
		keyParis := strings.Split(str, d.redisDelimeter)
		if len(keyParis) == 0 {
			return nil
		}
		for _, pair := range keyParis {
			idx := strings.Index(pair, ":")
			if idx > 0 {
				name := pair[:idx]
				value := pair[idx+1:]
				properties[name] = value
			}
		}
	} else if d.redisFieldType == REDIS_FIELD_TYPE_JSON {
		err := json.Unmarshal([]byte(str), &properties)
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tuid=%s\terror=get user feature error(%v)", context.RecommendId, user.Id, err))
			return err
		}
	}
	if d.cacheFeaturesName != "" {
		user.AddCacheFeatures(d.cacheFeaturesName, properties)
	} else {
		user.AddProperties(properties)
	}

	return nil
}
func (d *FeatureRedisDao) userFeatureFetchByHash(user *User, context *context.RecommendContext, conn redis.Conn, key string) error {
	properties := make(map[string]interface{})
	if len(d.userSelectFields) == 0 {
		// get all fields
		strs, err := redis.Strings(conn.Do("HGETALL", key))
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tuid=%s\terror=get user feature error(%v)", context.RecommendId, user.Id, err))
			return err
		}
		for i := 0; i < len(strs); i += 2 {
			properties[strs[i]] = strs[i+1]
		}

	} else {
		var params []interface{}
		params = append(params, key)
		params = append(params, d.userSelectFields...)
		strs, err := redis.Strings(conn.Do("HMGET", params...))
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tuid=%s\terror=get user feature error(%v)", context.RecommendId, user.Id, err))
			return err
		}
		for i, val := range strs {
			if val == "" {
				continue
			}
			properties[d.userSelectFields[i].(string)] = val
		}

	}

	if d.cacheFeaturesName != "" {
		user.AddCacheFeatures(d.cacheFeaturesName, properties)
	} else {
		user.AddProperties(properties)
	}
	return nil
}

func (d *FeatureRedisDao) itemsFeatureFetch(items []*Item, context *context.RecommendContext) {
	fk := d.featureKey
	if fk != "item:id" {
		comms := strings.Split(d.featureKey, ":")
		if len(comms) < 2 {
			log.Error(fmt.Sprintf("requestId=%s\tevent=itemsFeatureFetch\terror=featureKey error(%s)", context.RecommendId, d.featureKey))
			return
		}

		fk = comms[1]
	}

	cpuCount := utils.MaxInt(int(len(items)/100), 1)
	maps := make(map[int][]*Item)
	for i, item := range items {
		maps[i%cpuCount] = append(maps[i%cpuCount], item)
	}

	requestCh := make(chan []*Item, cpuCount)
	defer close(requestCh)

	for _, itemlist := range maps {
		requestCh <- itemlist
	}

	var wg sync.WaitGroup
	for i := 0; i < cpuCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case itemlist := <-requestCh:
				var keys []interface{}
				for _, item := range itemlist {
					var key string
					if fk == "item:id" {
						key = string(item.Id)
					} else {
						key = item.StringProperty(fk)
					}
					key = d.redisPrefix + key

					keys = append(keys, key)
				}

				conn := d.redis.Get()
				defer conn.Close()
				if d.redisDataType == REDIS_DATA_TYPE_STRING {
					d.itemFeatureFetchByString(itemlist, context, conn, keys)
				} else if d.redisDataType == REDIS_DATA_TYPE_HASH {
					d.itemFeatureFetchByHash(itemlist, context, conn, keys)
				}
			default:
			}
		}()
	}
	wg.Wait()
}

func (d *FeatureRedisDao) itemFeatureFetchByString(items []*Item, context *context.RecommendContext, conn redis.Conn, keys []interface{}) error {
	values, err := redis.Strings(conn.Do("MGET", keys...))
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureRedisDao\terror=%v", context.RecommendId, err))
		return err
	}

	for i, str := range values {
		if str == "" {
			continue
		}
		item := items[i]
		properties := make(map[string]interface{})

		if d.redisFieldType == REDIS_FIELD_TYPE_CSV {
			keyParis := strings.Split(str, d.redisDelimeter)
			if len(keyParis) == 0 {
				continue
			}

			for _, pair := range keyParis {
				keyValues := strings.Split(pair, ":")
				if len(keyValues) == 2 {
					name := keyValues[0]
					val := keyValues[1]
					f, err := strconv.ParseFloat(val, 64)
					if err == nil {
						properties[name] = f
					} else {
						properties[name] = val
					}
				}
			}

		} else if d.redisFieldType == REDIS_FIELD_TYPE_JSON {
			if err := json.Unmarshal([]byte(str), &properties); err != nil {
				continue
			}
		}
		item.AddProperties(properties)
	}

	return nil
}

func (d *FeatureRedisDao) itemFeatureFetchByHash(items []*Item, context *context.RecommendContext, conn redis.Conn, keys []interface{}) error {
	if len(d.itemSelectFields) == 0 {
		// get all fields
		for _, key := range keys {
			conn.Send("HGETALL", key)
		}
		conn.Flush()

		for i := range keys {
			strs, err := redis.Strings(conn.Receive())
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureRedisDao\terror=%v", context.RecommendId, err))
				continue
			}
			item := items[i]
			properties := make(map[string]interface{})
			for i := 0; i < len(strs); i += 2 {
				properties[strs[i]] = strs[i+1]
			}
			item.AddProperties(properties)

		}
	} else {
		for _, key := range keys {
			var params []interface{}
			params = append(params, key)
			params = append(params, d.userSelectFields...)
			conn.Send("HMGET", params...)
		}
		conn.Flush()
		for i := range keys {
			strs, err := redis.Strings(conn.Receive())
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureRedisDao\terror=%v", context.RecommendId, err))
				continue
			}
			item := items[i]
			properties := make(map[string]interface{})
			for k, val := range strs {
				if val == "" {
					continue
				}
				properties[d.userSelectFields[k].(string)] = val
			}
			item.AddProperties(properties)

		}

	}

	return nil
}
