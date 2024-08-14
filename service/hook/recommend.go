package hook

import (
	"sync"

	"github.com/alibaba/pairec/v2/context"
)

type RecommendCleanHookFunc func(context *context.RecommendContext, params ...interface{})

var RecommendCleanHooks = make([]RecommendCleanHookFunc, 0)

func AddRecommendCleanHook(hf ...RecommendCleanHookFunc) {
	RecommendCleanHooks = append(RecommendCleanHooks, hf...)
}

var mu sync.Mutex
var RecommendCleanHookMap = make(map[string]int)

func RegisterRecommendCleanHook(name string, hf RecommendCleanHookFunc) {
	mu.Lock()
	defer mu.Unlock()

	if index, exist := RecommendCleanHookMap[name]; exist {
		var hookfuncs []RecommendCleanHookFunc
		for i, hookfunc := range RecommendCleanHooks {
			if i != index {
				hookfuncs = append(hookfuncs, hookfunc)
			}
		}
		RecommendCleanHooks = hookfuncs
	}
	AddRecommendCleanHook(hf)
	RecommendCleanHookMap[name] = len(RecommendCleanHooks) - 1
}
