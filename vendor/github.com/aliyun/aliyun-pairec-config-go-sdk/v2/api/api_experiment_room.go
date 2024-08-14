package api

import (
	"context"
	"strconv"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/pairecservice"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/common"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
	"github.com/antihax/optional"
)

// Linger please
var (
	_ context.Context
)

type ExperimentRoomApiService service

/*
ExperimentRoomApiService list all ExperimentRooms By filter condition
list all ExperimentRooms By filter condition
 * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param environment environment of experiment room
 * @param optional nil or *ExperimentRoomApiListExperimentRoomsOpts - Optional Parameters:
     * @param "SceneId" (optional.Int64) -  list all experiment rooms of the scene_id
@return InlineResponse2003
*/

type ExperimentRoomApiListExperimentRoomsOpts struct {
	SceneId optional.Int64
	Status  optional.Uint32
}

func (a *ExperimentRoomApiService) ListExperimentRooms(environment string, localVarOptionals *ExperimentRoomApiListExperimentRoomsOpts) (ListExperimentRoomsResponse, error) {
	listLaboratoriesRequest := pairecservice.CreateListLaboratoriesRequest()
	listLaboratoriesRequest.InstanceId = a.instanceId
	listLaboratoriesRequest.Environment = common.EnvironmentDesc2OpenApiString[environment]

	listLaboratoriesRequest.SceneId = strconv.Itoa(int(localVarOptionals.SceneId.Value()))
	if localVarOptionals.Status.Value() == common.ExpRoom_Status_Online {
		listLaboratoriesRequest.Status = "Online"
	} else if localVarOptionals.Status.Value() == common.ExpRoom_Status_Offline {
		listLaboratoriesRequest.Status = "Offline"
	}
	listLaboratoriesRequest.SetDomain(a.client.GetDomain())
	//listLaboratoriesRequest.Status =
	var (
		localVarReturnValue ListExperimentRoomsResponse
	)

	response, err := a.client.ListLaboratories(listLaboratoriesRequest)
	if err != nil {
		return localVarReturnValue, err
	}

	for _, item := range response.Laboratories {
		if id, err := strconv.Atoi(item.LaboratoryId); err == nil {
			experimentRoom := &model.ExperimentRoom{
				ExpRoomId:      int64(id),
				ExpRoomName:    item.Name,
				ExpRoomInfo:    item.Description,
				DebugUsers:     item.DebugUsers,
				BucketCount:    int32(item.BucketCount),
				Filter:         item.Filter,
				ExpRoomBuckets: item.Buckets,
				DebugCrowdId:   0,
			}

			if item.DebugCrowdId != "" {
				if crowdId, err := strconv.Atoi(item.DebugCrowdId); err == nil {
					experimentRoom.DebugCrowdId = int64(crowdId)
				}
			}

			if sceneId, err := strconv.Atoi(item.SceneId); err == nil {
				experimentRoom.SceneId = int64(sceneId)
			}

			// experiment room type
			if item.Type == "Base" {
				experimentRoom.Type = common.ExpRoom_Type_Base
			} else if item.Type == "NonBase" {
				experimentRoom.Type = common.ExpRoom_Type_Normal
			}

			if item.BucketType == "Filter" {
				experimentRoom.BucketType = common.Bucket_Type_Filter
			} else if item.BucketType == "Uid" {
				experimentRoom.BucketType = common.Bucket_Type_UID
			} else if item.BucketType == "UidHash" {
				experimentRoom.BucketType = common.Bucket_Type_UID_HASH
			} else {
				experimentRoom.BucketType = common.Bucket_Type_Custom
			}
			experimentRoom.Environment = int32(common.OpenapiEnvironment2Environment[item.Environment])

			localVarReturnValue.ExperimentRooms = append(localVarReturnValue.ExperimentRooms, experimentRoom)
		}
	}

	return localVarReturnValue, nil
}
