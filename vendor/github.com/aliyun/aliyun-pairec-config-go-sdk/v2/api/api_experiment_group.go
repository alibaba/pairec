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

type ExperimentGroupApiService service

/*
 ExperimentGroupApiService list all ExperimentGroups By filter condition
  * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
  * @param layerId list all experiment groups of the layer
  * @param optional nil or *ExperimentGroupApiListExperimentGroupsOpts - Optional Parameters:
	  * @param "Status" (optional.Int32) -  list the status experiment groups of the layer
 @return InlineResponse2009
*/

type ExperimentGroupApiListExperimentGroupsOpts struct {
	Status optional.Uint32
}

func (a *ExperimentGroupApiService) ListExperimentGroups(layerId int64, localVarOptionals *ExperimentGroupApiListExperimentGroupsOpts) (ListExperimentGroupsResponse, error) {

	listExperimentGroupsRequest := pairecservice.CreateListExperimentGroupsRequest()
	listExperimentGroupsRequest.InstanceId = a.instanceId
	listExperimentGroupsRequest.LayerId = strconv.Itoa(int(layerId))
	if localVarOptionals.Status.Value() == common.ExpGroup_Status_Online {
		listExperimentGroupsRequest.Status = "Online"
	} else if localVarOptionals.Status.Value() == common.ExpGroup_Status_Offline {
		listExperimentGroupsRequest.Status = "Offline"
	}
	listExperimentGroupsRequest.SetDomain(a.client.GetDomain())
	var (
		localVarReturnValue ListExperimentGroupsResponse
	)

	response, err := a.client.ListExperimentGroups(listExperimentGroupsRequest)
	if err != nil {
		return localVarReturnValue, err
	}
	for _, item := range response.ExperimentGroups {
		if id, err := strconv.Atoi(item.ExperimentGroupId); err == nil {
			experimentGroup := model.ExperimentGroup{
				ExpGroupId:               int64(id),
				LayerId:                  layerId,
				ExpGroupName:             item.Name,
				ExpGroupInfo:             item.Description,
				DebugUsers:               item.DebugUsers,
				Owner:                    item.Owner,
				Filter:                   item.Filter,
				DistributionTimeDuration: item.DistributionTimeDuration,
				ExpGroupConfig:           item.Config,
				ReserveBuckets:           item.ReservedBuckets,
			}
			if item.DebugCrowdId != "" {
				if crowdId, err := strconv.Atoi(item.DebugCrowdId); err == nil {
					experimentGroup.DebugCrowdId = int64(crowdId)
				}
			}

			if sceneId, err := strconv.Atoi(item.SceneId); err == nil {
				experimentGroup.SceneId = int64(sceneId)
			}

			// exproom id
			if laboratoryId, err := strconv.Atoi(item.LaboratoryId); err == nil {
				experimentGroup.ExpRoomId = int64(laboratoryId)
			}
			if item.DistributionType == "UserId" {
				experimentGroup.DistributionType = common.ExpGroup_Distribution_Type_User
			} else if item.DistributionType == "TimeDuration" {
				experimentGroup.DistributionType = common.ExpGroup_Distribution_Type_TimeDuration
			}

			if item.CrowdId != "" {
				if crowdId, err := strconv.Atoi(item.CrowdId); err == nil {
					experimentGroup.CrowdId = int64(crowdId)
				}
			}

			localVarReturnValue.ExperimentGroups = append(localVarReturnValue.ExperimentGroups, &experimentGroup)
		}
	}

	return localVarReturnValue, nil
}
