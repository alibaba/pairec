package module

import (
	"fmt"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/goburrow/cache"
)

type ItemStateFilterDao interface {
	Filter(user *User, ids []*Item, ctx *context.RecommendContext) (ret []*Item)
}

type FeatureTransFunc func(user *User, item *Item, ctx *context.RecommendContext)

func NewItemStateFilterDao(config recconf.FilterConfig, f FeatureTransFunc) ItemStateFilterDao {
	// FeatureViewCache takes priority over AdapterType
	if config.ItemStateDaoConf.FeatureViewCacheName != "" {
		return NewItemStateFilterFeatureViewCacheDao(config, f)
	}
	if config.ItemStateDaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewItemStateFilterHologresDao(config, f)
	} else if config.ItemStateDaoConf.AdapterType == recconf.DaoConf_Adapter_TableStore {
		return NewItemStateFilterTablestoreDao(config)
	} else if config.ItemStateDaoConf.AdapterType == recconf.DataSource_Type_HBase_Thrift {
		return NewItemStateFilterHBaseThriftDao(config)
	} else if config.ItemStateDaoConf.AdapterType == recconf.DataSource_Type_FeatureStore {
		return NewItemStateFilterFeatureStoreDao(config, f)
	}

	panic(fmt.Sprintf("ItemStateFilterDao:not found, name:%s", config.Name))
}

func NewItemStateFilterDaoWithCache(config recconf.FilterConfig, f FeatureTransFunc, inherited cache.Cache) ItemStateFilterDao {
	if config.ItemStateDaoConf.FeatureViewCacheName != "" ||
		(config.ItemStateDaoConf.AdapterType != recconf.DataSource_Type_FeatureStore && config.ItemStateDaoConf.AdapterType != recconf.DaoConf_Adapter_Hologres) {
		return NewItemStateFilterDao(config, f)
	}
	baseConfig := config
	baseConfig.ItemStateCacheSize = 0
	dao := NewItemStateFilterDao(baseConfig, f)
	c := inherited
	if c == nil && config.ItemStateCacheSize > 0 {
		cacheTime := 3600
		if config.ItemStateCacheTime > 0 {
			cacheTime = config.ItemStateCacheTime
		}
		c = cache.New(cache.WithMaximumSize(config.ItemStateCacheSize),
			cache.WithExpireAfterWrite(time.Second*time.Duration(cacheTime)))
	}
	switch d := dao.(type) {
	case *ItemStateFilterFeatureStoreDao:
		d.itmCache = c
	case *ItemStateFilterHologresDao:
		d.itmCache = c
	}
	return dao
}
