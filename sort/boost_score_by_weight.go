package sort

import (
	"errors"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"time"
)

type BoostScoreByWeight struct {
	BoostScoreByWeightDao module.BoostScoreByWeightDao
}

func NewBoostScoreByWeight(config recconf.SortConfig) *BoostScoreByWeight {
	boostScoreByWeight := BoostScoreByWeight{
		BoostScoreByWeightDao: module.NewBoostScoreByWeightDao(config),
	}

	return &boostScoreByWeight
}

func (s *BoostScoreByWeight) Sort(sortData *SortData) error {
	if _, ok := sortData.Data.([]*module.Item); !ok {
		return errors.New("sort data type error")
	}

	return s.doSort(sortData)
}

func (s *BoostScoreByWeight) doSort(sortData *SortData) error {
	start := time.Now()
	items := sortData.Data.([]*module.Item)

	resultItems := s.BoostScoreByWeightDao.Sort(items)

	sortData.Data = resultItems
	sortInfoLog(sortData, "BoostScoreByWeight", len(resultItems), start)
	return nil
}
