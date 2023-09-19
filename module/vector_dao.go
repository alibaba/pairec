package module

import (
	"errors"

	"github.com/alibaba/pairec/v2/recconf"
)

var (
	VectoryEmptyError = errors.New("vector empty")
)

type VectorDao interface {
	VectorString(id string) (string, error)
}

func NewVectorDao(config recconf.RecallConfig) VectorDao {
	if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_Redis {
		return NewVectorRedisDao(config)
	} else if config.DaoConf.AdapterType == recconf.DaoConf_Adapter_HBase {
		return NewVectorHBaseDao(config)
	} else if config.VectorDaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewVectorHologresDao(config)
	} else if config.VectorDaoConf.AdapterType == recconf.DaoConf_Adapter_Mysql {
		return NewVectorMysqlDao(config)
	} else if config.VectorDaoConf.AdapterType == recconf.DataSource_Type_ClickHouse {
		return NewVectorClickHouseDao(config)
	} else if config.VectorDaoConf.AdapterType == recconf.DataSource_Type_BE {
		return NewVectorBeDao(config)
	}

	panic("not found VectorDao implement")
}
