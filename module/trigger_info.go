package module

import (
	"database/sql"
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
