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

type UserGroupHotRecall struct {
	*BaseRecall
	userGroupHotRecallDao module.UserGroupHotRecallDao
}

func NewUserGroupHotRecall(config recconf.RecallConfig) *UserGroupHotRecall {
	recall := &UserGroupHotRecall{
		BaseRecall:            NewBaseRecall(config),
		userGroupHotRecallDao: module.NewUserGroupHotRecallDao(config),
	}
	return recall
}

func (r *UserGroupHotRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()
	if r.cache != nil {
		triggerValue := r.userGroupHotRecallDao.TriggerValue(user)
		key := r.cachePrefix + triggerValue
		cacheRet := r.cache.Get(key)
		switch itemStr := cacheRet.(type) {
		case []uint8:
			itemIds := strings.Split(string(itemStr), ",")
			for _, id := range itemIds {
				var item *module.Item
				if strings.Contains(id, ":") {
					vars := strings.Split(id, ":")
					item = module.NewItem(vars[0])
					f, _ := strconv.ParseFloat(vars[2], 64)
					item.AddAlgoScore("hot_score", f)
					item.Score = f
				} else {
					item = module.NewItem(id)
				}
				item.ItemType = r.itemType
				item.RetrieveId = r.modelName

				ret = append(ret, item)
			}
		case string:
			itemIds := strings.Split(itemStr, ",")
			for _, id := range itemIds {
				var item *module.Item
				if strings.Contains(id, ":") {
					vars := strings.Split(id, ":")
					item = module.NewItem(vars[0])
					f, _ := strconv.ParseFloat(vars[2], 64)
					item.AddAlgoScore("hot_score", f)
					item.Score = f
				} else {
					item = module.NewItem(id)
				}
				item.ItemType = r.itemType
				item.RetrieveId = r.modelName

				ret = append(ret, item)
			}
		default:
		}

		if len(ret) > 0 {
			log.Info(fmt.Sprintf("requestId=%s\tmodule=UserGroupHotRecall\tfrom=cache\tcount=%d\tcost=%d", context.RecommendId, len(ret), utils.CostTime(start)))
			return
		}
	}
	ret = r.userGroupHotRecallDao.ListItemsByUser(user, context)
	if r.cache != nil && len(ret) > 0 {
		go func() {
			triggerValue := r.userGroupHotRecallDao.TriggerValue(user)
			key := r.cachePrefix + triggerValue
			var itemIds string
			for _, item := range ret {
				itemIds += fmt.Sprintf("%s::%v", string(item.Id), item.Score) + ","
			}
			itemIds = itemIds[:len(itemIds)-1]

			cacheTime := r.cacheTime
			if cacheTime == 0 {
				cacheTime = 1800
			}

			if err := r.cache.Put(key, itemIds, time.Duration(cacheTime)*time.Second); err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=UserGroupHotRecall\terror=%v",
					context.RecommendId, err))
			}
		}()
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=UserGroupHotRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return
}
