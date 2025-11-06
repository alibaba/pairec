package module

import (
	"strconv"
	"strings"

	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

const (
	TIRRGER_SPLIT = "\u001E"
)

type TriggerItem struct {
	Key          string
	DefaultValue string
	Boundaries   []int
}

func (tr *TriggerItem) GetValue(feature interface{}) string {
	if len(tr.Boundaries) == 0 {
		switch fval := feature.(type) {
		case []any:
			strs := make([]string, 0, len(fval))
			for _, f := range fval {
				strs = append(strs, utils.ToString(f, ""))
			}
			return strings.Join(strs, TIRRGER_SPLIT)
		case []string:
			return strings.Join(fval, TIRRGER_SPLIT)
		case []int:
			strs := make([]string, 0, len(fval))
			for _, f := range fval {
				strs = append(strs, utils.ToString(f, ""))
			}
			return strings.Join(strs, TIRRGER_SPLIT)
		case []int64:
			strs := make([]string, 0, len(fval))
			for _, f := range fval {
				strs = append(strs, utils.ToString(f, ""))
			}
			return strings.Join(strs, TIRRGER_SPLIT)
		default:
			return utils.ToString(feature, tr.DefaultValue)
		}
	}

	val := utils.ToInt(feature, 0)

	index := -1
	for i, boundary := range tr.Boundaries {
		if val <= boundary {
			break
		} else {
			index = i
		}
	}

	if index == -1 {
		return "<=" + strconv.Itoa(tr.Boundaries[0])
	} else if index == len(tr.Boundaries)-1 {
		return ">" + strconv.Itoa(tr.Boundaries[len(tr.Boundaries)-1])
	}

	return strconv.Itoa(tr.Boundaries[index]) + "-" + strconv.Itoa(tr.Boundaries[index+1])
}

type Trigger struct {
	triggers []*TriggerItem
}

func NewTrigger(triggers []recconf.TriggerConfig) *Trigger {

	t := &Trigger{}
	for _, trigger := range triggers {
		triggerItem := &TriggerItem{
			Key:          trigger.TriggerKey,
			DefaultValue: trigger.DefaultValue,
			Boundaries:   trigger.Boundaries,
		}

		if triggerItem.DefaultValue == "" {
			triggerItem.DefaultValue = "NULL"
		}

		t.triggers = append(t.triggers, triggerItem)
	}

	return t
}

func (t *Trigger) GetValue(features map[string]interface{}) string {

	values := make([]string, 0, len(t.triggers))
	for _, trigger := range t.triggers {
		if val, ok := features[trigger.Key]; ok {
			values = append(values, trigger.GetValue(val))
		} else {
			values = append(values, trigger.DefaultValue)
		}
	}

	return strings.Join(values, "_")
}

func ParseTriggerId(triggerId string) []any {
	if !strings.ContainsAny(triggerId, TIRRGER_SPLIT) {
		return []any{triggerId}
	}
	multiTriggerValues := strings.Split(triggerId, "_")
	multiTriggerList := make([][]string, 0, len(multiTriggerValues))
	for _, multiTrigger := range multiTriggerValues {
		multiTriggerList = append(multiTriggerList, strings.Split(multiTrigger, TIRRGER_SPLIT))
	}

	triggers := make([]any, 0, len(multiTriggerList))
	var backtrack func(index int, currentCombination string)
	backtrack = func(index int, currentCombination string) {
		if index == len(multiTriggerList) {
			triggers = append(triggers, currentCombination)
			return
		}

		for _, item := range multiTriggerList[index] {
			if currentCombination == "" {
				backtrack(index+1, item)
			} else {
				backtrack(index+1, currentCombination+"_"+item)
			}
		}
	}
	backtrack(0, "")

	return triggers
}
