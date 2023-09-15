package berecall

import (
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
)

type U2IRealtimeTrigger struct {
	*U2IBaseTrigger
	user2ItemDao module.RealTimeUser2ItemDao
}

func NewU2IRealtimeTrigger(config *recconf.UserTriggerDaoConfig, rulesConfig *recconf.UserTriggerRulesConfig) *U2IRealtimeTrigger {
	conf := recconf.RecallConfig{
		RealTimeUser2ItemDaoConf: recconf.RealTimeUser2ItemDaoConfig{
			UserTriggerDaoConf: *config,
		},
	}
	trigger := &U2IRealtimeTrigger{
		user2ItemDao:   module.NewRealTimeUser2ItemDao(conf),
		U2IBaseTrigger: NewU2IBaseTrigger(rulesConfig),
	}

	return trigger
}
func (t *U2IRealtimeTrigger) GetTriggerKey(user *module.User, context *context.RecommendContext) *TriggerResult {
	triggerInfos := t.user2ItemDao.GetTriggerInfos(user, context)

	return t.CreateTriggerResult(triggerInfos)
}
