package module

import (
	"fmt"
	"strings"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/goburrow/cache"
)

// Item2XFeatureStoreDao gets item property(x) values from a featurestore item attribute feature view.
type Item2XFeatureStoreDao struct {
	fsClient    *fs.FSClient
	item2XTable string
	xKey        string
	cache       cache.Cache
}

func NewItem2XFeatureStoreDao(config recconf.Item2XConfig) *Item2XFeatureStoreDao {
	dao := &Item2XFeatureStoreDao{
		item2XTable: config.Item2XTable,
		xKey:        config.XKey,
		cache:       newItem2XCache(config),
	}

	fsclient, err := fs.GetFeatureStoreClient(config.FeatureStoreName)
	if err != nil {
		panic(fmt.Sprintf("get featurestore client error, name:%s, error:%v", config.FeatureStoreName, err))
	}
	dao.fsClient = fsclient

	return dao
}

func (d *Item2XFeatureStoreDao) ListItem2X(itemIds []string, context *context.RecommendContext) map[string]string {
	item2XMap := make(map[string]string, len(itemIds))
	if len(itemIds) == 0 {
		return item2XMap
	}

	missedIds := fetchItem2XFromCache(d.cache, itemIds, item2XMap)
	if len(missedIds) == 0 {
		return item2XMap
	}

	featureView := d.fsClient.GetProject().GetFeatureView(d.item2XTable)
	if featureView == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=Item2XFeatureStoreDao\terror=featureView not found, featureview:%s", context.RecommendId, d.item2XTable))
		return item2XMap
	}

	ids := make([]interface{}, 0, len(missedIds))
	for _, id := range missedIds {
		ids = append(ids, id)
	}

	features, err := featureView.GetOnlineFeatures(ids, []string{d.xKey}, nil)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=Item2XFeatureStoreDao\terror=%v", context.RecommendId, err))
		return item2XMap
	}

	featureEntity := d.fsClient.GetProject().GetFeatureEntity(featureView.GetFeatureEntityName())
	if featureEntity == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=Item2XFeatureStoreDao\terror=featureEntity not found, featureEntity:%s", context.RecommendId, featureView.GetFeatureEntityName()))
		return item2XMap
	}

	for _, featureMap := range features {
		itemId := utils.ToString(featureMap[featureEntity.FeatureEntityJoinid], "")
		xVal := utils.ToString(featureMap[d.xKey], "")
		if itemId != "" && xVal != "" && !strings.EqualFold(xVal, "null") {
			item2XMap[itemId] = xVal
		}
	}

	putItem2XToCache(d.cache, missedIds, item2XMap)

	return item2XMap
}
