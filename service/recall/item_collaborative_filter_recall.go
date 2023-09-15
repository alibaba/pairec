package recall

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/utils"
)

type ItemCollaborativeFilterRecall struct {
	*BaseRecall
	itemCollaborativeDao module.ItemCollaborativeDao
}

func NewItemCollaborativeFilterRecall(config recconf.RecallConfig) *ItemCollaborativeFilterRecall {
	recall := &ItemCollaborativeFilterRecall{
		BaseRecall:           NewBaseRecall(config),
		itemCollaborativeDao: module.NewItemCollaborativeDao(config),
	}
	return recall
}

func (r *ItemCollaborativeFilterRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()
	if r.cache != nil {
		key := r.cachePrefix + string(user.Id)
		cacheRet := r.cache.Get(key)
		if itemStr, ok := cacheRet.([]uint8); ok {
			itemIds := strings.Split(string(itemStr), ",")
			for _, id := range itemIds {
				var item *module.Item
				if strings.Contains(id, ":") {
					vars := strings.Split(id, ":")
					item = module.NewItem(vars[0])
					f, _ := strconv.ParseFloat(vars[1], 64)
					item.Score = f
				} else {
					item = module.NewItem(id)
				}
				item.RetrieveId = r.modelName
				item.ItemType = r.itemType
				ret = append(ret, item)
			}
			log.Info(fmt.Sprintf("requestId=%s\tmodule=ItemCollaborativeFilterRecall\tcount=%d\tcost=%d", context.RecommendId, len(ret), utils.CostTime(start)))
			return
		}
	}
	ret = r.itemCollaborativeDao.ListItemsByItem(user, context)
	if r.cache != nil && len(ret) > 0 {
		go func() {
			key := r.cachePrefix + string(user.Id)
			var itemIds string
			for _, item := range ret {
				//item.AddProperty(r.modelName, 0)
				itemIds += fmt.Sprintf("%s:%v", string(item.Id), item.Score) + ","
			}
			itemIds = itemIds[:len(itemIds)-1]
			if err := r.cache.Put(key, itemIds, 1800*time.Second); err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=ItemCollaborativeFilterRecall\terror=%v",
					context.RecommendId, err))
			}
		}()
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=ItemCollaborativeFilterRecall\tcount=%d\tcost=%d", context.RecommendId, len(ret), utils.CostTime(start)))
	return
}
