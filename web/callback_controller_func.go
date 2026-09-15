package web

import (
	"sync"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
)

type CallBackProcessFunc func(user *module.User, items []*module.Item, context *context.RecommendContext)

// callBackProcessFuncMap holds the per-scene callback processors. It is
// registered into at process start (typically from init() of user code) but
// read on every callback request from the worker pool goroutines, so all
// access must go through callBackProcessFuncMutex to remain race-free.
var (
	callBackProcessFuncMap   map[string]CallBackProcessFunc
	callBackProcessFuncMutex sync.RWMutex
)

func init() {
	callBackProcessFuncMap = make(map[string]CallBackProcessFunc)
}

// RegisterCallBackProcessFunc registers f as the processor for scene. Safe to
// call concurrently with lookupCallBackProcessFunc.
func RegisterCallBackProcessFunc(scene string, f CallBackProcessFunc) {
	callBackProcessFuncMutex.Lock()
	defer callBackProcessFuncMutex.Unlock()
	callBackProcessFuncMap[scene] = f
}

// lookupCallBackProcessFunc returns the registered processor for scene, if
// any. Safe to call concurrently with RegisterCallBackProcessFunc.
func lookupCallBackProcessFunc(scene string) (CallBackProcessFunc, bool) {
	callBackProcessFuncMutex.RLock()
	defer callBackProcessFuncMutex.RUnlock()
	f, ok := callBackProcessFuncMap[scene]
	return f, ok
}

// unregisterCallBackProcessFunc removes the processor for scene. Intended for
// test cleanup; production code does not need this.
func unregisterCallBackProcessFunc(scene string) {
	callBackProcessFuncMutex.Lock()
	defer callBackProcessFuncMutex.Unlock()
	delete(callBackProcessFuncMap, scene)
}
