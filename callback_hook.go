package pairec

import (
	"encoding/json"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/service/hook"
	"github.com/alibaba/pairec/v2/web"
)

func CallBackHookFunc(context *context.RecommendContext, params ...any) {
	if !context.Debug {
		return
	}
	scene := context.GetParameter("scene").(string)
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
		"debug":      true,
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
		for k, v := range item.GetProperties() {
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
