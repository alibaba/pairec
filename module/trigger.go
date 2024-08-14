package module

import (
	"strconv"
	"strings"

	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type TriggerItem struct {
	Key          string
	DefaultValue string
	Boundaries   []int
}

func (tr *TriggerItem) GetValue(feature interface{}) string {
	if len(tr.Boundaries) == 0 {
		return utils.ToString(feature, tr.DefaultValue)
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
