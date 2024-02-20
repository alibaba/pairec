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
		key := r.cachePrefix + string(user.Id)
		cacheRet := r.cache.Get(key)
		switch itemStr := cacheRet.(type) {
		case []uint8:
			itemIds := strings.Split(string(itemStr), ",")
			for _, id := range itemIds {
				item := &module.Item{
					Id:         module.ItemId(id),
					ItemType:   r.itemType,
					RetrieveId: r.modelName,
				}
				ret = append(ret, item)
			}
		case string:
			itemIds := strings.Split(itemStr, ",")
			for _, id := range itemIds {
				item := &module.Item{
					Id:         module.ItemId(id),
					ItemType:   r.itemType,
					RetrieveId: r.modelName,
				}
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
			key := r.cachePrefix + string(user.Id)
			var itemIds string
			for _, item := range ret {
				itemIds += string(item.Id) + ","
			}
			itemIds = itemIds[:len(itemIds)-1]
			if err := r.cache.Put(key, itemIds, 1800*time.Second); err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=UserGroupHotRecall\terror=%v",
					context.RecommendId, err))
			}
		}()
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=UserGroupHotRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return
}
