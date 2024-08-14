package module

import (
	"fmt"

	"github.com/alibaba/pairec/v2/recconf"
)

type ItemStateFilterDao interface {
	Filter(user *User, ids []*Item) (ret []*Item)
}

func NewItemStateFilterDao(config recconf.FilterConfig) ItemStateFilterDao {
	if config.ItemStateDaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewItemStateFilterHologresDao(config)
	} else if config.ItemStateDaoConf.AdapterType == recconf.DaoConf_Adapter_TableStore {
		return NewItemStateFilterTablestoreDao(config)
	} else if config.ItemStateDaoConf.AdapterType == recconf.DataSource_Type_HBase_Thrift {
		return NewItemStateFilterHBaseThriftDao(config)
	}

	panic(fmt.Sprintf("ItemStateFilterDao:not found, name:%s", config.Name))
}
