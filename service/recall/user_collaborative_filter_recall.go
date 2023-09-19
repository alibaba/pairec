package recall

import (
	"fmt"
	psort "github.com/alibaba/pairec/v2/sort"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type UserCollaborativeFilterRecall struct {
	*BaseRecall
	userCollaborativeDao module.UserCollaborativeDao
}

func NewUserCollaborativeFilterRecall(config recconf.RecallConfig) *UserCollaborativeFilterRecall {
	recall := &UserCollaborativeFilterRecall{
		BaseRecall:           NewBaseRecall(config),
		userCollaborativeDao: module.NewUserCollaborativeDao(config),
	}
	return recall
}

func (r *UserCollaborativeFilterRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
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
					// item.AddProperty(vars[1], f)
					item.Score = f
				} else {
					item = module.NewItem(id)
				}
				// item := module.NewItem(id)
				// item.AddProperty(r.modelName, 0)
				item.RetrieveId = r.modelName
				item.ItemType = r.itemType
				ret = append(ret, item)
			}
			log.Info(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeFilterRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
			return
		}
	}
	ret = r.userCollaborativeDao.ListItemsByUser(user, context)
	sort.Sort(sort.Reverse(psort.ItemScoreSlice(ret)))
	if r.recallCount < len(ret) {
		ret = ret[:r.recallCount]
	}
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
				log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeFilterRecall\terror=%v",
					context.RecommendId, err))
			}
		}()
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeFilterRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return
}
