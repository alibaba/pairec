package module

import (
	"fmt"

	"github.com/alibaba/pairec/v2/recconf"
)

type ItemCustomFilterDao interface {
	GetFilterItems() (ret map[ItemId]bool)
}

func NewItemCustomFilterDao(config recconf.FilterConfig) ItemCustomFilterDao {
	if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_TableStore {
		return NewItemCustomFilterTableStoreDao(config)
	}
	if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewItemCustomFilterHoloDao(config)
	}

	panic(fmt.Sprintf("ItemCustomFilterDao:not found, name:%s", config.Name))
}
