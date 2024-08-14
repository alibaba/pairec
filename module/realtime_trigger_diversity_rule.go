package module

import (
	"strings"

	"github.com/alibaba/pairec/v2/recconf"
)

type TriggerDiversityRule struct {
	DiversityRuleConfig recconf.TriggerDiversityRuleConfig
	DimensionItemMap    map[string]string
	DimensionValueSize  map[string]int
}

func NewTriggerDiversityRule(config recconf.TriggerDiversityRuleConfig) *TriggerDiversityRule {
	rule := TriggerDiversityRule{
		DiversityRuleConfig: config,
		DimensionItemMap:    make(map[string]string),
		DimensionValueSize:  make(map[string]int),
	}

	return &rule
}

func (r *TriggerDiversityRule) GetDimensionValue(trigger *TriggerInfo, propertyFieldMap map[string]int) string {
	if value, ok := r.DimensionItemMap[trigger.ItemId]; ok {
		return value
	}

	var dimensionValues []string
	for _, dimension := range r.DiversityRuleConfig.Dimensions {
		value := trigger.StringProperty(dimension, propertyFieldMap)
		dimensionValues = append(dimensionValues, value)
	}

	r.DimensionItemMap[trigger.ItemId] = strings.Join(dimensionValues, "_")

	return r.DimensionItemMap[trigger.ItemId]
}

func (r *TriggerDiversityRule) Match(trigger *TriggerInfo, propertyFieldMap map[string]int, triggerList []*TriggerInfo) bool {
	itemDimensionValue := r.GetDimensionValue(trigger, propertyFieldMap)
	if size, exist := r.DimensionValueSize[itemDimensionValue]; exist {
		//r.DimensionValueSize[itemDimensionValue] = size + 1
		return size < r.DiversityRuleConfig.Size

	} else {
		//r.DimensionValueSize[itemDimensionValue] = 1
		return true
	}
}
func (r *TriggerDiversityRule) AddDimensionValue(trigger *TriggerInfo, propertyFieldMap map[string]int) {
	itemDimensionValue := r.GetDimensionValue(trigger, propertyFieldMap)
	r.DimensionValueSize[itemDimensionValue]++
}
