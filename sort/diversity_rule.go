package sort

import (
	"strings"

	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type DiversityMatchFunc func(item *module.Item) bool

type DiversityRuleInterface interface {
	Match(item *module.Item, itemList []*module.Item) bool
	GetWeight() int
}

var _ DiversityRuleInterface = (*DiversityRule)(nil)

type DiversityRule struct {
	DiversityRuleConfig recconf.DiversityRuleConfig
	DimensionItemMap    map[module.ItemId]string
}

func NewDiversityRule(config recconf.DiversityRuleConfig, size int) *DiversityRule {
	rule := DiversityRule{
		DiversityRuleConfig: config,
		DimensionItemMap:    make(map[module.ItemId]string, size),
	}

	return &rule
}

func (r *DiversityRule) GetDimensionValue(item *module.Item) string {
	if value, ok := r.DimensionItemMap[item.Id]; ok {
		return value
	}

	var dimensionValues []string
	for _, dimension := range r.DiversityRuleConfig.Dimensions {
		value := item.StringProperty(dimension)
		dimensionValues = append(dimensionValues, value)
	}

	r.DimensionItemMap[item.Id] = strings.Join(dimensionValues, "_")

	return r.DimensionItemMap[item.Id]
}

func (r *DiversityRule) Match(item *module.Item, itemList []*module.Item) bool {
	size := len(itemList)

	itemDimensionValue := r.GetDimensionValue(item)
	if r.DiversityRuleConfig.IntervalSize > 0 && size >= r.DiversityRuleConfig.IntervalSize {
		end := size
		begin := size - r.DiversityRuleConfig.IntervalSize
		sameValue := 1
		for i := end - 1; i >= begin; i-- {
			if itemDimensionValue == r.GetDimensionValue(itemList[i]) {
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
			if itemDimensionValue == r.GetDimensionValue(itemList[i]) {
				sameValue++
			}

			if sameValue > r.DiversityRuleConfig.FrequencySize {
				return false
			}
		}
	}
	return true
}

func (r *DiversityRule) GetWeight() int {
	return r.DiversityRuleConfig.Weight

}

type DiversityRuleMultiDimension struct {
	DiversityRuleConfig recconf.DiversityRuleConfig
	DimensionItemMap    map[module.ItemId][]any
	multiDimensionMap   map[int]recconf.MultiValueDimensionConfig
}

func NewDiversityRuleMultiDimension(config recconf.DiversityRuleConfig, size int, multiDimensionMap map[int]recconf.MultiValueDimensionConfig) *DiversityRuleMultiDimension {
	rule := DiversityRuleMultiDimension{
		DiversityRuleConfig: config,
		DimensionItemMap:    make(map[module.ItemId][]any, size),
		multiDimensionMap:   multiDimensionMap,
	}

	return &rule
}

func (r *DiversityRuleMultiDimension) GetDimensionValue(item *module.Item) []any {
	if value, ok := r.DimensionItemMap[item.Id]; ok {
		return value
	}

	var dimensionValues []any
	for i, dimension := range r.DiversityRuleConfig.Dimensions {
		multiDimensionConf, ok := r.multiDimensionMap[i]
		if ok {
			value := item.StringProperty(dimension)
			strs := strings.Split(value, multiDimensionConf.Delimiter)
			dimensionValues = append(dimensionValues, strs)
		} else {
			value := item.StringProperty(dimension)
			dimensionValues = append(dimensionValues, value)
		}
	}

	r.DimensionItemMap[item.Id] = dimensionValues

	return r.DimensionItemMap[item.Id]
}

func (r *DiversityRuleMultiDimension) Match(item *module.Item, itemList []*module.Item) bool {
	size := len(itemList)

	itemDimensionValues := r.GetDimensionValue(item)
	if r.DiversityRuleConfig.IntervalSize > 0 && size >= r.DiversityRuleConfig.IntervalSize {
		end := size
		begin := size - r.DiversityRuleConfig.IntervalSize
		sameValue := 1
		for i := end - 1; i >= begin; i-- {
			if r.isDimensionValuesEqual(itemDimensionValues, r.GetDimensionValue(itemList[i])) {
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
			if r.isDimensionValuesEqual(itemDimensionValues, r.GetDimensionValue(itemList[i])) {
				sameValue++
			}

			if sameValue > r.DiversityRuleConfig.FrequencySize {
				return false
			}
		}
	}
	return true
}

func (r *DiversityRuleMultiDimension) GetWeight() int {
	return r.DiversityRuleConfig.Weight
}
func (r *DiversityRuleMultiDimension) isDimensionValuesEqual(left, right []any) bool {
	if len(left) != len(right) {
		return false
	}
	for i := 0; i < len(left); i++ {
		switch left[i].(type) {
		case string:
			if left[i] != right[i] {
				return false
			}
		case []string:
			leftValues := left[i].([]string)
			if rightValues, ok := right[i].([]string); ok {
				if !utils.StringContains(leftValues, rightValues) {
					return false
				}

			} else {
				return false
			}
		}
	}
	return true
}
