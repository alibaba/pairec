package hook

import (
	"github.com/alibaba/pairec/context"
)

type RecommendCleanHookFunc func(context *context.RecommendContext, params ...interface{})

var RecommendCleanHooks = make([]RecommendCleanHookFunc, 0)

func AddRecommendCleanHook(hf ...RecommendCleanHookFunc) {
	RecommendCleanHooks = append(RecommendCleanHooks, hf...)
}
