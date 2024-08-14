package module

import (
	"fmt"
	"github.com/alibaba/pairec/v2/context"
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
