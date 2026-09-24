package hook

import (
	"fmt"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
)

type RecommendCleanHookFunc func(context *context.RecommendContext, params ...interface{})

var RecommendCleanHooks = make([]RecommendCleanHookFunc, 0)

// SafeRun invokes a RecommendCleanHookFunc with a top-level recover guard.
//
// RecommendCleanHooks are dispatched in their own goroutines from the
// recommend main path (service/user_recommend.go,
// service/user_recall_service.go). A goroutine that panics without a
// deferred recover terminates the entire process — that would let any
// downstream bug in a clean hook (nil deref, type assertion, user-registered
// CallBackProcessFunc, etc.) crash the recommendation server. SafeRun
// localizes the failure: the panic is captured, logged with stack and
// requestId, and the request goroutine continues.
//
// All hook dispatch sites should call SafeRun instead of invoking the hook
// directly.
func SafeRun(hf RecommendCleanHookFunc, ctx *context.RecommendContext, params ...any) {
	defer func() {
		if r := recover(); r != nil {
			requestId := ""
			if ctx != nil {
				requestId = ctx.RecommendId
			}
			stack := strings.ReplaceAll(string(debug.Stack()), "\n", "\t")
			log.Error(fmt.Sprintf("requestId=%s\tmodule=RecommendCleanHook\terror=%v\tstack=%s",
				requestId, r, stack))
		}
	}()
	hf(ctx, params...)
}

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
