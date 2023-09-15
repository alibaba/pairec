package module

import (
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/recconf"
)

type User2ItemDao interface {
	ListItemsByUser(user *User, context *context.RecommendContext) []*Item
}

func NewUser2ItemDao(config recconf.RecallConfig) User2ItemDao {
	if config.User2ItemDaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewUser2ItemHologresDao(config)
	}
	panic("not found User2ItemDao implement")
}
