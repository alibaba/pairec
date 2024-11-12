package filter

import (
	"errors"
	"math/rand"
	"sort"
	"time"

	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	psort "github.com/alibaba/pairec/v2/sort"
)

type PriorityAdjustCountFilterV2 struct {
	name    string
	configs []recconf.AdjustCountConfig
}

func NewPriorityAdjustCountFilterV2(config recconf.FilterConfig) *PriorityAdjustCountFilterV2 {
	filter := PriorityAdjustCountFilterV2{
		name:    config.Name,
		configs: config.AdjustCountConfs,
	}

	return &filter
}
func (f *PriorityAdjustCountFilterV2) Filter(filterData *FilterData) error {
	if _, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")
	}

	return f.doFilter(filterData)
}

func (f *PriorityAdjustCountFilterV2) doFilter(filterData *FilterData) error {
	start := time.Now()
	items := filterData.Data.([]*module.Item)
	newItems := make([]*module.Item, 0, 200)
	recallToItemMap := make(map[string][]*module.Item)
	duplicateRecallItemMap := make(map[*module.Item]bool)
	accumulator := 0

	sortItems := func(items []*module.Item) {
		rand.Shuffle(len(items)/2, func(i, j int) {
			items[i], items[j] = items[j], items[i]
		})
		sort.Sort(sort.Reverse(psort.ItemScoreSlice(items)))
	}

	for _, item := range items {
		if len(item.RecallScores) > 1 { // item that from multi recall
			duplicateRecallItemMap[item] = true
		} else {
			recallToItemMap[item.RetrieveId] = append(recallToItemMap[item.RetrieveId], item)
		}
	}

	for _, config := range f.configs {
		recallItems := recallToItemMap[config.RecallName]

		// append item that can be considered as it from this recall
		for item := range duplicateRecallItemMap {
			if score, ok := item.RecallScores[config.RecallName]; ok {
				item.RetrieveId = config.RecallName
				item.Score = score
				recallItems = append(recallItems, item)
			}
		}

		sortItems(recallItems)

		if config.Type == Fix_Count_Type {
			for i := 0; i < len(recallItems) && i < config.Count; i++ {
				newItems = append(newItems, recallItems[i])

				if duplicateRecallItemMap[recallItems[i]] {
					// already used in current recall, can't be used in other recall
					delete(duplicateRecallItemMap, recallItems[i])
				}
			}
		} else if config.Type == Accumulate_Count_Type {
			count := config.Count - accumulator

			for i := 0; i < len(recallItems) && i < count; i++ {
				newItems = append(newItems, recallItems[i])

				if duplicateRecallItemMap[recallItems[i]] {
					delete(duplicateRecallItemMap, recallItems[i])
				}

				accumulator++
			}
		}
	}

	filterData.Data = newItems
	filterInfoLogV2(filterData, "PriorityAdjustCountFilterV2", f.name, len(newItems), start)
	return nil
}
