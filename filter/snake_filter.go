package filter

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

type SnakeConfigType uint8

const (
	Snake_Type_Refill SnakeConfigType = iota + 1
	Snake_Type_Skip
)

type snakeAdjustCountConfig struct {
	Id         int
	Weight     int
	RecallName string
	Count      int
	Type       SnakeConfigType
}

func newSnakeAdjustCountConfig(config recconf.AdjustCountConfig) (*snakeAdjustCountConfig, error) {
	c := &snakeAdjustCountConfig{
		Weight:     config.Weight,
		RecallName: config.RecallName,
	}

	return c, nil
}

type snakeItemIterator struct {
	items               []*module.Item
	scoreMap            map[module.ItemId]float64
	index               int // current index for items
	recallName          string
	config              *snakeAdjustCountConfig
	alreadyExistItemMap map[module.ItemId]bool
	itemRankMap         map[module.ItemId][]string
}

func newSnakeItemIterator(config *snakeAdjustCountConfig, alreadyExistItemMap map[module.ItemId]bool, itemRankMap map[module.ItemId][]string) *snakeItemIterator {
	iter := &snakeItemIterator{
		items:               make([]*module.Item, 0, config.Count),
		scoreMap:            make(map[module.ItemId]float64, config.Count),
		index:               0,
		recallName:          config.RecallName,
		config:              config,
		alreadyExistItemMap: alreadyExistItemMap,
		itemRankMap:         itemRankMap,
	}
	return iter
}
func (s *snakeItemIterator) AddItem(item *module.Item) {
	s.items = append(s.items, item)
	if item.RetrieveId != s.recallName {
		s.scoreMap[item.Id] = item.RecallScores[s.recallName]
	} else {
		s.scoreMap[item.Id] = item.Score
	}
}

// Sort items by score
func (s *snakeItemIterator) Sort() {
	sort.Slice(s.items, func(i, j int) bool {
		return s.scoreMap[s.items[i].Id] > s.scoreMap[s.items[j].Id]
	})
}
func (s *snakeItemIterator) Next(size int) (ret []*module.Item) {

	i := 0
	for (i < size) && (s.index < len(s.items)) {
		item := s.items[s.index]
		if _, ok := s.alreadyExistItemMap[item.Id]; !ok {
			s.alreadyExistItemMap[item.Id] = true
			if item.RetrieveId != s.recallName {
				item.RetrieveId = s.recallName
				if score, ok := item.RecallScores[s.recallName]; ok {
					item.Score = score
				}
			}
			s.itemRankMap[item.Id] = append(s.itemRankMap[item.Id], fmt.Sprintf("%s:%d:%f", item.RetrieveId, s.index, item.Score))
			s.index++
			i++
			ret = append(ret, item)
		} else {
			if item.RetrieveId != s.recallName {
				if score, ok := item.RecallScores[s.recallName]; ok {
					s.itemRankMap[item.Id] = append(s.itemRankMap[item.Id], fmt.Sprintf("%s:%d:%f", s.recallName, s.index, score))
				}
			} else {
				s.itemRankMap[item.Id] = append(s.itemRankMap[item.Id], fmt.Sprintf("%s:%d:%f", item.RetrieveId, s.index, item.Score))
			}
			s.index++
			if s.config.Type == Snake_Type_Skip {
				i++
			}
		}
	}

	return
}

type SnakeFilter struct {
	name           string
	configs        []*snakeAdjustCountConfig
	cloneInstances sync.Map
	retainNum      int
}

func NewSnakeFilter(config recconf.FilterConfig) *SnakeFilter {
	filter := SnakeFilter{
		name:      config.Name,
		retainNum: config.RetainNum,
	}
	totalWeight := 0
	for i, conf := range config.AdjustCountConfs {
		snakeConfig, err := newSnakeAdjustCountConfig(conf)
		if err != nil {
			panic(err)
		}
		snakeConfig.Id = i
		snakeConfig.Type = Snake_Type_Refill // default REFILL_ON_DUPLICATE
		if config.SnakeType == "SKIP_ON_DUPLICATE" {
			snakeConfig.Type = Snake_Type_Skip
		}
		filter.configs = append(filter.configs, snakeConfig)
		totalWeight += snakeConfig.Weight
	}

	totalNum := 0
	for _, config := range filter.configs {
		config.Count = int(float64(config.Weight) / float64(totalWeight) * float64(filter.retainNum))
		totalNum += config.Count
	}
	for totalNum > filter.retainNum {
		for _, config := range filter.configs {
			config.Count--
			totalNum--
			if totalNum == filter.retainNum {
				break
			}
		}
	}

	for totalNum < filter.retainNum {
		for _, config := range filter.configs {
			config.Count++
			totalNum++
			if totalNum == filter.retainNum {
				break
			}
		}
	}

	return &filter
}
func (f *SnakeFilter) Filter(filterData *FilterData) error {
	if _, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")
	}

	return f.doFilter(filterData)
}

func (f *SnakeFilter) doFilter(filterData *FilterData) error {
	start := time.Now()
	items := filterData.Data.([]*module.Item)
	newItems := make([]*module.Item, 0, f.retainNum)
	snakeItemIteratorMap := make(map[string]*snakeItemIterator, len(f.configs))
	snakeItemIterators := make([]*snakeItemIterator, 0, len(f.configs))
	alreadyExistItemMap := make(map[module.ItemId]bool, f.retainNum)
	itemRankMap := make(map[module.ItemId][]string, f.retainNum)
	for _, config := range f.configs {
		iter := newSnakeItemIterator(config, alreadyExistItemMap, itemRankMap)
		snakeItemIteratorMap[config.RecallName] = iter
		snakeItemIterators = append(snakeItemIterators, iter)
	}

	for _, item := range items {
		if len(item.RecallScores) > 1 { // item that from multi recall
			currItem := item
			if iter, ok := snakeItemIteratorMap[currItem.RetrieveId]; ok {
				iter.AddItem(currItem)
			}
			for recallName := range item.RecallScores {
				if recallName == currItem.RetrieveId {
					continue
				}
				if iter, ok := snakeItemIteratorMap[recallName]; ok {
					iter.AddItem(currItem) // currItme's RetrieveId is not recallName
				}
			}
		} else {
			if iter, ok := snakeItemIteratorMap[item.RetrieveId]; ok {
				iter.AddItem(item)
			}
		}
	}

	for _, iter := range snakeItemIteratorMap {
		iter.Sort()
	}

	size := 0
	for size < f.retainNum {
		iterSize := 0
		for i, config := range f.configs {
			iter := snakeItemIterators[i]
			items := iter.Next(config.Weight)
			if len(items) > 0 {
				iterSize += len(items)
				newItems = append(newItems, items...)
			}
		}
		if iterSize == 0 {
			break
		}
		size += iterSize
	}

	if len(newItems) > f.retainNum {
		newItems = newItems[:f.retainNum]
	}
	for _, item := range newItems {
		if rankStrings, ok := itemRankMap[item.Id]; ok {
			item.AddProperty("snake_filter", strings.Join(rankStrings, ","))
		}
	}

	filterData.Data = newItems
	filterInfoLog(filterData, "SnakeFilter", f.name, len(newItems), start)
	return nil
}
