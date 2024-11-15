package filter

import (
	"errors"
	"time"

	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

type ItemStateFilter struct {
	name         string
	itemStateDao module.ItemStateFilterDao
}

func NewItemStateFilter(config recconf.FilterConfig) *ItemStateFilter {
	filter := ItemStateFilter{
		name:         config.Name,
		itemStateDao: module.NewItemStateFilterDao(config),
	}

	return &filter
}
func (f *ItemStateFilter) Filter(filterData *FilterData) error {
	if _, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")

	}
	return f.doFilter(filterData)
}

func (f *ItemStateFilter) doFilter(filterData *FilterData) error {
	start := time.Now()
	items := filterData.Data.([]*module.Item)

	newItems := f.itemStateDao.Filter(filterData.User, items)

	filterData.Data = newItems
	filterInfoLog(filterData, "ItemStateFilter", f.name, len(newItems), start)
	return nil
}
