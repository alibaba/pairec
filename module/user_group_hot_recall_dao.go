package module

import (
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/recconf"
)

type UserGroupHotRecallDao interface {
	ListItemsByUser(user *User, context *context.RecommendContext) []*Item
}

func NewUserGroupHotRecallDao(config recconf.RecallConfig) UserGroupHotRecallDao {
	// if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_Mysql {
	// return NewUserCusteomRecallMysqlDao(config)
	// } else if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_TableStore {
	// return NewUserGroupHotRecallTableStoreDao(config)
	// } else if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
	// return NewUserCusteomRecallHologresDao(config)
	// }

	if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewUserGroupHotRecallHologresDao(config)
	}

	panic("not found UserGroupHotRecallDao implement")
}
