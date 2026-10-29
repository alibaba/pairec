package filter

import (
	"errors"
	"fmt"
	"time"

	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

const (
	uniqueKeepStrategyMin = "min"
	uniqueKeepStrategyMax = "max"
)

// remove duplicate item
type DimensionFieldUniqueFilter struct {
	name         string
	field        string
	keepProperty string // item property name used for comparison within duplicate group
	keepStrategy string // "min" or "max"; empty means keep the first occurrence
}

func NewDimensionFieldUniqueFilter(config recconf.FilterConfig) *DimensionFieldUniqueFilter {
	keepProperty := config.UniqueKeepConf.Property
	keepStrategy := config.UniqueKeepConf.Strategy
	if keepStrategy != "" && keepStrategy != uniqueKeepStrategyMin && keepStrategy != uniqueKeepStrategyMax {
		panic(fmt.Sprintf("DimensionFieldUniqueFilter:%s invalid UniqueKeepConf.Strategy:%s, must be min or max", config.Name, keepStrategy))
	}
	if (keepStrategy == "") != (keepProperty == "") {
		panic(fmt.Sprintf("DimensionFieldUniqueFilter:%s UniqueKeepConf.Property and UniqueKeepConf.Strategy must be configured together", config.Name))
	}

	filter := DimensionFieldUniqueFilter{
		name:         config.Name,
		field:        config.Dimension,
		keepProperty: keepProperty,
		keepStrategy: keepStrategy,
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
	uniqPos := make(map[string]int, len(items))
	// number of duplicate pairs whose min/max comparison was skipped because the property is unreadable on both
	// items; kept as a local counter so the shared filter instance stays concurrency-safe, and reported once per
	// call to avoid log flooding
	skippedComparisons := 0

	for _, item := range items {
		val := item.StringProperty(f.field)
		if val == "" {
			newItems = append(newItems, item)
			continue
		}

		old, ok := uniq[val]
		if !ok {
			uniq[val] = item
			uniqPos[val] = len(newItems)
			newItems = append(newItems, item)
			continue
		}

		// keep strategy is configured, replace the group winner in place if the new item is better
		if f.keepStrategy != "" {
			better, comparable := f.isBetterKeepItem(old, item)
			if !comparable {
				skippedComparisons++
			} else if better {
				uniq[val] = item
				newItems[uniqPos[val]] = item
			}
		}
	}

	filterData.Data = newItems
	filterInfoLog(filterData, "DimensionFieldUniqueFilter", f.name, len(newItems), start)
	if skippedComparisons > 0 {
		filterData.Context.LogWarning(fmt.Sprintf("module=DimensionFieldUniqueFilter\tname=%s\tevent=KeepPropertyUnreadable\tproperty=%s\tskipped_comparison_count=%d\tboth items lacked a readable property so the min/max comparison was skipped for these duplicate pairs",
			f.name, f.keepProperty, skippedComparisons))
	}
	return nil
}

// isBetterKeepItem reports whether the new item should replace the old one within the same duplicate group.
// The item with a comparable property value wins over the one without; on tie the old (first occurring) item is kept.
// The second result reports whether the comparison property is readable on at least one side; when it is false the
// winner can not be decided by value comparison and the current group winner is kept unchanged.
func (f *DimensionFieldUniqueFilter) isBetterKeepItem(old, item *module.Item) (better bool, comparable bool) {
	oldVal, oldErr := old.FloatProperty(f.keepProperty)
	newVal, newErr := item.FloatProperty(f.keepProperty)
	if oldErr != nil && newErr != nil {
		return false, false
	}
	if oldErr != nil {
		return true, true
	}
	if newErr != nil {
		return false, true
	}

	if f.keepStrategy == uniqueKeepStrategyMax {
		return newVal > oldVal, true
	}
	return newVal < oldVal, true
}

func (f *DimensionFieldUniqueFilter) MatchTag(tag string) bool {
	// default filter, so filter all tag
	return true
}
