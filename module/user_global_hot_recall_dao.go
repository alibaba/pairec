package module

import (
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/recconf"
)

type UserGlobalHotRecallDao interface {
	ListItemsByUser(user *User, context *context.RecommendContext) []*Item
}

func NewUserGlobalHotRecallDao(config recconf.RecallConfig) UserGlobalHotRecallDao {
	if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewUserGlobalHotRecallHologresDao(config)
	} else if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_TableStore {
		return NewUserGlobalHotRecallTableStoreDao(config)
	}

	panic("not found UserGlobalHotRecallDao implement")
}
