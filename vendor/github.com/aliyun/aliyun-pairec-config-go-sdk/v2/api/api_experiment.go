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

type ExperimentApiService service

/*
 ExperimentApiService list all Experiments By filter condition
  * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
  * @param expGroupId list all experiments of the experiment group
  * @param optional nil or *ExperimentApiListExperimentsOpts - Optional Parameters:
	  * @param "Status" (optional.Int32) -  list the  experiments of the status
 @return ListExperimentsResponse
*/

type ExperimentApiListExperimentsOpts struct {
	Status optional.Uint32
}

func (a *ExperimentApiService) ListExperiments(expGroupId int64, localVarOptionals *ExperimentApiListExperimentsOpts) (ListExperimentsResponse, error) {
	listExperimentsRequest := pairecservice.CreateListExperimentsRequest()
	listExperimentsRequest.InstanceId = a.instanceId
	listExperimentsRequest.ExperimentGroupId = strconv.Itoa(int(expGroupId))
	if localVarOptionals.Status.Value() == common.Experiment_Status_Online {
		listExperimentsRequest.Status = "Online"
	} else if localVarOptionals.Status.Value() == common.Experiment_Status_Offline {
		listExperimentsRequest.Status = "Offline"
	}
	listExperimentsRequest.SetDomain(a.client.GetDomain())
	var (
		localVarReturnValue ListExperimentsResponse
	)
	response, err := a.client.ListExperiments(listExperimentsRequest)
	if err != nil {
		return localVarReturnValue, err
	}

	for _, item := range response.Experiments {
		if id, err := strconv.Atoi(item.ExperimentId); err == nil {
			experiment := model.Experiment{
				ExperimentId:      int64(id),
				ExpGroupId:        expGroupId,
				ExperimentName:    item.Name,
				ExperimentInfo:    item.Description,
				ExperimentFlow:    uint32(item.FlowPercent),
				ExperimentBuckets: item.Buckets,
				DebugUsers:        item.DebugUsers,
				ExperimentConfig:  item.Config,
			}
			if item.DebugCrowdId != "" {
				if crowdId, err := strconv.Atoi(item.DebugCrowdId); err == nil {
					experiment.DebugCrowdId = int64(crowdId)
				}
			}

			if sceneId, err := strconv.Atoi(item.SceneId); err == nil {
				experiment.SceneId = int64(sceneId)
			}

			// exproom id
			if laboratoryId, err := strconv.Atoi(item.LaboratoryId); err == nil {
				experiment.ExpRoomId = int64(laboratoryId)
			}
			if layerId, err := strconv.Atoi(item.LayerId); err == nil {
				experiment.LayerId = int64(layerId)
			}

			switch item.Type {
			case "Baseline":
				experiment.Type = common.Experiment_Type_Base
			case "Normal":
				experiment.Type = common.Experiment_Type_Test

			}

			localVarReturnValue.Experiments = append(localVarReturnValue.Experiments, &experiment)
		}

	}

	return localVarReturnValue, nil
}
