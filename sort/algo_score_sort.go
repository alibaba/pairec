package sort

import (
	"errors"
	"fmt"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"sort"
)

type AlgoScoreSort struct {
	sortByField     string
	switchThreshold float64
}

func NewAlgoScoreSort(config recconf.SortConfig) *AlgoScoreSort {
	if config.SortByField == "" {
		log.Error(fmt.Sprintf("sort by fields is not configured for sort %s", config.Name))
		config.SortByField = "current_score"
	}
	return &AlgoScoreSort{
		sortByField:     config.SortByField,
		switchThreshold: config.SwitchThreshold,
	}
}

func GetMaxScore(items []*module.Item) float64 {
	maxScore := -1e300
	for _, item := range items {
		if item.Score > maxScore {
			maxScore = item.Score
		}
	}
	return maxScore
}

func (s *AlgoScoreSort) Sort(sortData *SortData) error {
	items, ok := sortData.Data.([]*module.Item)
	if !ok {
		return errors.New("sort data type error")
	}

	sortByField := s.sortByField
	maxRankScore := GetMaxScore(items)
	if maxRankScore > s.switchThreshold {
		sortByField = "current_score"
	}

	ctx := sortData.Context
	sort.Slice(items, func(i, j int) bool {
		iScore, err1 := items[i].FloatExprData(sortByField)
		if err1 != nil {
			iScore = items[i].Score
			ctx.LogWarning(fmt.Sprintf("get sort field %s from item %s failed", sortByField, items[i].Id))
		}
		jScore, err2 := items[j].FloatExprData(sortByField)
		if err2 != nil {
			iScore = items[j].Score
			ctx.LogWarning(fmt.Sprintf("get sort field %s from item %s failed", sortByField, items[j].Id))
		}
		return iScore > jScore
	})
	sortData.Data = items
	return nil
}
