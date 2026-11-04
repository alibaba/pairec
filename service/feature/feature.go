package feature

import (
	"encoding/json"
	"fmt"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/goburrow/cache"
)

// Feature is a type for load user or item feature
type Feature struct {
	// featureDao use for load feature from specific dao, example:redis or hologres
	featureDao module.FeatureDao
	// featureTrans use for transform feature, specific operator by FeatureOp
	featureTrans []FeatureTrans
	daoSign      string
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

	prepareFSEmbeddings(f)
	return f
}

func featureDaoSign(config recconf.FeatureDaoConfig) string {
	encoded, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	return utils.Md5(string(encoded))
}

func (f *Feature) featureCache() cache.Cache {
	switch dao := f.featureDao.(type) {
	case *module.FeatureFeatureStoreDao:
		return dao.FeatureCache()
	case *module.FeatureHologresDao:
		return dao.FeatureCache()
	}
	return nil
}

func loadWithCache(config recconf.FeatureLoadConfig, inherited cache.Cache) *Feature {
	if (config.FeatureDaoConf.AdapterType != recconf.DataSource_Type_FeatureStore && config.FeatureDaoConf.AdapterType != recconf.DaoConf_Adapter_Hologres) || config.FeatureDaoConf.FeatureCacheName != "" {
		return LoadWithConfig(config)
	}
	sign := featureDaoSign(config.FeatureDaoConf)
	f := &Feature{featureDao: module.NewFeatureDaoWithCache(config.FeatureDaoConf, inherited),
		featureTrans: make([]FeatureTrans, 0), daoSign: sign}
	if utils.IsNil(f.featureDao) {
		panic(fmt.Sprintf("feature DAO construction failed, adapter:%s", config.FeatureDaoConf.AdapterType))
	}
	completed := false
	defer func() {
		if !completed && inherited == nil {
			if c := f.featureCache(); c != nil {
				_ = c.Close()
			}
		}
	}()
	for _, conf := range config.Features {
		source := conf.FeatureSource
		if conf.FeatureValue != "" {
			source = conf.FeatureValue
		}
		f.featureTrans = append(f.featureTrans, NewFeatureTrans(conf.FeatureType, conf.FeatureName,
			conf.FeatureStore, source, conf.RemoveFeatureSource, conf.Normalizer, conf.Expression))
	}
	prepareFSEmbeddings(f)
	completed = true
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
