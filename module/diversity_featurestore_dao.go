package module

import (
	"fmt"
	"time"

	pctx "github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/goburrow/cache"
)

type DiversityFeatureStoreDao struct {
	fsClient       *fs.FSClient
	table          string
	distinctFields []string
	cache          cache.Cache
}

func NewDiversityFeatureStoreDao(config recconf.DiversityDaoConfig) *DiversityFeatureStoreDao {
	fsclient, err := fs.GetFeatureStoreClient(config.FeatureStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("module=DiversityFeatureStoreDao\terror=%v", err))
		return nil
	}
	cacheTime := 30
	if config.CacheTimeInMinutes > 0 {
		cacheTime = config.CacheTimeInMinutes
	}
	cacheSize := 100000
	if config.CacheSize > 0 {
		cacheSize = config.CacheSize
	}
	d := &DiversityFeatureStoreDao{
		fsClient:       fsclient,
		table:          config.FeatureStoreViewName,
		distinctFields: config.DistinctFields,
		cache:          cache.New(cache.WithMaximumSize(cacheSize), cache.WithExpireAfterWrite(time.Duration(cacheTime)*time.Minute)),
	}
	return d
}

func (d *DiversityFeatureStoreDao) GetDistinctFields() []string {
	return d.distinctFields
}

func (d *DiversityFeatureStoreDao) GetDistinctValue(items []*Item, ctx *pctx.RecommendContext) error {
	itemMap := make(map[ItemId]*Item)

	itemIds := make([]interface{}, 0, len(items))
	for _, item := range items {
		if distinct, ok := d.cache.GetIfPresent(item.Id); ok {
			values := distinct.(map[string]interface{})
			item.AddProperties(values)
		} else {
			itemIds = append(itemIds, string(item.Id))
			itemMap[item.Id] = item
		}
	}

	if len(itemIds) == 0 {
		return nil
	}

	featureView := d.fsClient.GetProject().GetFeatureView(d.table)
	if featureView == nil {
		err := fmt.Errorf("featureView not found, table:%s", d.table)
		ctx.LogError(fmt.Sprintf("module=DiversityFeatureStoreDao\terror=%v", err))
		return err
	}
	features, err := featureView.GetOnlineFeatures(itemIds, d.distinctFields, map[string]string{})
	if err != nil {
		ctx.LogError(fmt.Sprintf("module=DiversityFeatureStoreDao\terror=featurestore error(%v)", err))
		return err
	}
	featureEntity := d.fsClient.GetProject().GetFeatureEntity(featureView.GetFeatureEntityName())
	if featureEntity == nil {
		err := fmt.Errorf("featureEntity not found, name:%s", featureView.GetFeatureEntityName())
		ctx.LogError(fmt.Sprintf("module=DiversityFeatureStoreDao\terror=%v", err))
		return err
	}

	resolved := make(map[ItemId]bool, len(features))
	for _, itemFeatures := range features {
		itemId := ItemId(utils.ToString(itemFeatures[featureEntity.FeatureEntityJoinid], ""))
		if itemId == "" {
			continue
		}
		distinct := make(map[string]interface{}, len(d.distinctFields))
		for _, field := range d.distinctFields {
			if value, ok := itemFeatures[field]; ok && value != nil {
				// keep raw value, align with FeatureFeatureStoreDao behavior
				distinct[field] = value
			}
		}
		d.cache.Put(itemId, distinct)
		resolved[itemId] = true
		if item, ok := itemMap[itemId]; ok {
			item.AddProperties(distinct)
		}
	}

	// negative cache support: only cache items that featurestore did not return
	for itemId := range itemMap {
		if !resolved[itemId] {
			d.cache.Put(itemId, map[string]interface{}{})
		}
	}

	return nil
}
