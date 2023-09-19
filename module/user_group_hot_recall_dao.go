package module

import (
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/recconf"
)

type UserGroupHotRecallDao interface {
	ListItemsByUser(user *User, context *context.RecommendContext) []*Item
}

func NewUserGroupHotRecallDao(config recconf.RecallConfig) UserGroupHotRecallDao {
	if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewUserGroupHotRecallHologresDao(config)
	}

	panic("not found UserGroupHotRecallDao implement")
}
