package sort

import (
	"errors"
	"time"

	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

type DiversityRuleSort struct {
	diversitySize    int
	diversityRules   []recconf.DiversityRuleConfig
	excludeRecallMap map[string]bool
	filterParam      *module.FilterParam
	cloneInstances   map[string]*DiversityRuleSort
	name             string
}

func NewDiversityRuleSort(config recconf.SortConfig) *DiversityRuleSort {
	sort := DiversityRuleSort{
		diversitySize:    config.DiversitySize,
		diversityRules:   config.DiversityRules,
		excludeRecallMap: make(map[string]bool, len(config.ExcludeRecalls)),
		filterParam:      nil,
		name:             config.Name,
		cloneInstances:   make(map[string]*DiversityRuleSort),
	}

	for _, recallName := range config.ExcludeRecalls {
		sort.excludeRecallMap[recallName] = true
	}

	if len(config.Conditions) > 0 {
		filterParam := module.NewFilterParamWithConfig(config.Conditions)
		sort.filterParam = filterParam
	}

	return &sort
}
func (s *DiversityRuleSort) Sort(sortData *SortData) error {
	if _, ok := sortData.Data.([]*module.Item); !ok {
		return errors.New("sort data type error")
	}

	// if condition is empty
	if s.filterParam == nil {
		return s.doSort(sortData)
	} else {
		userProperties := sortData.User.MakeUserFeatures2()
		flag, err := s.filterParam.EvaluateByDomain(userProperties, nil)
		if err != nil {
			return err
		}
		if flag {
			return s.doSort(sortData)
		}
	}

	return nil
}

func (s *DiversityRuleSort) createDiversityRules() (ret []*DiversityRule) {
	for _, config := range s.diversityRules {
		rule := NewDiversityRule(config)

		ret = append(ret, rule)
	}

	return
}
func (s *DiversityRuleSort) doSort(sortData *SortData) error {
	start := time.Now()
	items := sortData.Data.([]*module.Item)

	diversityRules := s.createDiversityRules()
	if len(diversityRules) == 0 {
		return nil
	}

	var excludeItems []*module.Item
	if len(s.excludeRecallMap) > 0 {
		newItems := make([]*module.Item, 0, len(items))
		for _, item := range items {
			if _, ok := s.excludeRecallMap[item.GetRecallName()]; ok {
				excludeItems = append(excludeItems, item)
			} else {
				newItems = append(newItems, item)
			}
		}

		items = newItems
	}

	itemLength := len(items)
	//if items empty
	if itemLength == 0 {
		return nil
	}

	diversitySize := sortData.Context.Size

	if s.diversitySize > 0 {
		diversitySize = s.diversitySize
		if diversitySize > itemLength {
			diversitySize = itemLength
		}
	}

	result := make([]*module.Item, 0, diversitySize)
	alreadyMatchItems := make(map[module.ItemId]bool, diversitySize)
	alreadyMatchItems[items[0].Id] = true
	result = append(result, items[0])
	items = items[1:]

	index := 1
	for len(result) <= diversitySize {
		if index == itemLength {
			break
		}

		flag := true
		// if all the rest items not match diversity rule, use the first item append to the result
		firstItemIndex := -1
		for i, item := range items {
			if _, ok := alreadyMatchItems[item.Id]; ok {
				continue
			}

			if firstItemIndex == -1 {
				firstItemIndex = i
			}
			flag = true
			for _, rule := range diversityRules {
				if flag = rule.Match(item, result); !flag {
					break
				}
			}

			// if the item match all the diversity rule, so add it to the result
			if flag {
				alreadyMatchItems[item.Id] = true
				result = append(result, item)
				index++
				break
			}
		}

		if !flag {
			alreadyMatchItems[items[firstItemIndex].Id] = true
			result = append(result, items[firstItemIndex])
			index++
		}
	}

	for _, item := range items {
		if _, ok := alreadyMatchItems[item.Id]; ok {
			continue
		}
		result = append(result, item)
	}

	result = append(result, excludeItems...)

	sortData.Data = result
	sortInfoLog(sortData, "DiversityRuleSort", len(result), start)
	return nil
}
