package pairec

import (
	"maps"
	"math"

	randv2 "math/rand/v2"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/service/hook"
	"github.com/alibaba/pairec/v2/web"
)

func CallBackHookFunc(context *context.RecommendContext, params ...any) {
	scene := context.GetParameter("scene").(string)
	callbackFlag := false
	rate := 0
	if sceneConf, ok := context.Config.SceneConfs[scene]; ok {
		if categoryConf, ok := sceneConf["default"]; ok {
			if categoryConf.AutoInvokeCallBack {
				callbackFlag = true
				rate = categoryConf.AutoInvokeCallBackRate
			}
		}
	}

	if !context.Debug && !callbackFlag {
		return
	}

	if callbackFlag && (rate > 0 && rate < 100) {
		if randv2.IntN(100) >= rate {
			return
		}
	}

	callbackConfig, ok := context.Config.CallBackConfs[scene]
	if !ok {
		return
	}

	user := params[0].(*module.User)
	items := params[1].([]*module.Item)

	if callbackConfig.ItemSize > 0 && len(items) > callbackConfig.ItemSize {
		items = items[:callbackConfig.ItemSize]
	}
	if callbackConfig.ItemSizeRate > 0 && callbackConfig.ItemSizeRate < 100 {
		originSize := len(items)
		targetSize := int(math.Round(float64(originSize) * float64(callbackConfig.ItemSizeRate) / 100.0))
		if targetSize == 0 && originSize > 0 {
			targetSize = 1
		}
		if targetSize < originSize {
			step := float64(originSize) / float64(targetSize)
			newItems := make([]*module.Item, 0, targetSize)
			start := randv2.Float64() * step

			for i := 0; i < targetSize; i++ {
				index := int(start + float64(i)*step)
				if index < originSize {
					newItems = append(newItems, items[index])
				}
			}
			if len(newItems) > 0 {
				items = newItems
			}
		}
	}
	var features map[string]any
	if callbackConfig.UseUserFeatures {
		features = user.MakeUserFeatures2()
	} else {
		// Clone the shared features map from RecommendContext. The original
		// HTTP self-call path implicitly isolated this map via a
		// jsonFast.Marshal + json.Unmarshal round-trip; without that copy,
		// the worker goroutine (doCallbackLog) and the recommend main path
		// would share the same map, which can trigger a Go fatal on
		// concurrent map read/write.
		if raw := context.GetParameter("features"); raw != nil {
			if src, ok := raw.(map[string]any); ok {
				features = maps.Clone(src)
			}
		}
	}

	itemList := make([]map[string]any, 0, len(items))
	for _, item := range items {
		data := make(map[string]any)
		data["item_id"] = item.Id
		data["score"] = item.Score
		itemFeatutres := item.GetCloneFeatures()
		for k, v := range itemFeatutres {
			data[k] = v
		}

		scores := item.CloneAlgoScores()
		data["algo_scores"] = scores

		itemList = append(itemList, data)
	}

	// Build CallBackParam directly and dispatch to the worker channel via
	// SendDirect, bypassing the HTTP self-call (jsonFast.Marshal +
	// ForwardWithReader + io.ReadAll + json.Unmarshal). The downstream
	// worker behavior is identical to /api/callback so DataHub output
	// stays byte-for-byte consistent.
	param := &web.CallBackParam{
		SceneId:   scene,
		RequestId: context.RecommendId,
		Uid:       string(user.Id),
		Features:  web.FeaturesMap(features),
		ItemList:  itemList,
		Debug:     context.Debug,
	}
	if context.GetParameter("complex_type_features") != nil {
		if complexTypeFeatures, ok := context.GetParameter("complex_type_features").(web.ComplexTypeFeatures); ok {
			if len(complexTypeFeatures.Features) > 0 {
				param.ComplexTypeFeatures = complexTypeFeatures
			}
		}
	}

	web.SendDirect(param)
}

func init() {
	hook.AddRecommendCleanHook(CallBackHookFunc)
}
