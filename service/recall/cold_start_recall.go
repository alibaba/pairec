package recall

import (
	"fmt"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type ColdStartRecall struct {
	*BaseRecall
	coldStartRecallDao module.ColdStartRecallDao
}

func NewColdStartRecall(config recconf.RecallConfig) *ColdStartRecall {
	recall := &ColdStartRecall{
		BaseRecall:         NewBaseRecall(config),
		coldStartRecallDao: module.NewColdStartRecallDao(config),
	}

	if recall.cache != nil {
		go recall.LoopLoadItems()
	}
	return recall
}

func (r *ColdStartRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()
	if r.cache != nil {
		key := r.modelName
		cacheRet := r.cache.Get(key)
		if itemStr, ok := cacheRet.([]uint8); ok {
			itemIds := strings.Split(string(itemStr), ",")
			for _, id := range itemIds {
				item := module.NewItem(id)
				item.ItemType = r.itemType
				item.RetrieveId = r.modelName
				ret = append(ret, item)
			}
			log.Info(fmt.Sprintf("requestId=%s\tmodule=ColdStartRecall\tname=%s\tfrom=cache\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
			return
		}

		if itemStr, ok := cacheRet.(string); ok {
			itemIds := strings.Split(itemStr, ",")
			for _, id := range itemIds {
				item := module.NewItem(id)
				item.ItemType = r.itemType
				item.RetrieveId = r.modelName
				ret = append(ret, item)
			}
			log.Info(fmt.Sprintf("requestId=%s\tmodule=ColdStartRecall\tname=%s\tfrom=cache\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
			return
		}
	}
	ret = r.coldStartRecallDao.ListItemsByUser(user, context)
	log.Info(fmt.Sprintf("requestId=%s\tmodule=ColdStartRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return
}

func (r *ColdStartRecall) LoopLoadItems() {
	for {
		user := &module.User{}
		recommendContext := &context.RecommendContext{}

		ret := r.coldStartRecallDao.ListItemsByUser(user, recommendContext)
		if len(ret) == 0 {
			log.Error(fmt.Sprintf("module=ColdStartRecall\terror=recall items is null"))
			time.Sleep(3 * time.Second)
			continue
		}
		key := r.modelName
		var itemIds string
		for _, item := range ret {
			itemIds += string(item.Id) + ","
		}
		itemIds = itemIds[:len(itemIds)-1]
		cacheTime := r.cacheTime
		if cacheTime == 0 {
			cacheTime = 1800
		}
		if err := r.cache.Put(key, itemIds, time.Duration(cacheTime)*time.Second); err != nil {
			log.Error(fmt.Sprintf("module=ColdStartRecall\terror=%v", err))
		}

		sleepTime := cacheTime - 1

		if sleepTime <= 0 {
			time.Sleep(time.Duration(500) * time.Millisecond)
		} else {
			time.Sleep(time.Duration(sleepTime) * time.Second)
		}
	}
}
