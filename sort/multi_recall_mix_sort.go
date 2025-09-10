package sort

import (
	"errors"
	"fmt"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

type MultiRecallMixSort struct {
	name           string
	remainItem     bool
	size           int
	mixRules       []recconf.MixSortConfig
	cloneInstances map[string]*MultiRecallMixSort
}

func NewMultiRecallMixSort(config recconf.SortConfig) *MultiRecallMixSort {
	sort := MultiRecallMixSort{
		name:           config.Name,
		remainItem:     config.RemainItem,
		mixRules:       config.MixSortRules,
		size:           config.Size,
		cloneInstances: make(map[string]*MultiRecallMixSort),
	}

	return &sort
}
func (s *MultiRecallMixSort) Sort(sortData *SortData) error {
	if _, ok := sortData.Data.([]*module.Item); !ok {
		return errors.New("sort data type error")
	}

	return s.doSort(sortData)
}

func (s *MultiRecallMixSort) createMixStrategies(sortData *SortData) []MixSortStrategy {
	var strategies []MixSortStrategy
	size := sortData.Context.Size
	if s.size > size {
		size = s.size
	}

	for _, config := range s.mixRules {
		switch config.MixStrategy {
		case "fix_position":
			strategy := newFixPositionStrategy(&config)
			strategy.totalSize = size
			strategies = append(strategies, strategy)
		case "random_position":
			strategy := newRandomPositionStrategy(&config, size)
			strategy.totalSize = size
			strategies = append(strategies, strategy)
		}
	}

	return strategies
}
func (s *MultiRecallMixSort) doSort(sortData *SortData) error {
	items := sortData.Data.([]*module.Item)

	size := sortData.Context.Size
	if s.size > size {
		size = s.size
	}

	if len(items) < size {
		return nil
	}

	result := make([]*module.Item, size)

	strategies := s.createMixStrategies(sortData)
	defaultStrategy := newDefaultStrategy(nil, size)
	userProperties := sortData.User.MakeUserFeatures2()

	found := false
	for _, item := range items {
		found = false
		properties := item.GetFeatures()
		for _, strategy := range strategies {
			if strategy.ContainsRecallName(item.GetRecallName()) {
				if !strategy.IsFull() {
					found = true
					strategy.AppendItem(item)
					break
				}
			}
			if strategy.IsUseCondition() {
				ok, err := strategy.EvaluateByDomain(userProperties, properties)
				if err != nil {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=MultiRecallMixSort\titemId=%s\terror=%v", sortData.Context.RecommendId, item.Id, err))
					break
				}
				if ok {
					if !strategy.IsFull() {
						found = ok
						strategy.AppendItem(item)
						break

					}
				}
			}
		}

		if !found {
			defaultStrategy.AppendItem(item)
		}

		if defaultStrategy.IsFull() {
			flag := true
			for _, strategy := range strategies {
				if !strategy.IsFull() {
					flag = false
					break
				}
			}

			if flag {
				break
			}

		}
	}

	for _, strategy := range strategies {
		if strategy.GetStrategyType() == FixPositionStrategyType {
			result = strategy.BuildItems(result)
		}
	}
	for _, strategy := range strategies {
		if strategy.GetStrategyType() == RandomPositionStrategyType {
			result = strategy.BuildItems(result)
		}
	}

	result = defaultStrategy.BuildItems(result)

	size = len(result)
	var iterator *remainItemIterator
	for i := 0; i < size; i++ {
		if result[i] == nil {
			if iterator == nil {
				iterator = newRemainItemIterator(result, items)
			}
			if item, exist := iterator.findRemainItem(); exist {
				result[i] = item
			} else {
				result[i] = result[size-1]
				result = result[:size-1]
				i--
				size--
			}
		}
	}

	if s.remainItem {
		if iterator == nil {
			iterator = newRemainItemIterator(result, items)
		}
		for {
			if item, exist := iterator.findRemainItem(); exist {
				result = append(result, item)
			} else {
				break
			}

		}

	}
	sortData.Data = result
	return nil
}

type remainItemIterator struct {
	index     int
	totalSize int
	resultMap map[module.ItemId]bool
	items     []*module.Item
}

func newRemainItemIterator(result []*module.Item, items []*module.Item) *remainItemIterator {
	iterator := remainItemIterator{
		index:     0,
		totalSize: len(items),
		items:     items,
		resultMap: make(map[module.ItemId]bool),
	}
	for _, item := range result {
		if item != nil {
			iterator.resultMap[item.Id] = true
		}
	}

	return &iterator
}

func (iterator *remainItemIterator) findRemainItem() (*module.Item, bool) {
	var item *module.Item
	for iterator.index < iterator.totalSize {
		if _, ok := iterator.resultMap[iterator.items[iterator.index].Id]; !ok {
			item = iterator.items[iterator.index]
			iterator.resultMap[item.Id] = true
			iterator.index++
			break
		}
		iterator.index++
	}

	return item, item != nil
}
