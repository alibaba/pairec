package module

import (
	"fmt"

	"github.com/alibaba/pairec/v2/recconf"
)

type User2ItemCustomFilterDao interface {
	Filter(uid UID, ids []*Item) (ret []*Item)
}

func NewUser2ItemCustomFilterDao(config recconf.FilterConfig) User2ItemCustomFilterDao {
	if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_TableStore {
		return NewUser2ItemCustomFilterTableStoreDao(config)
	}

	panic(fmt.Sprintf("User2ItemCustomFilterDao:not found, name:%s", config.Name))
}
