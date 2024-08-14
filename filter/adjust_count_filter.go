package filter

import (
	"errors"
	"math/rand"

	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

type AdjustNewItemsFunc func(filterData *FilterData, retainNum int) []*module.Item

var (
	adjustNewItemsFunc = DefaultAdjustNewItemsFunc
)

func DefaultAdjustNewItemsFunc(filterData *FilterData, retainNum int) []*module.Item {
	items := filterData.Data.([]*module.Item)
	newItems := make([]*module.Item, retainNum)

	rand.Shuffle(len(items), func(i, j int) {//打乱数组顺序
		items[i], items[j] = items[j], items[i]
	})

	copy(newItems, items)
	return newItems
}

func RegisterAdjustNewItemsFunc(f AdjustNewItemsFunc) {
	adjustNewItemsFunc = f
}

type AdjustCountFilter struct {
	retainNum   int
	shuffleItem bool
}

func NewAdjustCountFilter(config recconf.FilterConfig) *AdjustCountFilter {
	filter := AdjustCountFilter{
		retainNum:   config.RetainNum,
		shuffleItem: true,
	}

	if config.ShuffleItem == false {
		filter.shuffleItem = false
	}
	return &filter
}
func (f *AdjustCountFilter) Filter(filterData *FilterData) error {
	if _, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")

	}
	return f.doFilter(filterData)
}

func (f *AdjustCountFilter) doFilter(filterData *FilterData) error {
	items := filterData.Data.([]*module.Item)
	if len(items) <= f.retainNum {
		return nil
	}

	if f.shuffleItem {
		filterData.Data = adjustNewItemsFunc(filterData, f.retainNum)
	} else {
		newItems := make([]*module.Item, f.retainNum)
		copy(newItems, items)
		filterData.Data = newItems
	}
	return nil
}
