package filter

import (
	"errors"
	"time"

	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

// user exposure history filter
type ItemCustomFilter struct {
	itemCustomDao module.ItemCustomFilterDao
}

func NewItemCustomFilter(config recconf.FilterConfig) *ItemCustomFilter {
	filter := ItemCustomFilter{
		itemCustomDao: module.NewItemCustomFilterDao(config),
	}

	return &filter
}
func (f *ItemCustomFilter) Filter(filterData *FilterData) error {
	if _, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")

	}
	return f.doFilter(filterData)
}

func (f *ItemCustomFilter) doFilter(filterData *FilterData) error {
	start := time.Now()
	items := filterData.Data.([]*module.Item)

	filterIds := f.itemCustomDao.GetFilterItems()

	newItems := make([]*module.Item, 0, len(items))
	for _, item := range items {
		if _, ok := filterIds[item.Id]; !ok {
			newItems = append(newItems, item)
		}
	}

	filterData.Data = newItems
	filterInfoLog(filterData, "ItemCustomFilter", len(newItems), start)
	return nil
}
