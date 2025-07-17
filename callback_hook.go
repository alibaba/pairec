package pairec

import (
	"encoding/json"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/service/hook"
	"github.com/alibaba/pairec/v2/web"
)

func CallBackHookFunc(context *context.RecommendContext, params ...any) {
	scene := context.GetParameter("scene").(string)
	callbackFlag := false
	if sceneConf, ok := context.Config.SceneConfs[scene]; ok {
		if categoryConf, ok := sceneConf["default"]; ok {
			if categoryConf.AutoInvokeCallBack {
				callbackFlag = true
			}
		}
	}

	if !context.Debug && !callbackFlag {
		return
	}
	if _, ok := context.Config.CallBackConfs[scene]; !ok {
		return
	}
	user := params[0].(*module.User)
	items := params[1].([]*module.Item)
	var features map[string]any
	if context.GetParameter("features") != nil {
		features = context.GetParameter("features").(map[string]any)
	}
	requestData := map[string]any{
		"request_id": context.RecommendId,
		"scene_id":   scene,
		"features":   features,
		"uid":        user.Id,
		"debug":      context.Debug,
	}
	if context.GetParameter("complex_type_features") != nil {
		if complexTypeFeatures, ok := context.GetParameter("complex_type_features").(web.ComplexTypeFeatures); ok {
			if len(complexTypeFeatures.Features) > 0 {
				requestData["complex_type_features"] = complexTypeFeatures.Features
			}
		}
	}

	itemList := make([]map[string]any, 0, len(items))
	for _, item := range items {
		data := make(map[string]any)
		data["item_id"] = item.Id
		itemFeatutres := item.GetCloneFeatures()
		for k, v := range itemFeatutres {
			data[k] = v
		}

		itemList = append(itemList, data)
	}
	requestData["item_list"] = itemList

	d, _ := json.Marshal(requestData)
	response := Forward("POST", "/api/callback", string(d))
	response.Body.Close()
}

func init() {
	hook.AddRecommendCleanHook(CallBackHookFunc)
}
