package sort

import (
	"errors"
	"fmt"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

// ConditionSortItem represents a condition-sort mapping
type ConditionSortItem struct {
	filterParam *module.FilterParam
	sortName    string
}

// ConditionSort routes to different sort strategies based on user attributes
type ConditionSort struct {
	sortItems       []*ConditionSortItem
	defaultSortName string
	name            string
}

// NewConditionSort creates a new ConditionSort from config
func NewConditionSort(config recconf.SortConfig) *ConditionSort {
	var items []*ConditionSortItem
	for _, item := range config.ConditionSortConfs.SortConfs {
		sortItem := &ConditionSortItem{
			sortName: item.SortName,
		}
		if len(item.Conditions) > 0 {
			sortItem.filterParam = module.NewFilterParamWithConfig(item.Conditions)
		}
		items = append(items, sortItem)
	}

	return &ConditionSort{
		sortItems:       items,
		defaultSortName: config.ConditionSortConfs.DefaultSortName,
		name:            config.Name,
	}
}

// Sort implements ISort interface
func (s *ConditionSort) Sort(sortData *SortData) error {
	if _, ok := sortData.Data.([]*module.Item); !ok {
		return errors.New("sort data type error")
	}

	userProperties := sortData.User.MakeUserFeatures2()

	for _, item := range s.sortItems {
		if item.filterParam != nil {
			if flag, err := item.filterParam.EvaluateByDomain(userProperties, nil); err == nil {
				if flag {
					sort, err := GetSort(item.sortName)
					if err != nil {
						log.Error(fmt.Sprintf("requestId=%s\tmodule=ConditionSort\tsortName=%s\terror=%v",
							sortData.Context.RecommendId, item.sortName, err))
						return err
					}
					log.Info(fmt.Sprintf("requestId=%s\tmodule=ConditionSort\tsortName=%s",
						sortData.Context.RecommendId, item.sortName))
					return sort.Sort(sortData)
				}
			} else {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=ConditionSort\tsortName=%s\terror=%v",
					sortData.Context.RecommendId, item.sortName, err))
			}
		}
	}

	// Use default sort if no condition matched
	if len(s.defaultSortName) > 0 {
		sort, err := GetSort(s.defaultSortName)
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=ConditionSort\tdefaultSortName=%s\terror=%v",
				sortData.Context.RecommendId, s.defaultSortName, err))
			return err
		}
		log.Info(fmt.Sprintf("requestId=%s\tmodule=ConditionSort\tsortName=%s",
			sortData.Context.RecommendId, s.defaultSortName))
		return sort.Sort(sortData)
	}

	return nil
}
