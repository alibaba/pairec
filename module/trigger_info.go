package module

import (
	"database/sql"
	gosort "sort"
	"strings"
)

type TriggerInfo struct {
	ItemId              string
	event               string
	playTime            float64
	timestamp           int64
	Weight              float64
	propertyFieldValues []sql.NullString
}

func (t *TriggerInfo) StringProperty(dimension string, propertyFieldMap map[string]int) string {

	index, exist := propertyFieldMap[dimension]
	if !exist {
		return ""
	}

	if index < len(t.propertyFieldValues) {
		if value := t.propertyFieldValues[index]; value.Valid {
			return value.String
		}
	}
	return ""

}

type TriggerInfoSlice []*TriggerInfo

func (us TriggerInfoSlice) Len() int {
	return len(us)
}
func (us TriggerInfoSlice) Less(i, j int) bool {

	return us[i].Weight < us[j].Weight
}
func (us TriggerInfoSlice) Swap(i, j int) {
	tmp := us[i]
	us[i] = us[j]
	us[j] = tmp
}

// isTriggerOutOfTimeWindow reports whether the behavior timestamp is out of the recent time window.
// timeWindow <= 0 means no time window filter.
func isTriggerOutOfTimeWindow(timestamp, nowUnix, timeWindow int64) bool {
	return timeWindow > 0 && timestamp < nowUnix-timeWindow
}

// aggregatePropertyWeights splits the property value by delimiter and accumulates
// the trigger weight into each property value, then returns synthesized TriggerInfos
// (ItemId is the property value) sorted by weight in descending order.
func aggregatePropertyWeights(triggerInfos []*TriggerInfo, valueFunc func(*TriggerInfo) string, delimiter string) (ret []*TriggerInfo) {
	xWeightMap := make(map[string]float64, len(triggerInfos))
	for _, info := range triggerInfos {
		xVal := valueFunc(info)
		if xVal == "" || strings.EqualFold(xVal, "null") {
			continue
		}
		if delimiter != "" {
			for _, v := range strings.Split(xVal, delimiter) {
				if v != "" {
					xWeightMap[v] += info.Weight
				}
			}
		} else {
			xWeightMap[xVal] += info.Weight
		}
	}

	for xVal, weight := range xWeightMap {
		ret = append(ret, &TriggerInfo{ItemId: xVal, Weight: weight})
	}
	gosort.Sort(gosort.Reverse(TriggerInfoSlice(ret)))
	return
}

// AggregateTriggerPropertyWeights aggregates item trigger weights into property(x) level triggers,
// the property value comes from the trigger property fields (mode 1).
func AggregateTriggerPropertyWeights(triggerInfos []*TriggerInfo, xKey string, propertyFieldMap map[string]int, delimiter string) []*TriggerInfo {
	return aggregatePropertyWeights(triggerInfos, func(info *TriggerInfo) string {
		return info.StringProperty(xKey, propertyFieldMap)
	}, delimiter)
}

// AggregateItem2XWeights aggregates item trigger weights into property(x) level triggers,
// the property value comes from item2XMap joined by item id (mode 2).
func AggregateItem2XWeights(triggerInfos []*TriggerInfo, item2XMap map[string]string, delimiter string) []*TriggerInfo {
	return aggregatePropertyWeights(triggerInfos, func(info *TriggerInfo) string {
		return item2XMap[info.ItemId]
	}, delimiter)
}
