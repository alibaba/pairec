package berecall

import (
	"fmt"
	"strings"

	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

type U2IBaseTrigger struct {
	triggerCounts   []int
	defaultValue    int
	triggerCountLen int
}

func NewU2IBaseTrigger(rulesConfig *recconf.UserTriggerRulesConfig) *U2IBaseTrigger {
	trigger := &U2IBaseTrigger{}

	if len(rulesConfig.TriggerCounts) > 0 {
		trigger.triggerCounts = rulesConfig.TriggerCounts
		trigger.defaultValue = rulesConfig.DefaultValue
		if trigger.defaultValue == 0 {
			trigger.defaultValue = trigger.triggerCounts[len(trigger.triggerCounts)-1]
		}
		trigger.triggerCountLen = len(trigger.triggerCounts)
	}

	return trigger
}
func (t *U2IBaseTrigger) CreateTriggerResult(triggerInfos []*module.TriggerInfo) *TriggerResult {
	var itemList []string
	for _, trigger := range triggerInfos {
		itemList = append(itemList, fmt.Sprintf("%s:%f", trigger.ItemId, trigger.Weight))
	}
	triggerResult := &TriggerResult{
		TriggerItem: strings.Join(itemList, ","),
	}

	if t.triggerCountLen > 0 {
		var distinctParams []string
		count := 0
		for index, trigger := range triggerInfos {
			count = t.defaultValue
			if index < t.triggerCountLen {
				count = t.triggerCounts[index]
			}

			distinctParams = append(distinctParams, fmt.Sprintf("%s*%d", trigger.ItemId, count))
		}

		triggerResult.DistinctParam = strings.Join(distinctParams, ",")
	}
	return triggerResult
}
