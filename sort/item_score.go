package sort

import (
	"errors"
	gosort "sort"

	"github.com/alibaba/pairec/v2/module"
)

type ItemScoreSlice []*module.Item

func (us ItemScoreSlice) Len() int {
	return len(us)
}
func (us ItemScoreSlice) Less(i, j int) bool {

	return us[i].Score < us[j].Score
}
func (us ItemScoreSlice) Swap(i, j int) {
	tmp := us[i]
	us[i] = us[j]
	us[j] = tmp
}

type ItemScoreSort struct {
}

func (s *ItemScoreSort) Sort(sortData *SortData) error {
	if _, ok := sortData.Data.([]*module.Item); !ok {
		return errors.New("sort data type error")
	}

	return s.doSort(sortData)
}

func (s *ItemScoreSort) doSort(sortData *SortData) error {
	items := sortData.Data.([]*module.Item)

	gosort.Sort(ItemScoreSlice(items))
	sortData.Data = items
	return nil
}
