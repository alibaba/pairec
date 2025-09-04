package fallback

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/persist/cache"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/metrics"
	"github.com/alibaba/pairec/v2/utils"
)

type FeatureStoreFallback struct {
	fsClient        *fs.FSClient
	featureViewName string

	timeout time.Duration

	cache     cache.Cache
	cacheTime int

	completeItemsIfNeed bool
}

var timerPool = sync.Pool{
	New: func() interface{} {
		t := time.NewTimer(time.Hour) // init with any duration
		if !t.Stop() {
			<-t.C
		}
		return t
	},
}

func NewFeatureStoreFallback(conf recconf.FallbackConfig) *FeatureStoreFallback {
	fsclient, err := fs.GetFeatureStoreClient(conf.FeatureStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("event=NewFeatureStoreFallback\terror=%v", err))
		return nil
	}

	cacheConfig := conf.CacheConfig
	if cacheConfig == "" {
		cacheConfig = "{\"defaultExpiration\":1800, \"cleanupInterval\":1800}"
	}

	cacheTime := conf.CacheTime
	if cacheTime <= 0 {
		cacheTime = 7200
	}

	fallbackCache, err := cache.NewCache("localCache", cacheConfig)
	if err != nil {
		log.Error(fmt.Sprintf("event=NewFeatureStoreFallback\terror=%v", err))
		return nil
	}

	return &FeatureStoreFallback{
		fsClient:        fsclient,
		featureViewName: conf.FeatureStoreViewName,

		timeout: time.Duration(conf.Timeout) * time.Millisecond,

		cache:     fallbackCache,
		cacheTime: cacheTime,

		completeItemsIfNeed: conf.CompleteItemsIfNeed,
	}
}

func (r *FeatureStoreFallback) GetTimer() *time.Timer {
	t := timerPool.Get().(*time.Timer)
	t.Reset(r.timeout)
	return t
}
func (r *FeatureStoreFallback) PutTimer(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	timerPool.Put(t)
}

func (r *FeatureStoreFallback) CompleteItemsIfNeed() bool {
	return r.completeItemsIfNeed
}

func (r *FeatureStoreFallback) Recommend(context *context.RecommendContext) []*module.Item {
	start := time.Now()

	contextItemsMap := make(map[module.ItemId]*module.Item)

	if context.GetParameter("item_list") != nil {
		if itemList, ok := context.GetParameter("item_list").([]map[string]any); ok {
			for _, itemData := range itemList {
				itemId := itemData["item_id"]
				itemIdStr := utils.ToString(itemId, "")
				if itemIdStr == "" {
					continue
				}
				item := module.NewItem(itemIdStr)
				item.RetrieveId = "ContextItemRecall"

				for k, v := range itemData {
					if k == "item_id" {
						continue
					} else if k == "score" {
						item.Score = utils.ToFloat(v, 0)
					} else {
						item.AddProperty(k, v)
					}
				}

				if seenItem, ok := contextItemsMap[item.Id]; ok {
					seenItem.Score += item.Score
				} else {
					contextItemsMap[item.Id] = item
				}
			}
		}
	}

	fallbackItemsMap := make(map[module.ItemId]*module.Item)
	moduleName := "Fallback"

	if r.cache != nil {
		key := moduleName
		cacheRet := r.cache.Get(key)
		switch itemsMap := cacheRet.(type) {
		case map[module.ItemId]*module.Item:
			fallbackItemsMap = make(map[module.ItemId]*module.Item, len(itemsMap))
			for _, item := range itemsMap {
				newItem := module.NewItem(string(item.Id))
				newItem.Score = item.Score
				newItem.RetrieveId = item.RetrieveId
				if newItem.RetrieveId == "" {
					newItem.RetrieveId = moduleName
				}
				fallbackItemsMap[newItem.Id] = newItem
			}
		default:
		}
	}

	var useCache bool
	if len(fallbackItemsMap) > 0 {
		useCache = true
	} else {
		itemField := "item_ids"

		featureView := r.fsClient.GetProject().GetFeatureView(r.featureViewName)
		if featureView == nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureStoreFallback\terror=featureView not found, name:%s", context.RecommendId, r.featureViewName))
		} else {
			features, err := featureView.GetOnlineFeatures([]any{"-1"}, []string{"*"}, map[string]string{})
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureStoreFallback\terror=%v", context.RecommendId, err))
			}

			if len(features) != 0 {
				itemIdsStr := utils.ToString(features[0][itemField], "")

				itemIds := make([]string, 0)
				idList := strings.Split(itemIdsStr, ",")
				for _, id := range idList {
					if id != "" {
						itemIds = append(itemIds, id)
					}
				}

				for _, id := range itemIds {
					var item *module.Item

					strs := strings.Split(id, ":")
					if len(strs) == 1 {
						// itemid
						item = module.NewItem(strs[0])
						item.RetrieveId = "Fallback"
					} else if len(strs) == 2 {
						// itemid:RetrieveId
						item = module.NewItem(strs[0])
						if strs[1] != "" {
							item.RetrieveId = strs[1]
						} else {
							item.RetrieveId = "Fallback"
						}
					} else if len(strs) == 3 {
						item = module.NewItem(strs[0])
						if strs[1] != "" {
							item.RetrieveId = strs[1]
						} else {
							item.RetrieveId = "Fallback"
						}
						item.Score = utils.ToFloat(strs[2], float64(0))
					}

					if seenItem, ok := fallbackItemsMap[item.Id]; ok {
						seenItem.Score += item.Score
					} else {
						fallbackItemsMap[item.Id] = item
					}
				}
			}

			if r.cache != nil && len(fallbackItemsMap) > 0 {
				key := moduleName

				itemsMap := make(map[module.ItemId]*module.Item, len(fallbackItemsMap))
				for _, item := range fallbackItemsMap {
					newItem := module.NewItem(string(item.Id))
					newItem.Score = item.Score
					newItem.RetrieveId = item.RetrieveId
					itemsMap[newItem.Id] = newItem
				}

				cacheTime := r.cacheTime
				if cacheTime == 0 {
					cacheTime = 7200
				}
				if err := r.cache.Put(key, itemsMap, time.Duration(cacheTime)*time.Second); err != nil {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureStoreFallback\terror=%v",
						context.RecommendId, err))
				}
			}
		}
	}

	// 优先选择同时出现在上下文召回和兜底数据集的 item，按召回分排序
	// 次选只出现在上下文召回的 item
	// 最后用只出现在兜底数据集的 item 补全
	var firstPriorityItems, secondPriorityItems, remainingItems []*module.Item

	for id, item := range contextItemsMap {
		if fallbackItem, ok := fallbackItemsMap[id]; ok {
			item.Score += fallbackItem.Score
			delete(fallbackItemsMap, id)
			firstPriorityItems = append(firstPriorityItems, item)
		} else {
			secondPriorityItems = append(secondPriorityItems, item)
		}
	}
	for _, item := range fallbackItemsMap {
		remainingItems = append(remainingItems, item)
	}

	sort.Sort(sort.Reverse(module.ItemScoreSlice(firstPriorityItems)))
	rand.Shuffle(len(secondPriorityItems), func(i, j int) {
		secondPriorityItems[i], secondPriorityItems[j] = secondPriorityItems[j], secondPriorityItems[i]
	})
	rand.Shuffle(len(remainingItems), func(i, j int) {
		remainingItems[i], remainingItems[j] = remainingItems[j], remainingItems[i]
	})

	ret := make([]*module.Item, 0, len(firstPriorityItems)+len(secondPriorityItems)+len(remainingItems))
	ret = append(firstPriorityItems, secondPriorityItems...)
	ret = append(ret, remainingItems...)

	log.Info(fmt.Sprintf("requestId=%s\tmodule=fallback\tuseCache=%v\tcost=%d", context.RecommendId, useCache, utils.CostTime(start)))

	if metrics.Enabled() {
		scene, _ := context.Param.GetParameter("scene").(string)
		metrics.FallbackTotal.WithLabelValues(scene).Inc()
	}

	if len(ret) > context.Size {
		return ret[:context.Size]
	}

	return ret
}
