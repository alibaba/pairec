package module

import (
	"fmt"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/recconf"
)

type ItemStateFilterDao interface {
	Filter(user *User, ids []*Item, ctx *context.RecommendContext) (ret []*Item)
}

type FeatureTransFunc func(user *User, item *Item, ctx *context.RecommendContext)

func NewItemStateFilterDao(config recconf.FilterConfig, f FeatureTransFunc) ItemStateFilterDao {
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
