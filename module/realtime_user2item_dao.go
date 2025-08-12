package module

import (
	"strconv"
	"strings"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/recconf"
)

type RealTimeUser2ItemDao interface {
	ListItemsByUser(user *User, context *context.RecommendContext) []*Item
	GetTriggers(user *User, context *context.RecommendContext) map[string]float64
	GetTriggerInfos(user *User, context *context.RecommendContext) []*TriggerInfo
}

func NewRealTimeUser2ItemDao(config recconf.RecallConfig) RealTimeUser2ItemDao {

	if config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres {
		if config.RealTimeUser2ItemDaoConf.Item2XTable != "" && config.RealTimeUser2ItemDaoConf.X2ItemTable != "" {
			return NewRealtimeUser2Item2X2ItemHologresDao(config)
		}

		return NewRealtimeUser2ItemHologresDao(config)
	} else if config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.AdapterType == recconf.DataSource_Type_BE {
		return NewRealtimeUser2ItemBeDao(config)
	} else if config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.AdapterType == recconf.DataSource_Type_FeatureStore {
		return NewRealtimeUser2ItemFeatureStoreDao(config)
	} else {
		panic("RealTimeUser2ItemDao not implement")
	}
}

type ItemScoreSlice []*Item

func (us ItemScoreSlice) Len() int {
	return len(us)
}
func (us ItemScoreSlice) Less(i, j int) bool {

	return us[i].Score < us[j].Score
}
func (us ItemScoreSlice) Swap(i, j int) {
	tmp := us[i]
	us[i] = us[j]
	us[j] = tmp
}

func uniqItems(items []*Item) (ret []*Item) {
	uniq := make(map[ItemId]bool, len(items))

	for _, item := range items {
		if _, ok := uniq[item.Id]; !ok {
			uniq[item.Id] = true
			ret = append(ret, item)
		}
	}

	return
}

type RealtimeUser2ItemBaseDao struct {
	recallCount      int
	triggerCount     int
	limit            int
	recallName       string
	propertyFields   []string
	propertyFieldMap map[string]int
	diversityRules   []recconf.TriggerDiversityRuleConfig
	eventPlayTimeMap map[string]float64
	eventWeightMap   map[string]float64
	mergeMode        string
}

func NewRealtimeUser2ItemBaseDao(config *recconf.RecallConfig) *RealtimeUser2ItemBaseDao {
	dao := &RealtimeUser2ItemBaseDao{
		recallName:       config.Name,
		recallCount:      config.RecallCount,
		diversityRules:   config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.DiversityRules,
		propertyFields:   config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.PropertyFields,
		propertyFieldMap: make(map[string]int, len(config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.PropertyFields)),
		triggerCount:     config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.TriggerCount,
		limit:            config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.Limit,
		eventPlayTimeMap: make(map[string]float64),
		eventWeightMap:   make(map[string]float64),
		mergeMode:        config.RealTimeUser2ItemDaoConf.MergeMode,
	}

	if dao.triggerCount == 0 {
		dao.triggerCount = dao.limit
	}

	if config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.EventPlayTime != "" {
		playTimes := strings.Split(config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.EventPlayTime, ";")
		for _, eventTime := range playTimes {
			strs := strings.Split(eventTime, ":")
			if len(strs) == 2 {
				if t, err := strconv.ParseFloat(strs[1], 64); err == nil {
					dao.eventPlayTimeMap[strs[0]] = t
				}
			}
		}
	}
	if config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.EventWeight != "" {
		weights := strings.Split(config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.EventWeight, ";")
		for _, weight := range weights {
			strs := strings.Split(weight, ":")
			if len(strs) == 2 {
				if t, err := strconv.ParseFloat(strs[1], 64); err == nil {
					dao.eventWeightMap[strs[0]] = t
				}
			}
		}
	}

	if len(dao.propertyFields) > 0 {
		for i, field := range dao.propertyFields {
			dao.propertyFieldMap[field] = i
		}
	}
	return dao
}

func (d *RealtimeUser2ItemBaseDao) DiversityTriggers(triggers []*TriggerInfo) []*TriggerInfo {
	if len(d.diversityRules) == 0 {
		return triggers
	}
	length := len(triggers)
	if length == 0 {
		return triggers
	}

	var diversityRules []*TriggerDiversityRule
	for _, config := range d.diversityRules {
		rule := NewTriggerDiversityRule(config)

		diversityRules = append(diversityRules, rule)
	}

	diversitySize := d.triggerCount
	if diversitySize > length {
		diversitySize = length
	}
	var triggerResult []*TriggerInfo
	alreadyMatch := make(map[string]bool, diversitySize)
	//alreadyMatch[triggers[0].itemId] = true
	//triggerResult = append(triggerResult, triggers[0])
	//triggers = triggers[1:]

	index := 0
	for len(triggerResult) <= diversitySize {
		if index == length {
			break
		}

		flag := true
		// if all the rest items not match diversity rule, use the first item append to the result
		firstIndex := -1
		for i, trigger := range triggers {
			if _, ok := alreadyMatch[trigger.ItemId]; ok {
				continue
			}

			if firstIndex == -1 {
				firstIndex = i
			}
			flag = true
			for _, rule := range diversityRules {
				if flag = rule.Match(trigger, d.propertyFieldMap, triggerResult); !flag {
					break
				}
			}

			// if the item match all the diversity rule, so add it to the result
			if flag {
				alreadyMatch[trigger.ItemId] = true
				triggerResult = append(triggerResult, trigger)
				index++
				for _, rule := range diversityRules {
					rule.AddDimensionValue(trigger, d.propertyFieldMap)
				}
				break
			}
		}

		if !flag {
			alreadyMatch[triggers[firstIndex].ItemId] = true
			triggerResult = append(triggerResult, triggers[firstIndex])
			index++
		}
	}

	return triggerResult
}

func (d *RealtimeUser2ItemBaseDao) isSnakeMergeMode() bool {
	return d.mergeMode == "snake"
}
