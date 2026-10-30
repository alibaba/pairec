package filter

import (
	"errors"
	"fmt"
	"math"
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
	// uniqPos is only used by the keep strategy; allocate lazily to keep the legacy path at single-map overhead
	var uniqPos map[string]int
	if f.keepStrategy != "" {
		uniqPos = make(map[string]int, len(items))
	}
	// local counter of duplicate pairs skipped for unreadable property, reported once per call to avoid log flooding
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
			if f.keepStrategy != "" {
				uniqPos[val] = len(newItems)
			}
			newItems = append(newItems, item)
			continue
		}

		// replace the group winner in place if the new item is better
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

// isBetterKeepItem reports whether the new item should replace the current group winner: the item with a readable
// property wins, ties keep the old one. comparable is false when neither side is readable (missing or NaN), so the
// winner is kept unchanged. NaN counts as unreadable because every min/max comparison against it is false.
func (f *DimensionFieldUniqueFilter) isBetterKeepItem(old, item *module.Item) (better bool, comparable bool) {
	oldVal, oldErr := old.FloatProperty(f.keepProperty)
	newVal, newErr := item.FloatProperty(f.keepProperty)
	oldReadable := oldErr == nil && !math.IsNaN(oldVal)
	newReadable := newErr == nil && !math.IsNaN(newVal)
	if !oldReadable && !newReadable {
		return false, false
	}
	if !oldReadable {
		return true, true
	}
	if !newReadable {
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
