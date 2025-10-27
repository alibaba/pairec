package recall

import (
	"bytes"
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

type ItemCollaborativeFilterRecall struct {
	enableMultipleItemId bool
	*BaseRecall
	itemCollaborativeDao    module.ItemCollaborativeDao
	multipleItemIdDelimiter string
}

func NewItemCollaborativeFilterRecall(config recconf.RecallConfig) *ItemCollaborativeFilterRecall {
	recall := &ItemCollaborativeFilterRecall{
		BaseRecall:              NewBaseRecall(config),
		itemCollaborativeDao:    module.NewItemCollaborativeDao(config),
		enableMultipleItemId:    false,
		multipleItemIdDelimiter: ",",
	}
	if config.EnableMultipleItemId {
		recall.enableMultipleItemId = true
		if config.MultipleItemIdDelimiter != "" {
			recall.multipleItemIdDelimiter = config.MultipleItemIdDelimiter
		}
	}
	return recall
}

func (r *ItemCollaborativeFilterRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	if r.enableMultipleItemId {
		return r.doGetCandidateItemsByMultipleItemId(user, context)
	}
	start := time.Now()
	item_id := utils.ToString(context.GetParameter("item_id"), "")
	if item_id == "" {
		return
	}
	if r.cache != nil {
		key := item_id
		cacheRet := r.cache.Get(key)
		switch itemStr := cacheRet.(type) {
		case []uint8:
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
				ret = append(ret, item)
			}
		case string:
			itemIds := strings.Split(itemStr, ",")
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
				ret = append(ret, item)
			}
		default:
		}
		if len(ret) > 0 {
			log.Info(fmt.Sprintf("requestId=%s\tmodule=ItemCollaborativeFilterRecall\tfrom=cache\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
			return
		}
	}
	ret = r.itemCollaborativeDao.ListItemsByItem(user, context)
	if r.cache != nil && len(ret) > 0 {
		go func() {
			key := item_id
			var itemIds string
			for _, item := range ret {
				itemIds += fmt.Sprintf("%s:%v", string(item.Id), item.Score) + ","
			}
			itemIds = itemIds[:len(itemIds)-1]
			if err := r.cache.Put(key, itemIds, time.Duration(r.cacheTime)*time.Second); err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=ItemCollaborativeFilterRecall\terror=%v",
					context.RecommendId, err))
			}
		}()
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=ItemCollaborativeFilterRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return
}

func (r *ItemCollaborativeFilterRecall) doGetCandidateItemsByMultipleItemId(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()
	item_id := utils.ToString(context.GetParameter("item_id"), "")
	if item_id == "" {
		return
	}
	itemIds := strings.Split(item_id, r.multipleItemIdDelimiter)
	fetchItems := make([]any, 0, len(itemIds))

	if r.cache != nil {
		for _, item_id := range itemIds {
			key := item_id
			cacheRet := r.cache.Get(key)
			var newItems []*module.Item
			switch itemStr := cacheRet.(type) {
			case []uint8:
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
					newItems = append(newItems, item)
				}
			case string:
				itemIds := strings.Split(itemStr, ",")
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
					newItems = append(newItems, item)
				}
			default:
			}
			if len(newItems) > 0 {
				ret = append(ret, newItems...)
			} else {
				fetchItems = append(fetchItems, item_id)
			}
		}
		if len(ret) > 0 && len(fetchItems) == 0 {
			if len(ret) > r.recallCount {
				ret = ret[:r.recallCount]
			}
			log.Info(fmt.Sprintf("requestId=%s\tmodule=ItemCollaborativeFilterRecall\tfrom=cache\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
			return
		}
	} else {
		for _, itemId := range itemIds {
			fetchItems = append(fetchItems, itemId)
		}
	}
	fetchItemResult := r.itemCollaborativeDao.ListItemsByMultiItemIds(user, context, fetchItems)
	for _, items := range fetchItemResult {
		ret = append(ret, items...)
	}
	if len(ret) > r.recallCount {
		ret = ret[:r.recallCount]
	}
	if r.cache != nil && len(ret) > 0 {
		go func() {
			var buf bytes.Buffer
			for id, items := range fetchItemResult {
				buf.Reset()
				for i, item := range items {
					if i > 0 {
						buf.WriteByte(',')
					}
					buf.WriteString(string(item.Id))
					buf.WriteByte(':')
					buf.Write([]byte(strconv.FormatFloat(item.Score, 'f', -1, 64)))
				}
				if err := r.cache.Put(id, buf.String(), time.Duration(r.cacheTime)*time.Second); err != nil {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=ItemCollaborativeFilterRecall\terror=%v",
						context.RecommendId, err))
				}
			}

		}()
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=ItemCollaborativeFilterRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return
}
