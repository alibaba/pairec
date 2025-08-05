package sort

import (
	"fmt"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

type DiversityExclusionRule struct {
	exclusionRuleConfig recconf.ExclusionRuleConfig
	DimensionItemMap    map[module.ItemId]bool
	positions           map[int]bool
	filterParam         *module.FilterParam
	userFeatures        map[string]any
}

func NewDiversityExclusionRule(config recconf.ExclusionRuleConfig, userFeatures map[string]any, size int) *DiversityExclusionRule {
	rule := DiversityExclusionRule{
		exclusionRuleConfig: config,
		DimensionItemMap:    make(map[module.ItemId]bool, size),
		positions:           make(map[int]bool, len(config.Positions)),
		userFeatures:        userFeatures,
	}
	for _, position := range config.Positions {
		rule.positions[position] = true
	}
	if len(config.Conditions) > 0 {
		filterParam := module.NewFilterParamWithConfig(config.Conditions)
		rule.filterParam = filterParam
	}

	return &rule
}

func (r *DiversityExclusionRule) Match(position int, item *module.Item) bool {
	if _, ok := r.positions[position]; !ok {
		return false
	}
	if r.filterParam == nil {
		return false
	}
	if flag, ok := r.DimensionItemMap[item.Id]; ok {
		return flag
	}

	flag, err := r.filterParam.EvaluateByDomain(r.userFeatures, item.GetFeatures())
	if err != nil {
		log.Error(fmt.Sprintf("DiversityExclusionRule.Match err: %v", err))
		r.DimensionItemMap[item.Id] = false
		return false
	}
	r.DimensionItemMap[item.Id] = flag
	return flag
}
