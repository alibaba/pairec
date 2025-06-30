package module

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/redisdb"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/gomodule/redigo/redis"
)

type User2ItemExposureRedisDao struct {
	redis                 *redisdb.Redis
	prefix                string
	maxItems              int
	timeInterval          int //  second
	writeLogExcludeScenes map[string]bool
	clearLogScene         string
}

func NewUser2ItemExposureRedisDao(config recconf.FilterConfig) *User2ItemExposureRedisDao {
	dao := &User2ItemExposureRedisDao{
		maxItems:              100,
		timeInterval:          -1,
		writeLogExcludeScenes: make(map[string]bool),
		clearLogScene:         config.ClearLogIfNotEnoughScene,
	}
	redis, err := redisdb.GetRedis(config.DaoConf.RedisName)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return nil
	}

	dao.redis = redis
	if config.MaxItems > 0 {
		dao.maxItems = config.MaxItems
	}

	if config.TimeInterval > 0 {
		dao.timeInterval = config.TimeInterval
	}

	for _, scene := range config.WriteLogExcludeScenes {
		dao.writeLogExcludeScenes[scene] = true
	}
	dao.prefix = config.DaoConf.RedisPrefix
	return dao
}

type exposureItemRedis struct {
	ItemIds   []string `json:"item_ids"`
	Timestamp int64    `json:"timestamp"`
}

func (d *User2ItemExposureRedisDao) LogHistory(user *User, items []*Item, context *context.RecommendContext) {
	scene := context.GetParameter("scene").(string)
	if _, exist := d.writeLogExcludeScenes[scene]; exist {
		return
	}

	if len(items) == 0 {
		log.Warning(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureRedisDao\terr=items empty", context.RecommendId))
		return
	}

	prefix := d.prefix
	/**
	if prefix == "" {
		scene := context.GetParameter("scene").(string)
		prefix = scene + "_"
	}
	**/
	addTime := time.Now().Unix()
	uid := string(user.Id)
	key := prefix + uid

	exposureItem := exposureItemRedis{
		Timestamp: addTime,
	}

	for i := 0; i < len(items); i++ {
		exposureItem.ItemIds = append(exposureItem.ItemIds, string(items[i].Id))
	}

	data, _ := json.Marshal(exposureItem)
	conn := d.redis.Get()
	defer conn.Close()

	len, err := redis.Int(conn.Do("LPUSH", key, string(data)))
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureRedisDao\tuid=%s\terr=%v", context.RecommendId, uid, err))
		return
	}

	if len > d.maxItems {
		conn.Do("LTRIM", key, 0, d.maxItems-1)
	}

	if d.timeInterval > 0 {
		conn.Do("EXPIRE", key, d.timeInterval)
	}

	log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tmsg=log history success", context.RecommendId, user.Id))

}
func (d *User2ItemExposureRedisDao) FilterByHistory(uid UID, items []*Item, context *context.RecommendContext) (ret []*Item) {
	prefix := d.prefix
	key := prefix + string(uid)
	conn := d.redis.Get()
	defer conn.Close()

	bytes, err := redis.ByteSlices(conn.Do("LRANGE", key, 0, d.maxItems-1))
	if err != nil {
		log.Error(fmt.Sprintf("module=User2ItemExposureRedisDao\tuid=%s\terr=%v", uid, err))
		ret = items
		return
	}

	fiterIds := make(map[string]bool)
	t := time.Now().Unix()
	for _, byte := range bytes {
		exposureItem := exposureItemRedis{}
		if err := json.Unmarshal(byte, &exposureItem); err == nil {
			if d.timeInterval > 0 && (exposureItem.Timestamp < t-int64(d.timeInterval)) {
				continue
			}

			for _, id := range exposureItem.ItemIds {
				fiterIds[id] = true
			}

		}
	}

	for _, item := range items {
		if _, ok := fiterIds[string(item.Id)]; !ok {
			ret = append(ret, item)
		}
	}
	return
}

func (d *User2ItemExposureRedisDao) ClearHistory(user *User, context *context.RecommendContext) {
	scene := context.GetParameter("scene").(string)
	if scene != d.clearLogScene {
		return
	}
	prefix := d.prefix
	key := prefix + string(user.Id)
	conn := d.redis.Get()
	defer conn.Close()

	err := conn.Send("DEL", key)
	if err != nil {
		context.LogError(fmt.Sprintf("delete user [%s] exposure items from redis failed with err:%v", user.Id, err))
	}
}

func (d *User2ItemExposureRedisDao) GetExposureItemIds(user *User, context *context.RecommendContext) (ret string) {
	uid := string(user.Id)
	prefix := d.prefix
	key := prefix + string(uid)
	conn := d.redis.Get()
	defer conn.Close()

	bytes, err := redis.ByteSlices(conn.Do("LRANGE", key, 0, d.maxItems-1))
	if err != nil {
		log.Error(fmt.Sprintf("module=User2ItemExposureRedisDao\tuid=%s\terr=%v", uid, err))
		return
	}

	fiterIds := make([]string, 0, 10)
	t := time.Now().Unix()
	for _, byte := range bytes {
		exposureItem := exposureItemRedis{}
		if err := json.Unmarshal(byte, &exposureItem); err == nil {
			if d.timeInterval > 0 && (exposureItem.Timestamp < t-int64(d.timeInterval)) {
				continue
			}

			for _, id := range exposureItem.ItemIds {
				fiterIds = append(fiterIds, id)
			}

		}
	}

	ret = strings.Join(fiterIds, ",")
	return
}
