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

type LayerApiService service

/*
LayerApiService list all Layers By filter condition
  - @param expRoomId list all layers of the experiment room

@return Response
*/
func (a *LayerApiService) ListLayers(expRoomId int64) (ListLayersResponse, error) {
	listLayersRequest := pairecservice.CreateListLayersRequest()
	listLayersRequest.LaboratoryId = strconv.Itoa(int(expRoomId))
	listLayersRequest.InstanceId = a.instanceId
	listLayersRequest.SetDomain(a.client.GetDomain())
	var (
		localVarReturnValue ListLayersResponse
	)

	response, err := a.client.ListLayers(listLayersRequest)
	if err != nil {
		return localVarReturnValue, err
	}
	for _, item := range response.Layers {
		if id, err := strconv.Atoi(item.LayerId); err == nil {
			layer := &model.Layer{
				LayerId:   int64(id),
				ExpRoomId: expRoomId,
				LayerName: item.Name,
				LayerInfo: item.Description,
			}
			if sceneId, err := strconv.Atoi(item.SceneId); err == nil {
				layer.SceneId = int64(sceneId)
			}

			localVarReturnValue.Layers = append(localVarReturnValue.Layers, layer)
		}
	}
	return localVarReturnValue, nil
}
