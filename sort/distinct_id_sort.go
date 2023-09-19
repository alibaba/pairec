package sort

import (
	"errors"
	"time"

	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

type DistinctIdCondition struct {
	filterParam *module.FilterParam
	distinctId  int
}

func NewDistinctIdCondition(config *recconf.DistinctIdCondition) *DistinctIdCondition {
	filterParam := module.NewFilterParamWithConfig(config.Conditions)
	condition := DistinctIdCondition{
		filterParam: filterParam,
		distinctId:  config.DistinctId,
	}
	return &condition
}

type DistinctIdSort struct {
	name       string
	conditions []*DistinctIdCondition
}

func NewDistinctIdSort(config recconf.SortConfig) *DistinctIdSort {
	sort := DistinctIdSort{name: config.Name}
	for _, boostScoreConditionConfig := range config.DistinctIdConditions {
		condition := NewDistinctIdCondition(&boostScoreConditionConfig)
		sort.conditions = append(sort.conditions, condition)
	}
	return &sort
}

func (s *DistinctIdSort) Sort(sortData *SortData) error {
	if _, ok := sortData.Data.([]*module.Item); !ok {
		return errors.New("sort data type error")
	}

	return s.doSort(sortData)
}

func (s *DistinctIdSort) doSort(sortData *SortData) error {
	start := time.Now()
	items := sortData.Data.([]*module.Item)
	//ctx := sortData.Context
	userProperties := sortData.User.MakeUserFeatures2()
	for i, item := range items {
		item.AddProperty("__distinct_id__", i+1)
		properties := item.GetProperties()
		for _, condition := range s.conditions {
			if flag, err := condition.filterParam.EvaluateByDomain(userProperties, properties); err == nil && flag {
				item.AddProperty("__distinct_id__", condition.distinctId)
				//ctx.LogDebug(fmt.Sprintf("module=DistinctIdSort\titem=%s\tdistinct_id=%d", item.Id, condition.distinctId))
				break
			}
		}
	}
	sortData.Data = items
	sortInfoLogWithName(sortData, "DistinctIdSort", s.name, len(items), start)
	return nil
}
