package module

import (
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/recconf"
)

type UserGroupHotRecallDao interface {
	ListItemsByUser(user *User, context *context.RecommendContext) []*Item
	TriggerValue(user *User) string
}

func NewUserGroupHotRecallDao(config recconf.RecallConfig) UserGroupHotRecallDao {
	if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewUserGroupHotRecallHologresDao(config)
	} else if config.DaoConf.AdapterType == recconf.DataSource_Type_FeatureStore {
		return NewUserGroupHotRecallFeatureStoreDao(config)
	}

	panic("not found UserGroupHotRecallDao implement")
}
