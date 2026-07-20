package sort

import (
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

// DiversityRuleV2Interface extends the V1 match semantics with a history
// prefix, so that diversity rules can take effect across pages. The history
// items are virtually placed before the result list when matching.
type DiversityRuleV2Interface interface {
	Match(item *module.Item, itemList []*module.Item) bool
	GetWeight() int
	SetHistory(prevItems []*module.Item) // warm up window with prev page tail items
	HistoryLen() int                     // 0 means no history
}

// historyMaxLen returns the max count of history items that can influence
// the match result of the given rule config.
func historyMaxLen(config recconf.DiversityRuleConfig) int {
	maxLen := 0
	if config.WindowSize > 0 && config.FrequencySize > 0 && config.WindowSize > config.FrequencySize {
		maxLen = config.WindowSize - 1
	}
	if config.IntervalSize > maxLen {
		maxLen = config.IntervalSize
	}
	return maxLen
}

var _ DiversityRuleV2Interface = (*DiversityRuleV2)(nil)

// DiversityRuleV2 embeds the V1 rule to reuse dimension value evaluation and
// its cache, and only overrides Match to scan across the history prefix.
type DiversityRuleV2 struct {
	*DiversityRule
	historyValues []string // prev page tail dimension values, oldest first
}

func NewDiversityRuleV2(config recconf.DiversityRuleConfig, size int) *DiversityRuleV2 {
	return &DiversityRuleV2{
		DiversityRule: NewDiversityRule(config, size),
	}
}

func (r *DiversityRuleV2) SetHistory(prevItems []*module.Item) {
	maxLen := historyMaxLen(r.DiversityRuleConfig)
	if maxLen <= 0 || len(prevItems) == 0 {
		return
	}
	if len(prevItems) > maxLen {
		prevItems = prevItems[len(prevItems)-maxLen:]
	}
	r.historyValues = make([]string, 0, len(prevItems))
	for _, item := range prevItems {
		r.historyValues = append(r.historyValues, r.GetDimensionValue(item))
	}
}

func (r *DiversityRuleV2) HistoryLen() int {
	return len(r.historyValues)
}

// valueAt returns the dimension value at virtual position i of the sequence
// historyValues ++ itemList.
func (r *DiversityRuleV2) valueAt(i int, itemList []*module.Item) string {
	if i < len(r.historyValues) {
		return r.historyValues[i]
	}
	return r.GetDimensionValue(itemList[i-len(r.historyValues)])
}

// Match is identical to the V1 DiversityRule.Match when historyValues is
// empty; otherwise the history prefix is included in the scan window.
func (r *DiversityRuleV2) Match(item *module.Item, itemList []*module.Item) bool {
	size := len(itemList) + len(r.historyValues)

	itemDimensionValue := r.GetDimensionValue(item)
	if r.DiversityRuleConfig.IntervalSize > 0 && size >= r.DiversityRuleConfig.IntervalSize {
		end := size
		begin := size - r.DiversityRuleConfig.IntervalSize
		sameValue := 1
		for i := end - 1; i >= begin; i-- {
			if itemDimensionValue == r.valueAt(i, itemList) {
				sameValue++
			} else {
				break
			}
		}

		if sameValue > r.DiversityRuleConfig.IntervalSize {
			return false
		}

	}
	if r.DiversityRuleConfig.WindowSize > 0 &&
		r.DiversityRuleConfig.FrequencySize > 0 &&
		r.DiversityRuleConfig.WindowSize > r.DiversityRuleConfig.FrequencySize {
		end := size
		begin := size - r.DiversityRuleConfig.WindowSize + 1
		if begin < 0 {
			begin = 0
		}

		sameValue := 1
		for i := begin; i < end; i++ {
			if itemDimensionValue == r.valueAt(i, itemList) {
				sameValue++
			}

			if sameValue > r.DiversityRuleConfig.FrequencySize {
				return false
			}
		}
	}
	return true
}

var _ DiversityRuleV2Interface = (*DiversityRuleMultiDimensionV2)(nil)

// DiversityRuleMultiDimensionV2 is the multi value dimension variant, it
// reuses the V1 dimension evaluation (delimiter split) and equality check.
type DiversityRuleMultiDimensionV2 struct {
	*DiversityRuleMultiDimension
	historyValues [][]any // prev page tail dimension values, oldest first
}

func NewDiversityRuleMultiDimensionV2(config recconf.DiversityRuleConfig, size int, multiDimensionMap map[int]recconf.MultiValueDimensionConfig) *DiversityRuleMultiDimensionV2 {
	return &DiversityRuleMultiDimensionV2{
		DiversityRuleMultiDimension: NewDiversityRuleMultiDimension(config, size, multiDimensionMap),
	}
}

func (r *DiversityRuleMultiDimensionV2) SetHistory(prevItems []*module.Item) {
	maxLen := historyMaxLen(r.DiversityRuleConfig)
	if maxLen <= 0 || len(prevItems) == 0 {
		return
	}
	if len(prevItems) > maxLen {
		prevItems = prevItems[len(prevItems)-maxLen:]
	}
	r.historyValues = make([][]any, 0, len(prevItems))
	for _, item := range prevItems {
		r.historyValues = append(r.historyValues, r.GetDimensionValue(item))
	}
}

func (r *DiversityRuleMultiDimensionV2) HistoryLen() int {
	return len(r.historyValues)
}

func (r *DiversityRuleMultiDimensionV2) valueAt(i int, itemList []*module.Item) []any {
	if i < len(r.historyValues) {
		return r.historyValues[i]
	}
	return r.GetDimensionValue(itemList[i-len(r.historyValues)])
}

func (r *DiversityRuleMultiDimensionV2) Match(item *module.Item, itemList []*module.Item) bool {
	size := len(itemList) + len(r.historyValues)

	itemDimensionValues := r.GetDimensionValue(item)
	if r.DiversityRuleConfig.IntervalSize > 0 && size >= r.DiversityRuleConfig.IntervalSize {
		end := size
		begin := size - r.DiversityRuleConfig.IntervalSize
		sameValue := 1
		for i := end - 1; i >= begin; i-- {
			if r.isDimensionValuesEqual(itemDimensionValues, r.valueAt(i, itemList)) {
				sameValue++
			} else {
				break
			}
		}

		if sameValue > r.DiversityRuleConfig.IntervalSize {
			return false
		}

	}
	if r.DiversityRuleConfig.WindowSize > 0 &&
		r.DiversityRuleConfig.FrequencySize > 0 &&
		r.DiversityRuleConfig.WindowSize > r.DiversityRuleConfig.FrequencySize {
		end := size
		begin := size - r.DiversityRuleConfig.WindowSize + 1
		if begin < 0 {
			begin = 0
		}

		sameValue := 1
		for i := begin; i < end; i++ {
			if r.isDimensionValuesEqual(itemDimensionValues, r.valueAt(i, itemList)) {
				sameValue++
			}

			if sameValue > r.DiversityRuleConfig.FrequencySize {
				return false
			}
		}
	}
	return true
}
