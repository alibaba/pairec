package module

import (
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/recconf"
)

type ColdStartRecallDao interface {
	ListItemsByUser(user *User, context *context.RecommendContext) []*Item
}

func NewColdStartRecallDao(config recconf.RecallConfig) ColdStartRecallDao {
	if config.ColdStartDaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewColdStartRecallHologresDao(config)
	}

	panic("not found ColdStartRecallDao implement")
}
