package recall

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type RealTimeU2IRecall struct {
	*BaseRecall
	user2ItemDao module.RealTimeUser2ItemDao
}

func NewRealTimeU2IRecall(config recconf.RecallConfig) *RealTimeU2IRecall {
	recall := &RealTimeU2IRecall{
		BaseRecall:   NewBaseRecall(config),
		user2ItemDao: module.NewRealTimeUser2ItemDao(config),
	}
	return recall
}

func (r *RealTimeU2IRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
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
			if err := r.cache.Put(key, itemIds, 1800*time.Second); err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=RealTimeU2IRecall\terror=%v",
					context.RecommendId, err))
			}
		}()
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=RealTimeU2IRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return
}
