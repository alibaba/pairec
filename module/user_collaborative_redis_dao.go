package module

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/redisdb"
	"github.com/alibaba/pairec/v2/recconf"
	"math/rand"
	"strconv"
	"strings"
)

type UserCollaborativeRedisDao struct {
	redis         *redisdb.Redis
	itemType      string
	recallName    string
	prefix        string
	recallCount   int
	normalization bool
}

func NewUserCollaborativeRedisDao(config recconf.RecallConfig) *UserCollaborativeRedisDao {
	redisIns, err := redisdb.GetRedis(config.DaoConf.RedisName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}

	dao := &UserCollaborativeRedisDao{
		recallCount: config.RecallCount,
		redis:       redisIns,
		prefix:      config.DaoConf.RedisPrefix,
		itemType:    config.ItemType,
		recallName:  config.Name,
	}
	if config.UserCollaborativeDaoConf.Normalization == "on" || config.UserCollaborativeDaoConf.Normalization == "" {
		dao.normalization = true
	}
	return dao
}

func (d *UserCollaborativeRedisDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	conn := d.redis.Get()
	defer conn.Close()
	uid := string(user.Id)
	key := d.prefix + uid
	value, err := redis.String(conn.Do("GET", key))
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeRedisDao\tuid=%s\terror=%v", context.RecommendId, uid, err))
		return
	}
	itemIds := make([]string, 0)
	preferScoreMap := make(map[string]float64)
	if value != "" {
		idList := strings.Split(value, ",")
		for _, id := range idList {
			strs := strings.Split(id, ":")
			if strs[0] == "" {
				continue
			}
			itemIds = append(itemIds, strs[0])
			preferScoreMap[strs[0]] = 1
			if len(strs) > 1 {
				if score, err := strconv.ParseFloat(strs[1], 64); err == nil {
					preferScoreMap[strs[0]] = score
				} else {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeRedisDao\tevent=ParsePreferScore\tuid=%s\terr=%v", context.RecommendId, uid, err))
				}
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

	cpuCount := 4
	maps := make(map[int][]interface{})
	for i, id := range itemIds {
		maps[i%cpuCount] = append(maps[i%cpuCount], id)
	}

	itemIdCh := make(chan []interface{}, cpuCount)
	for _, ids := range maps {
		itemIdCh <- ids
	}

	itemCh := make(chan []*Item, cpuCount)

	for i := 0; i < cpuCount; i++ {
		go func() {
			result := make([]*Item, 0)
		LOOP:
			for {
				select {
				case ids := <-itemIdCh:
					res, err := conn.Do("MGET", ids...)
					if err != nil {
						log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeRedisDao\tuid=%s\terror=%v", context.RecommendId, uid, err))
						goto LOOP
					}

					similarIds := res.([]interface{})

					for _, id := range similarIds {
						tmp := id.([]uint8)
						list := strings.Split(string(tmp), ",")

						for _, str := range list {
							strs := strings.Split(str, ":")
							preferScore := preferScoreMap[strs[0]]
							if len(strs) == 2 && len(strs[0]) > 0 && strs[0] != "null" {
								item := NewItem(strs[0])
								item.RetrieveId = d.recallName
								item.ItemType = d.itemType
								if tmpScore, err := strconv.ParseFloat(strings.TrimSpace(strs[1]), 64); err == nil {
									item.Score = tmpScore * preferScore
								} else {
									item.Score = preferScore
								}
								result = append(result, item)
							}
						}
					}
				default:
					goto DONE
				}
			}
		DONE:
			itemCh <- result
		}()
	}
	ret = mergeUserCollaborativeItemsResult(itemCh, cpuCount, d.normalization)

	close(itemCh)
	close(itemIdCh)
	return
}

func (d *UserCollaborativeRedisDao) GetTriggers(user *User, context *context.RecommendContext) (itemTriggers map[string]float64) {
	itemTriggers = make(map[string]float64)
	triggerInfos := d.GetTriggerInfos(user, context)

	for _, trigger := range triggerInfos {
		itemTriggers[trigger.ItemId] = trigger.Weight
	}
	return
}

func (d *UserCollaborativeRedisDao) GetTriggerInfos(user *User, context *context.RecommendContext) (triggerInfos []*TriggerInfo) {
	conn := d.redis.Get()
	defer conn.Close()
	uid := string(user.Id)
	key := d.prefix + uid
	value, err := redis.String(conn.Do("GET", key))
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeRedisDao\tuid=%s\terror=%v", context.RecommendId, uid, err))
		return
	}
	if value != "" {
		idList := strings.Split(value, ",")
		for _, id := range idList {
			strs := strings.Split(id, ":")
			if strs[0] == "" {
				continue
			}
			trigger := &TriggerInfo{
				ItemId: strs[0],
				Weight: 1,
			}
			if len(strs) > 1 {
				if score, err := strconv.ParseFloat(strs[1], 64); err == nil {
					//itemTriggers[strs[0]] = score
					trigger.Weight = score
				} else {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeRedisDao\tevent=ParsePreferScore\tuid=%s\terr=%v", context.RecommendId, uid, err))
				}
			}
			triggerInfos = append(triggerInfos, trigger)
		}
	}
	return
}
