package api

import (
	"context"
	"strconv"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/pairecservice"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
)

// Linger please
var (
	_ context.Context
)

type SceneApiService service

/*
SceneApiService Get all scenes
@return ListScenesResponse
*/
func (a *SceneApiService) ListAllScenes() (ListScenesResponse, error) {
	listScenesRequest := pairecservice.CreateListScenesRequest()
	listScenesRequest.InstanceId = a.instanceId
	listScenesRequest.SetDomain(a.client.GetDomain())

	var (
		localVarReturnValue ListScenesResponse
	)

	response, err := a.client.ListScenes(listScenesRequest)
	if err != nil {
		return localVarReturnValue, err
	}
	var scenes []*model.Scene
	for _, sceneItem := range response.Scenes {
		if id, err := strconv.Atoi(sceneItem.SceneId); err == nil {
			scene := &model.Scene{
				SceneName: sceneItem.Name,
				SceneId:   int64(id),
			}
			scenes = append(scenes, scene)
		}
	}
	localVarReturnValue.Scenes = scenes

	return localVarReturnValue, nil
}
