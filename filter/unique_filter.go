package filter

import (
	"errors"

	"github.com/alibaba/pairec/v2/module"
)

// remove duplicate item
type UniqueFilter struct {
}

func NewUniqueFilter() *UniqueFilter {
	filter := UniqueFilter{}

	return &filter
}
func (f *UniqueFilter) Filter(filterData *FilterData) error {
	if _, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")

	}
	return f.doFilter(filterData)
}

func (f *UniqueFilter) doFilter(filterData *FilterData) error {
	items := filterData.Data.([]*module.Item)
	newItems := make([]*module.Item, 0)
	uniq := make(map[module.ItemId]*module.Item, len(items))

	for _, item := range items {
		if exist, ok := uniq[item.Id]; !ok {
			uniq[item.Id] = item
			newItems = append(newItems, item)
		} else {
			algoScores := item.GetAlgoScores()
			for name, score := range algoScores {
				exist.AddAlgoScore(name, score)
			}
		}
	}

	filterData.Data = newItems
	return nil
}

func (f *UniqueFilter) MatchTag(tag string) bool {
	// default filter, so filter all tag
	return true
}
