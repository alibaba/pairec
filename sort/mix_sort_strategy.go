package sort

import (
	"math/rand"

	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type MixSortStrategyType int

const (
	FixPositionStrategyType MixSortStrategyType = 1

	RandomPositionStrategyType MixSortStrategyType = 2

	DefaultStrategyType MixSortStrategyType = 3
)

type MixSortStrategy interface {
	ContainsRecallName(name string) bool
	AppendItem(item *module.Item)
	IsFull() bool
	GetStrategyType() MixSortStrategyType
	BuildItems(items []*module.Item) []*module.Item
	Evaluate(properties map[string]interface{}) (bool, error)
	EvaluateByDomain(userProperties, itemProperties map[string]interface{}) (bool, error)
	IsUseCondition() bool
}
type mixSortStrategy struct {
	number        int
	totalSize     int
	index         int
	strategyType  MixSortStrategyType
	recallNameMap map[string]bool
	items         []*module.Item
	filterParam   *module.FilterParam
}

func (s *mixSortStrategy) ContainsRecallName(name string) bool {
	if s.strategyType == DefaultStrategyType {
		return true
	}

	if _, ok := s.recallNameMap[name]; ok {
		return true
	}

	return false
}

func (s *mixSortStrategy) Evaluate(properties map[string]interface{}) (bool, error) {
	ok, err := s.filterParam.Evaluate(properties)
	if err != nil {
		return false, err
	}
	return ok, nil
}
func (s *mixSortStrategy) EvaluateByDomain(userProperties, itemProperties map[string]interface{}) (bool, error) {
	ok, err := s.filterParam.EvaluateByDomain(userProperties, itemProperties)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (s *mixSortStrategy) AppendItem(item *module.Item) {
	if len(s.items) == s.number {
		return
	}
	s.items = append(s.items, item)
}
func (s *mixSortStrategy) IsFull() bool {
	return len(s.items) >= s.number
}

func (s *mixSortStrategy) IsUseCondition() bool {
	/**
	if s.filterParam != nil {
		return true
	}
	return false
	**/
	return s.filterParam != nil
}

func (s *mixSortStrategy) GetStrategyType() MixSortStrategyType {
	return s.strategyType
}

func newMixSortStrategy(config *recconf.MixSortConfig) *mixSortStrategy {
	strategy := mixSortStrategy{
		number:        0,
		index:         0,
		recallNameMap: make(map[string]bool),
	}

	if config != nil {
		for _, name := range config.RecallNames {
			strategy.recallNameMap[name] = true
		}

	}
	if config != nil {
		if len(config.Conditions) > 0 {
			filterParam := module.NewFilterParamWithConfig(config.Conditions)
			strategy.filterParam = filterParam
		}
	}

	return &strategy
}

type fixPositionStrategy struct {
	*mixSortStrategy
	positions     []int
	positionField string
}

func newFixPositionStrategy(config *recconf.MixSortConfig) *fixPositionStrategy {
	strategy := fixPositionStrategy{
		positions:       config.Positions,
		positionField:   config.PositionField,
		mixSortStrategy: newMixSortStrategy(config),
	}

	strategy.number = len(strategy.positions)
	if strategy.number == 0 && config.Number > 0 {
		strategy.number = config.Number
	}

	strategy.strategyType = FixPositionStrategyType

	return &strategy
}
func (s *fixPositionStrategy) BuildItems(items []*module.Item) []*module.Item {
	if len(s.positions) == 0 && s.positionField != "" {
		for i := 0; i < s.number; i++ {
			if i < len(s.items) {
				if p := utils.ToInt(s.items[i].GetProperty(s.positionField), -1); p > 0 {
					s.positions = append(s.positions, p)
				}
			}
		}
	}
	for _, pos := range s.positions {
		if pos <= s.totalSize && s.index < len(s.items) {
			items[pos-1] = s.items[s.index]
			s.index++
		}
	}

	return items
}

type randomPositionStrategy struct {
	*mixSortStrategy
}

func newRandomPositionStrategy(config *recconf.MixSortConfig, size int) *randomPositionStrategy {
	strategy := randomPositionStrategy{
		mixSortStrategy: newMixSortStrategy(config),
	}

	if config.NumberRate > float64(0) {
		strategy.number = int(float64(size) * config.NumberRate)
	} else if config.Number > 0 {
		strategy.number = config.Number
	}

	strategy.strategyType = RandomPositionStrategyType

	return &strategy

}

func (s *randomPositionStrategy) BuildItems(items []*module.Item) []*module.Item {
	start := 0
	end := 0
	for _, item := range s.items {
		end = start + rand.Intn(s.totalSize/s.number)
		if end >= s.totalSize {
			end = s.totalSize - 1
		}

		for items[end] != nil {
			end++
			end = end % s.totalSize
		}
		items[end] = item

		start += s.totalSize / s.number
	}

	return items
}

type defaultStrategy struct {
	*mixSortStrategy
}

func newDefaultStrategy(config *recconf.MixSortConfig, size int) *defaultStrategy {
	strategy := defaultStrategy{
		mixSortStrategy: newMixSortStrategy(config),
	}

	strategy.totalSize = size
	strategy.number = size

	strategy.strategyType = DefaultStrategyType

	return &strategy

}

func (s *defaultStrategy) BuildItems(items []*module.Item) []*module.Item {
	for i, item := range items {
		if item == nil && s.index < len(s.items) {
			items[i] = s.items[s.index]
			s.index++
		}
	}

	return items
}
