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
		if index >= 0 && index < len(RecommendCleanHooks) {
			RecommendCleanHooks[index] = hf
		}
	} else {
		AddRecommendCleanHook(hf)
		RecommendCleanHookMap[name] = len(RecommendCleanHooks) - 1
	}
}

func RemoveRecommendCleanHook(name string) {
	mu.Lock()
	defer mu.Unlock()

	removeIndex := -1
	if index, exist := RecommendCleanHookMap[name]; exist {
		removeIndex = index
	} else {
		return
	}
	indexMap := make(map[int]string)
	for n, index := range RecommendCleanHookMap {
		if n == name {
			continue
		}
		indexMap[index] = n
	}

	var hookfuncs []RecommendCleanHookFunc
	for index, hf := range RecommendCleanHooks {
		if index == removeIndex {
			continue
		}
		if n, ok := indexMap[index]; ok {
			hookfuncs = append(hookfuncs, hf)
			RecommendCleanHookMap[n] = len(hookfuncs) - 1
		} else {
			hookfuncs = append(hookfuncs, hf)
		}
	}

	RecommendCleanHooks = hookfuncs
	delete(RecommendCleanHookMap, name)
}
