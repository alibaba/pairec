package dao

import (
	"context"
	"fmt"
	"strings"

	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/api"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/constants"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/datasource/redisdb"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/utils"
	"github.com/go-redis/redis/v8"
)

type FeatureViewRedisDao struct {
	redisClinet     *redis.Client
	redisPrefix     string
	redisDelimiter  string
	primaryKeyField string
	eventTimeField  string
	ttl             int
	fields          []string
	fieldTypeMap    map[string]constants.FSType
	fieldIndexMap   map[string]int
}

func NewFeatureViewRedisDao(config DaoConfig) *FeatureViewRedisDao {
	dao := FeatureViewRedisDao{
		redisPrefix:     config.RedisPrefix,
		redisDelimiter:  "\u0001",
		primaryKeyField: config.PrimaryKeyField,
		eventTimeField:  config.EventTimeField,
		ttl:             config.TTL,
		fields:          config.Fields,
		fieldTypeMap:    config.FieldTypeMap,
		fieldIndexMap:   make(map[string]int, len(config.Fields)),
	}
	client, err := redisdb.GetRedis(config.RedisName)
	if err != nil {
		return nil
	}

	dao.redisClinet = client.GetClient()
	for i, field := range dao.fields {
		dao.fieldIndexMap[field] = i

	}
	return &dao
}
func (d *FeatureViewRedisDao) GetFeatures(keys []interface{}, selectFields []string) ([]map[string]interface{}, error) {
	var pkeys []string
	for _, key := range keys {
		pkey := utils.ToString(key, "")
		pkeys = append(pkeys, fmt.Sprintf("%s%s", d.redisPrefix, pkey))
	}

	ctx := context.Background()
	values, err := d.redisClinet.MGet(ctx, pkeys...).Result()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(keys))
	for i, value := range values {
		if value == nil {
			result = append(result, map[string]interface{}{d.primaryKeyField: keys[i]})
		} else {
			if str, ok := value.(string); ok {
				m := make(map[string]interface{}, len(selectFields)+1)
				m[d.primaryKeyField] = keys[i]
				strs := strings.Split(str, d.redisDelimiter)
				for _, field := range selectFields {
					if index, ok := d.fieldIndexMap[field]; ok {
						switch d.fieldTypeMap[field] {
						case constants.FS_DOUBLE, constants.FS_FLOAT:
							m[field] = utils.ToFloat(strs[index], 0)
						case constants.FS_INT32, constants.FS_INT64:
							m[field] = utils.ToInt(strs[index], 0)
						default:
							m[field] = strs[index]
						}

					}
				}

				result = append(result, m)
			}

		}

	}

	return result, nil
}

func (d *FeatureViewRedisDao) GetUserSequenceFeature(keys []interface{}, userIdField string, sequenceConfig api.FeatureViewSeqConfig, onlineConfig []*api.SeqConfig) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}
