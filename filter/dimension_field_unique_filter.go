package filter

import (
	"errors"
	"time"

	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
)

// remove duplicate item
type DimensionFieldUniqueFilter struct {
	field string
}

func NewDimensionFieldUniqueFilter(config recconf.FilterConfig) *DimensionFieldUniqueFilter {
	filter := DimensionFieldUniqueFilter{
		field: config.Dimension,
	}

	return &filter
}
func (f *DimensionFieldUniqueFilter) Filter(filterData *FilterData) error {
	if _, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")

	}
	return f.doFilter(filterData)
}

func (f *DimensionFieldUniqueFilter) doFilter(filterData *FilterData) error {
	start := time.Now()
	items := filterData.Data.([]*module.Item)
	newItems := make([]*module.Item, 0)
	uniq := make(map[string]*module.Item, len(items))

	for _, item := range items {
		val := item.StringProperty(f.field)
		if val == "" {
			newItems = append(newItems, item)
		} else {

			if _, ok := uniq[val]; !ok {
				uniq[val] = item
				newItems = append(newItems, item)
			}
		}
	}

	filterData.Data = newItems
	filterInfoLog(filterData, "DimensionFieldUniqueFilter", len(newItems), start)
	return nil
}

func (f *DimensionFieldUniqueFilter) MatchTag(tag string) bool {
	// default filter, so filter all tag
	return true
}
