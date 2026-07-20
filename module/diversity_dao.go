package module

import (
	"fmt"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
)

type DiversityDao interface {
	GetDistinctValue(items []*Item, ctx *context.RecommendContext) error
	GetDistinctFields() []string
}

func NewDiversityDao(config recconf.FilterConfig) DiversityDao {
	if config.DiversityDaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewDiversityHologresDao(config)
	}

	panic(fmt.Sprintf("DiversityDao:not found, name:%s", config.Name))
}

// NewDiversityDaoByConf creates DiversityDao by the adapter type of the conf.
// Unlike NewDiversityDao, it returns nil instead of panic when the adapter
// type is not supported, so the caller can degrade gracefully.
func NewDiversityDaoByConf(conf recconf.DiversityDaoConfig) DiversityDao {
	switch conf.AdapterType {
	case recconf.DataSource_Type_FeatureStore:
		if dao := NewDiversityFeatureStoreDao(conf); dao != nil {
			return dao
		}
		return nil
	case recconf.DaoConf_Adapter_Hologres:
		// reuse the existing constructor, keep its signature untouched
		if dao := NewDiversityHologresDao(recconf.FilterConfig{DiversityDaoConf: conf}); dao != nil {
			return dao
		}
		return nil
	default:
		log.Error(fmt.Sprintf("module=NewDiversityDaoByConf	error=unsupported adapter type:%s", conf.AdapterType))
		return nil
	}
}
