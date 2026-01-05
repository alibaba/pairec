package recallenginerecall

import (
	"fmt"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/recallengine"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	be "github.com/aliyun/aliyun-be-go-sdk"
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
		TriggerItem: fmt.Sprintf("%s:%d", t.trigger.GetValue(user.MakeUserFeatures2()), 1),
	}
}

type BeTrigger struct {
	bizName   string
	fieldName string
	beClient  *be.Client
}

func (t *BeTrigger) GetTriggerKey(user *module.User, context *context.RecommendContext) *TriggerResult {
	x2iReadRequest := be.NewReadRequest(t.bizName, 1)
	x2iRecallParams := be.NewRecallParam().
		SetTriggerItems([]string{string(user.Id) + ":1"}).
		SetRecallType(be.RecallTypeX2I)
	x2iRecallParams.ReturnCount = 1
	x2iReadRequest.AddRecallParam(x2iRecallParams)

	triggerResult := &TriggerResult{}
	x2iReadResponse, err := t.beClient.Read(*x2iReadRequest)
	if err != nil {
		log.Error(fmt.Sprintf("BeTrigger read error:%v", err))
		return triggerResult
	}

	mathItems := x2iReadResponse.Result.MatchItems
	if mathItems == nil || len(mathItems.FieldValues) == 0 {
		return triggerResult
	}

	for i, name := range mathItems.FieldNames {
		if name == t.fieldName {
			triggerResult.TriggerItem = utils.ToString(mathItems.FieldValues[0][i], "")
			return triggerResult
		}
	}

	return triggerResult
}

type FixValueTrigger struct {
	value string
}

func (t *FixValueTrigger) GetTriggerKey(user *module.User, context *context.RecommendContext) *TriggerResult {
	return &TriggerResult{
		TriggerItem: fmt.Sprintf("%s:%d", t.value, 1),
	}
}
