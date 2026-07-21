package recallenginerecall

import (
	"fmt"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

type U2IRealtimeTrigger struct {
	*U2IBaseTrigger
	user2ItemDao module.RealTimeUser2ItemDao

	// item property(x) aggregation, enabled when xKey is not empty
	xKey             string
	xDelimiter       string
	xCount           int
	item2XDao        module.Item2XDao // mode 2, joins property values by item ids
	propertyFieldMap map[string]int   // mode 1, index of xKey in PropertyFields
}

func NewU2IRealtimeTrigger(config *recconf.UserTriggerDaoConfig, rulesConfig *recconf.UserTriggerRulesConfig, item2XConf *recconf.Item2XConfig) *U2IRealtimeTrigger {
	conf := recconf.RecallConfig{
		RealTimeUser2ItemDaoConf: recconf.RealTimeUser2ItemDaoConfig{
			UserTriggerDaoConf: *config,
		},
	}
	trigger := &U2IRealtimeTrigger{
		user2ItemDao:   module.NewRealTimeUser2ItemDao(conf),
		U2IBaseTrigger: NewU2IBaseTrigger(rulesConfig),
	}

	if item2XConf != nil && item2XConf.XKey != "" {
		trigger.xKey = item2XConf.XKey
		trigger.xDelimiter = item2XConf.XDelimiter
		trigger.xCount = item2XConf.XCount
		if item2XConf.Item2XTable != "" {
			// mode 2: join property values from the item attribute table
			trigger.item2XDao = module.NewItem2XDao(*item2XConf)
		} else {
			// mode 1: property values come from the behavior table PropertyFields
			trigger.propertyFieldMap = make(map[string]int, len(config.PropertyFields))
			for i, field := range config.PropertyFields {
				trigger.propertyFieldMap[field] = i
			}
			if _, exist := trigger.propertyFieldMap[trigger.xKey]; !exist {
				panic(fmt.Sprintf("Item2XConf.XKey(%s) not found in UserTriggerDaoConf.PropertyFields", trigger.xKey))
			}
		}
	}

	return trigger
}
func (t *U2IRealtimeTrigger) GetTriggerKey(user *module.User, context *context.RecommendContext) *TriggerResult {
	triggerInfos := t.user2ItemDao.GetTriggerInfos(user, context)

	if t.xKey == "" {
		return t.CreateTriggerResult(triggerInfos)
	}

	// aggregate item triggers to property(x) triggers
	var xTriggerInfos []*module.TriggerInfo
	if t.item2XDao != nil {
		itemIds := make([]string, 0, len(triggerInfos))
		for _, info := range triggerInfos {
			itemIds = append(itemIds, info.ItemId)
		}
		item2XMap := t.item2XDao.ListItem2X(itemIds, context)
		xTriggerInfos = module.AggregateItem2XWeights(triggerInfos, item2XMap, t.xDelimiter)
	} else {
		xTriggerInfos = module.AggregateTriggerPropertyWeights(triggerInfos, t.xKey, t.propertyFieldMap, t.xDelimiter)
	}

	if t.xCount > 0 && len(xTriggerInfos) > t.xCount {
		xTriggerInfos = xTriggerInfos[:t.xCount]
	}

	return t.CreateTriggerResult(xTriggerInfos)
}
