package filter

import (
	"errors"
	"sort"
	"time"

	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
	psort "github.com/alibaba/pairec/sort"
)

type CompletelyFairCountFilter struct {
	retainNum int
}

func NewCompletelyFairCountFilter(config recconf.FilterConfig) *CompletelyFairCountFilter {
	filter := CompletelyFairCountFilter{
		retainNum: config.RetainNum,
	}

	return &filter
}
func (f *CompletelyFairCountFilter) Filter(filterData *FilterData) error {
	if _, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")

	}
	return f.doFilter(filterData)
}

func (f *CompletelyFairCountFilter) doFilter(filterData *FilterData) error {
	start := time.Now()
	items := filterData.Data.([]*module.Item)
	if len(items) <= f.retainNum {
		return nil
	}

	newItems := make([]*module.Item, 0, 200)
	recallToItemMap := make(map[string][]*module.Item)
	recallNames := make([]string, 0, 10)
	recallNamesMap := make(map[string]bool, 10)

	/**
	// first random
	rand.Shuffle(len(items)/2, func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})
	**/

	sort.Sort(sort.Reverse(psort.ItemScoreSlice(items)))

	for _, item := range items {
		recallToItemMap[item.RetrieveId] = append(recallToItemMap[item.RetrieveId], item)
		if _, ok := recallNamesMap[item.RetrieveId]; !ok {
			recallNamesMap[item.RetrieveId] = true
			recallNames = append(recallNames, item.RetrieveId)
		}
	}

	var (
		count            = 0
		recallNamesCount = len(recallNames)
	)

	for count < f.retainNum {
		i := count % recallNamesCount

		itemList := recallToItemMap[recallNames[i]]

		newItems = append(newItems, itemList[0])
		count++

		if len(itemList) == 1 {
			recallNames[i] = recallNames[recallNamesCount-1]
			recallNames = recallNames[:recallNamesCount-1]
			recallNamesCount--
		} else {
			itemList = itemList[1:]
			recallToItemMap[recallNames[i]] = itemList

		}
	}

	filterData.Data = newItems
	filterInfoLog(filterData, "CompletelyFairCountFilter", len(newItems), start)
	return nil
}
