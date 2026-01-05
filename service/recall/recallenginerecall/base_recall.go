package recallenginerecall

import (
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
)

type RecallEngineBaseRecall interface {
	GetItems(user *module.User, context *context.RecommendContext) ([]*module.Item, error)
	//BuildRecallParam(user *module.User, context *context.RecommendContext) *be.RecallParam
	BuildQueryParams(user *module.User, context *context.RecommendContext) (ret map[string]string)
	CloneWithConfig(params map[string]interface{}) RecallEngineBaseRecall
}
