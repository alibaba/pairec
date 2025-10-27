package module

import (
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/recconf"
)

type ItemCollaborativeDao interface {
	ListItemsByItem(item *User, context *context.RecommendContext) []*Item
	ListItemsByMultiItemIds(item *User, context *context.RecommendContext, itemIds []any) map[string][]*Item
}

func NewItemCollaborativeDao(config recconf.RecallConfig) ItemCollaborativeDao {
	if config.ItemCollaborativeDaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewItemCollaborativeHologresDao(config)
	} else if config.ItemCollaborativeDaoConf.AdapterType == recconf.DataSource_Type_FeatureStore {
		return NewItemCollaborativeFeatureStoreDao(config)
	}
	panic("not found ItemCollaborativeDao implement")
}
