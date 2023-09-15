package berecall

import (
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
)

type U2ITrigger struct {
	*U2IBaseTrigger
	user2ItemDao module.UserCollaborativeDao
}

func NewU2ITrigger(config *recconf.UserCollaborativeDaoConfig, rulesConfig *recconf.UserTriggerRulesConfig) *U2ITrigger {
	conf := recconf.RecallConfig{
		UserCollaborativeDaoConf: *config,
	}
	trigger := &U2ITrigger{
		U2IBaseTrigger: NewU2IBaseTrigger(rulesConfig),
		user2ItemDao:   module.NewUserCollaborativeDao(conf),
	}

	return trigger
}
func (t *U2ITrigger) GetTriggerKey(user *module.User, context *context.RecommendContext) *TriggerResult {
	triggerInfos := t.user2ItemDao.GetTriggerInfos(user, context)

	return t.CreateTriggerResult(triggerInfos)
}
