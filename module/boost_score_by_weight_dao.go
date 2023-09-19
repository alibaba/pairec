package module

import (
	"fmt"
	"github.com/alibaba/pairec/v2/recconf"
)

type BoostScoreByWeightDao interface {
	Sort(items []*Item) (resultItems []*Item)
}

func NewBoostScoreByWeightDao(config recconf.SortConfig) BoostScoreByWeightDao {
	if config.BoostScoreByWeightDao.AdapterType == recconf.DaoConf_Adapter_Hologres {
		return NewBoostScoreByWeightHologresDao(config)
	}
	panic(fmt.Sprintf("BoostScoreByWeightByWeightDao:not found, name:%s", config.Name))
}
