package sort

import (
	"errors"
	"sort"

	"github.com/alibaba/pairec/module"
)

type ItemRankScoreSort struct {
}

func NewItemRankScoreSort() *ItemRankScoreSort {
	reverseSort := &ItemRankScoreSort{}
	return reverseSort
}

func (s *ItemRankScoreSort) Sort(sortData *SortData) error {
	if _, ok := sortData.Data.([]*module.Item); !ok {
		return errors.New("sort data type error")

	}
	return s.doSort(sortData)
}

func (s *ItemRankScoreSort) doSort(sortData *SortData) error {
	// scene := sortData.Context.GetParameter("scene").(string)
	items := sortData.Data.([]*module.Item)
	sort.Sort(sort.Reverse(ItemScoreSlice(items)))
	sortData.Data = items
	return nil
}

func init() {
	RegisterSort("ItemRankScore", NewItemRankScoreSort())
}
