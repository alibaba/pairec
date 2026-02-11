package recallenginerecall

import (
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/recallengine"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type TriggerResult struct {
	TriggerItem       string
	DistinctParam     string
	DistinctParamName string
}
type TriggerKey interface {
	GetTriggerKey(user *module.User, context *context.RecommendContext) *TriggerResult
}

func NewTriggerKey(recallParam *recconf.RecallEngineParam, client *recallengine.RecallEngineClient) TriggerKey {
	switch recallParam.TriggerType {
	case "user":
		trigger := NewUserTrigger(recallParam.UserTriggers)
		return trigger

	case "fixvalue":
		trigger := &FixValueTrigger{value: recallParam.TriggerValue}
		return trigger
	case "item":
		trigger := NewItemTrigger()
		return trigger
	/*
		case "user_vector":
			trigger := NewUserVectorTrigger(&recallParam.UserVectorTrigger)
			return trigger
		case "u2i_realtime":
			trigger := NewU2IRealtimeTrigger(&recallParam.UserTriggerDaoConf, &recallParam.UserTriggerRulesConf)
			return trigger
		case "u2i":
			trigger := NewU2ITrigger(&recallParam.UserCollaborativeDaoConf, &recallParam.UserTriggerRulesConf)
			return trigger
		case "user_realtime_embedding":
			trigger := NewUserRealtimeEmbeddingTrigger(&recallParam.UserRealtimeEmbeddingTrigger)
			return trigger
		case "user_realtime_embedding_mind":
			trigger := NewUserRealtimeEmbeddingMindTrigger(&recallParam.UserRealtimeEmbeddingTrigger)
			return trigger
		case "dssm_o2o":
			trigger := NewUserEmbeddingDssmO2OTrigger(&recallParam.UserEmbeddingO2OTrigger)
			return trigger
		case "mind_o2o":
			trigger := NewUserEmbeddingMindO2OTrigger(&recallParam.UserEmbeddingO2OTrigger)
			return trigger
	*/
	default:
		panic(recallParam.TriggerType + "not support")
	}
}

type UserTrigger struct {
	trigger *module.Trigger
}

func NewUserTrigger(userTriggers []recconf.TriggerConfig) *UserTrigger {
	t := UserTrigger{
		trigger: module.NewTrigger(userTriggers),
	}

	return &t
}
func (t *UserTrigger) GetTriggerKey(user *module.User, context *context.RecommendContext) *TriggerResult {
	return &TriggerResult{
		TriggerItem: t.trigger.GetValue(user.MakeUserFeatures2()),
	}
}

type FixValueTrigger struct {
	value string
}

func (t *FixValueTrigger) GetTriggerKey(user *module.User, context *context.RecommendContext) *TriggerResult {
	return &TriggerResult{
		TriggerItem: t.value,
	}
}

type ItemTrigger struct {
}

func NewItemTrigger() *ItemTrigger {
	t := ItemTrigger{}

	return &t
}
func (t *ItemTrigger) GetTriggerKey(user *module.User, context *context.RecommendContext) *TriggerResult {
	item_id := utils.ToString(context.GetParameter("item_id"), "")
	if item_id == "" {
		return &TriggerResult{}
	}

	return &TriggerResult{
		TriggerItem: item_id,
	}
}
