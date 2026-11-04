package filter

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/feature"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/goburrow/cache"
)

type ItemStateFilter struct {
	name             string
	itemStateDao     module.ItemStateFilterDao
	featureTransList []*featureTrans
	cacheSign        string
}

func NewItemStateFilter(config recconf.FilterConfig) *ItemStateFilter {
	featureTransList := loadWithConfig(config.Features)
	if len(featureTransList) == 0 {
		filter := ItemStateFilter{
			name:         config.Name,
			itemStateDao: module.NewItemStateFilterDao(config, nil),
		}

		return &filter
	} else {
		filter := ItemStateFilter{
			name:             config.Name,
			featureTransList: featureTransList,
		}

		filter.itemStateDao = module.NewItemStateFilterDao(config, filter.invokeFeatureTrans)

		return &filter
	}
}

var itemStateCacheAfterFunc = time.AfterFunc

func itemStateCacheConfig(conf recconf.FilterConfig, old IFilter) (cache.Cache, string) {
	if conf.FilterType != "ItemStateFilter" || conf.ItemStateDaoConf.FeatureViewCacheName != "" ||
		(conf.ItemStateDaoConf.AdapterType != recconf.DataSource_Type_FeatureStore && conf.ItemStateDaoConf.AdapterType != recconf.DaoConf_Adapter_Hologres) {
		return nil, ""
	}
	conf.Features = nil
	conf.FilterParams = nil
	encoded, err := json.Marshal(conf)
	if err != nil {
		panic(err)
	}
	sign := utils.Md5(string(encoded))
	if f, ok := old.(*ItemStateFilter); ok && f != nil && f.cacheSign == sign {
		return f.itemStateCache(), sign
	}
	return nil, sign
}

func newItemStateFilterWithCache(conf recconf.FilterConfig, inherited cache.Cache, sign string) *ItemStateFilter {
	if sign == "" {
		return NewItemStateFilter(conf)
	}
	f := &ItemStateFilter{name: conf.Name, featureTransList: loadWithConfig(conf.Features), cacheSign: sign}
	var transFunc module.FeatureTransFunc
	if len(f.featureTransList) > 0 {
		transFunc = f.invokeFeatureTrans
	}
	f.itemStateDao = module.NewItemStateFilterDaoWithCache(conf, transFunc, inherited)
	return f
}

func (f *ItemStateFilter) itemStateCache() cache.Cache {
	switch dao := f.itemStateDao.(type) {
	case *module.ItemStateFilterFeatureStoreDao:
		return dao.ItemStateCache()
	case *module.ItemStateFilterHologresDao:
		return dao.ItemStateCache()
	}
	return nil
}

func invalidateItemStateCache(old, fresh cache.Cache) {
	if old != nil && old != fresh {
		itemStateCacheAfterFunc(30*time.Second, old.InvalidateAll)
	}
}

func (f *ItemStateFilter) Filter(filterData *FilterData) error {
	if _, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")

	}
	return f.doFilter(filterData)
}

func (f *ItemStateFilter) doFilter(filterData *FilterData) error {
	start := time.Now()
	items := filterData.Data.([]*module.Item)
	ctx := filterData.Context

	newItems := f.itemStateDao.Filter(filterData.User, items, ctx)

	filterData.Data = newItems
	filterInfoLog(filterData, "ItemStateFilter", f.name, len(newItems), start)
	return nil
}
func (f *ItemStateFilter) invokeFeatureTrans(user *module.User, item *module.Item, ctx *context.RecommendContext) {
	for _, trans := range f.featureTransList {
		trans.FeatureTran(user, item, ctx)
	}
	// clean user features in context。Context_User_Features_Key will be create by use expr to trans  features
	ctx.DeleteContextParam(feature.Context_User_Features_Key)
}

func loadWithConfig(config []recconf.FeatureConfig) []*featureTrans {
	trans := make([]*featureTrans, 0, len(config))
	for _, conf := range config {
		source := conf.FeatureSource
		if conf.FeatureValue != "" {
			source = conf.FeatureValue
		}
		ft := newFeatureTrans(conf.FeatureType, conf.FeatureName, conf.FeatureStore, source, conf.RemoveFeatureSource, conf.Normalizer, conf.Expression)
		trans = append(trans, ft)
	}

	return trans
}

type featureTrans struct {
	// featureType         string
	featureOp           feature.FeatureOp
	featureName         string
	featureSource       string
	removeFeatureSource bool   // delete feature source
	featureStore        string // user or item
	normalizer          feature.Normalizer
}

func newFeatureTrans(t, name, store string, source string, remove bool, normalizer, expression string) *featureTrans {
	return &featureTrans{
		// featureType: t,
		featureOp:           feature.NewFeatureOp(t),
		featureName:         name,
		featureStore:        store,
		featureSource:       source,
		removeFeatureSource: remove,
		normalizer:          feature.NewNormalizer(normalizer, expression),
	}
}

func (f *featureTrans) FeatureTran(user *module.User, item *module.Item, context *context.RecommendContext) {
	if f.featureStore == feature.SOURCE_ITEM {
		f.featureOp.ItemTransOp(f.featureName, f.featureSource, f.removeFeatureSource, f.normalizer, user, item, context)
	} else {
		f.featureOp.UserTransOp(f.featureName, f.featureSource, f.removeFeatureSource, f.normalizer, user, context)
	}
}
