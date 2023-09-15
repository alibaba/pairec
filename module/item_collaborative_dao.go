package module

import (
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/recconf"
)

type ItemCollaborativeDao interface {
	ListItemsByItem(item *User, context *context.RecommendContext) []*Item
}

func NewItemCollaborativeDao(config recconf.RecallConfig) ItemCollaborativeDao {
	if config.ItemCollaborativeDaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewItemCollaborativeHologresDao(config)
	}
	panic("not found ItemCollaborativeDao implement")
}
