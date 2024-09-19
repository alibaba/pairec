package module

import (
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/recconf"
)

type UserCustomRecallDao interface {
	ListItemsByUser(user *User, context *context.RecommendContext) []*Item
}

func NewUserCustomRecallDao(config recconf.RecallConfig) UserCustomRecallDao {
	if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_Mysql {
		return NewUserCustomRecallMysqlDao(config)
	} else if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_TableStore {
		return NewUserCustomRecallTableStoreDao(config)
	} else if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewUserCustomRecallHologresDao(config)
	} else if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_Redis {
		return NewUserCustomRecallRedisDao(config)
	} else if config.DaoConf.AdapterType == recconf.DataSource_Type_ClickHouse {
		return NewUserCustomRecallClickHouseDao(config)
	} else if config.DaoConf.AdapterType == recconf.DataSource_Type_FeatureStore {
		return NewUserCustomRecallFeatureStoreDao(config)
	}

	panic("not found UserCustomRecallDao implement")
}
