package feature

import (
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

// Feature is a type for load user or item feature
type Feature struct {
	// featureDao use for load feature from specific dao, example:redis or hologres
	featureDao module.FeatureDao
	// featureTrans use for transform feature, specific operator by FeatureOp
	featureTrans []FeatureTrans
}

func LoadWithConfig(config recconf.FeatureLoadConfig) *Feature {
	f := &Feature{
		featureDao:   module.NewFeatureDao(config.FeatureDaoConf),
		featureTrans: make([]FeatureTrans, 0),
	}

	for _, conf := range config.Features {
		source := conf.FeatureSource
		if conf.FeatureValue != "" {
			source = conf.FeatureValue
		}
		ft := NewFeatureTrans(conf.FeatureType, conf.FeatureName, conf.FeatureStore, source, conf.RemoveFeatureSource, conf.Normalizer, conf.Expression)
		f.featureTrans = append(f.featureTrans, ft)
	}

	return f
}

func (f *Feature) LoadFeatures(user *module.User, items []*module.Item, context *context.RecommendContext) {
	f.featureDao.FeatureFetch(user, items, context)
	for _, trans := range f.featureTrans {
		trans.FeatureTran(user, items, context)
	}
}

const (
	SOURCE_USER = "user"
	SOURCE_ITEM = "item"
)

type FeatureTrans interface {
	FeatureTran(user *module.User, items []*module.Item, context *context.RecommendContext)
}

type featureTrans struct {
	// featureType         string
	featureOp           FeatureOp
	featureName         string
	featureSource       string
	removeFeatureSource bool   // delete feature source
	featureStore        string // user or item
	normalizer          Normalizer
}

func NewFeatureTrans(t, name, store string, source string, remove bool, normalizer, expression string) FeatureTrans {
	return &featureTrans{
		// featureType: t,
		featureOp:           NewFeatureOp(t),
		featureName:         name,
		featureStore:        store,
		featureSource:       source,
		removeFeatureSource: remove,
		normalizer:          NewNormalizer(normalizer, expression),
	}
}

func (f *featureTrans) FeatureTran(user *module.User, items []*module.Item, context *context.RecommendContext) {
	if f.featureStore == SOURCE_ITEM {
		for _, item := range items {
			f.featureOp.ItemTransOp(f.featureName, f.featureSource, f.removeFeatureSource, f.normalizer, user, item, context)
		}

	} else {
		f.featureOp.UserTransOp(f.featureName, f.featureSource, f.removeFeatureSource, f.normalizer, user, context)
	}
}
