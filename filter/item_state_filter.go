package filter

import (
	"errors"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/feature"
)

type ItemStateFilter struct {
	name             string
	itemStateDao     module.ItemStateFilterDao
	featureTransList []*featureTrans
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
