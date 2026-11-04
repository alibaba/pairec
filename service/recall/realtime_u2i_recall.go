package recall

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/goburrow/cache"
)

type RealTimeU2IRecall struct {
	*BaseRecall
	user2ItemDao module.RealTimeUser2ItemDao
	i2iCacheSign string
}

func NewRealTimeU2IRecall(config recconf.RecallConfig) *RealTimeU2IRecall {
	recall := &RealTimeU2IRecall{
		BaseRecall:   NewBaseRecall(config),
		user2ItemDao: module.NewRealTimeUser2ItemDao(config),
	}
	return recall
}

var i2iCacheAfterFunc = time.AfterFunc

func i2iCacheConfig(conf recconf.RecallConfig, old Recall) (cache.Cache, string) {
	if conf.RecallType != "RealTimeU2IRecall" || conf.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.AdapterType != recconf.DataSource_Type_FeatureStore {
		return nil, ""
	}
	conf.CacheAdapter, conf.CacheConfig, conf.CachePrefix, conf.CacheTime = "", "", "", 0
	encoded, err := json.Marshal(conf)
	if err != nil {
		panic(err)
	}
	sign := utils.Md5(string(encoded))
	if r, ok := old.(*RealTimeU2IRecall); ok && r != nil && r.i2iCacheSign == sign {
		return r.i2iCache(), sign
	}
	return nil, sign
}

func newRealTimeU2IRecallWithCache(conf recconf.RecallConfig, inherited cache.Cache, sign string) *RealTimeU2IRecall {
	if sign == "" {
		return NewRealTimeU2IRecall(conf)
	}
	r := &RealTimeU2IRecall{BaseRecall: NewBaseRecall(conf),
		user2ItemDao: module.NewRealTimeUser2ItemDaoWithCache(conf, inherited), i2iCacheSign: sign}
	if utils.IsNil(r.user2ItemDao) {
		panic(fmt.Sprintf("realtime U2I DAO construction failed, name:%s", conf.Name))
	}
	return r
}

func (r *RealTimeU2IRecall) i2iCache() cache.Cache {
	if dao, ok := r.user2ItemDao.(*module.RealtimeUser2ItemFeatureStoreDao); ok {
		return dao.I2ICache()
	}
	return nil
}

func (r *RealTimeU2IRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()
	if r.cache != nil {
		key := r.cachePrefix + string(user.Id)
		cacheRet := r.cache.Get(key)
		if itemStr, ok := cacheRet.(string); ok {
			cacheRet = []byte(itemStr)
		}
		if itemStr, ok := cacheRet.([]uint8); ok {
			itemIds := strings.Split(string(itemStr), ",")
			for _, id := range itemIds {
				var item *module.Item
				if strings.Contains(id, ":") {
					vars := strings.Split(id, ":")
					item = module.NewItem(vars[0])
					f, _ := strconv.ParseFloat(vars[2], 64)
					// item.AddProperty(vars[1], f)
					item.Score = f
				} else {
					item = module.NewItem(id)
				}
				item.RetrieveId = r.modelName
				ret = append(ret, item)
			}
			log.Info(fmt.Sprintf("requestId=%s\tmodule=RealTimeU2IRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
			return
		}
	}
	ret = r.user2ItemDao.ListItemsByUser(user, context)
	if r.cache != nil && len(ret) > 0 {
		go func() {
			key := r.cachePrefix + string(user.Id)
			var itemIds string
			for _, item := range ret {
				itemIds += fmt.Sprintf("%s::%v", string(item.Id), item.Score) + ","
			}
			itemIds = itemIds[:len(itemIds)-1]
			if err := r.cache.Put(key, itemIds, time.Duration(r.cacheTime)*time.Second); err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=RealTimeU2IRecall\terror=%v",
					context.RecommendId, err))
			}
		}()
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=RealTimeU2IRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return
}
