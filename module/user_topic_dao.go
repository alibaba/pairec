package module

import (
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/recconf"
)

type UserTopicDao interface {
	ListItemsByUser(user *User, context *context.RecommendContext) []*Item
}

func NewUserTopicDao(config recconf.RecallConfig) UserTopicDao {

	if config.UserTopicDaoConf.AdapterType == recconf.DaoConf_Adapter_Mysql {
		return NewUserTopicMysqlDao(config)
	} else if config.UserTopicDaoConf.AdapterType == recconf.DaoConf_Adapter_TableStore {
		return NewUserTopicTableStoreDao(config)
	} else {
		panic("UserTopicDao not implement")
	}
}
