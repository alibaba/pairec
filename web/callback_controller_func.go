package web

import (
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
)

type CallBackProcessFunc func(user *module.User, items []*module.Item, context *context.RecommendContext)

var callBackProcessFuncMap map[string]CallBackProcessFunc

func init() {
	callBackProcessFuncMap = make(map[string]CallBackProcessFunc)
}

func RegisterCallBackProcessFunc(scene string, f CallBackProcessFunc) {
	callBackProcessFuncMap[scene] = f
}
