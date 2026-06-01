package recallenginerecall

import (
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	re "github.com/aliyun/aliyun-pairec-config-go-sdk/v2/recallengine"
)

const (
	REScoreFieldName  = "score"
	REItemIdFieldName = "item_id"
	RERecallName      = "recall_name"
	//reRecallNameV2 = "__recall_name__"
)

type RecallEngineBaseRecall interface {
	GetItems(user *module.User, context *context.RecommendContext) ([]*module.Item, error)
	//BuildRecallParam(user *module.User, context *context.RecommendContext) *be.RecallParam
	BuildQueryParams(user *module.User, context *context.RecommendContext) re.RecallConf
	CloneWithConfig(params map[string]interface{}) RecallEngineBaseRecall
	GetRecallName() string
}
