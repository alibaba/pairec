package module

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/redisdb"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type UserCustomRecallRedisDao struct {
	redis       *redisdb.Redis
	itemType    string
	recallName  string
	prefix      string
	recallCount int
}

func NewUserCusteomRecallRedisDao(config recconf.RecallConfig) *UserCustomRecallRedisDao {
	redis, err := redisdb.GetRedis(config.DaoConf.RedisName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}

	dao := &UserCustomRecallRedisDao{
		recallCount: config.RecallCount,
		redis:       redis,
		prefix:      config.DaoConf.RedisPrefix,
		itemType:    config.ItemType,
		recallName:  config.Name,
	}
	return dao
}

func (d *UserCustomRecallRedisDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	conn := d.redis.Get()
	defer conn.Close()
	uid := string(user.Id)
	key := d.prefix + uid
	value, err := redis.String(conn.Do("GET", key))
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCustomRecallRedisDao\tuid=%s\terror=%v", context.RecommendId, uid, err))
		return
	}
	itemIds := make([]string, 0, d.recallCount)
	if value != "" {
		idList := strings.Split(value, ",")
		for _, id := range idList {
			if len(id) > 0 {
				itemIds = append(itemIds, id)
			}
		}
	}

	if len(itemIds) == 0 {
		return
	}

	if len(itemIds) > d.recallCount {
		rand.Shuffle(len(itemIds)/2, func(i, j int) {
			itemIds[i], itemIds[j] = itemIds[j], itemIds[i]
		})

		itemIds = itemIds[:d.recallCount]
	}

	for _, id := range itemIds {
		strs := strings.Split(id, ":")
		if len(strs) == 1 {
			// itemid
			item := NewItem(strs[0])
			item.ItemType = d.itemType
			item.RetrieveId = d.recallName
			ret = append(ret, item)
		} else if len(strs) == 2 {
			// itemid:RetrieveId
			item := NewItem(strs[0])
			item.ItemType = d.itemType
			if strs[1] != "" {
				item.RetrieveId = strs[1]
			} else {
				item.RetrieveId = d.recallName
			}
			ret = append(ret, item)
		} else if len(strs) == 3 {
			item := NewItem(strs[0])
			item.ItemType = d.itemType
			if strs[1] != "" {
				item.RetrieveId = strs[1]
			} else {
				item.RetrieveId = d.recallName
			}
			item.Score = utils.ToFloat(strs[2], float64(0))
			ret = append(ret, item)
		}
	}

	return
}
