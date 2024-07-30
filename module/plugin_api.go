package module

import "github.com/alibaba/pairec/v2/context"

type PluginAPIRequest struct {
	Uid       string           `json:"uid"`
	Size      int              `json:"size"`
	SceneId   string           `json:"scene_id"`
	Features  map[string]any   `json:"features"`
	ItemId    string           `json:"item_id"`
	ItemList  []map[string]any `json:"item_list"`
	Debug     bool             `json:"debug"`
	RequestId string           `json:"request_id"`
}

type PluginAPIFilterResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	RequestId string `json:"request_id"`
	Items     []string
}

type PluginAPISortResponse = PluginAPIFilterResponse

func NewPluginAPIRequest(user *User, items []*Item, ctx *context.RecommendContext) *PluginAPIRequest {
	request := &PluginAPIRequest{}

	request.Uid = string(user.Id)
	request.Size = ctx.Size

	if scene, ok := ctx.GetParameter("scene").(string); ok {
		request.SceneId = scene
	}

	request.Features = user.Properties

	if itemId, ok := ctx.GetParameter("item_id").(string); ok {
		request.ItemId = itemId
	}

	request.ItemList = make([]map[string]any, 0, len(items))
	for _, item := range items {
		itemData := make(map[string]interface{})
		itemData["item_id"] = item.Id
		itemData["score"] = item.Score
		itemData["retrieve_id"] = item.RetrieveId

		if item.ItemType != "" {
			itemData["item_type"] = item.ItemType
		}
		if item.Embedding != nil {
			itemData["embedding"] = item.Embedding
		}
		if item.Properties != nil {
			for k, v := range item.Properties {
				itemData[k] = v
			}
		}
		if item.algoScores != nil {
			itemData["algo_scores"] = item.algoScores
		}

		request.ItemList = append(request.ItemList, itemData)
	}

	request.Debug = ctx.Debug

	request.RequestId = ctx.RecommendId

	return request
}
