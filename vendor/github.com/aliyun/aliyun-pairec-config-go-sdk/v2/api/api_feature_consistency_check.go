package api

import (
	"encoding/json"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/pairecservice"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
)

type FeatureConsistencyCheckService service

/*
BackflowFeatureConsistencyCheckJobData send backflow log data to pairec config server

@return FeatureConsistencyBackflowResponse
*/
func (a *FeatureConsistencyCheckService) BackflowFeatureConsistencyCheckJobData(backflowData *model.FeatureConsistencyBackflowData) (FeatureConsistencyBackflowResponse, error) {
	backflowData.InstanceId = a.instanceId

	request := pairecservice.CreateBackflowFeatureConsistencyCheckJobDataRequest()
	request.Domain = a.client.GetDomain()
	request.Headers["Content-Type"] = "application/json"
	var (
		localVarReturnValue FeatureConsistencyBackflowResponse
	)
	body, _ := json.Marshal(backflowData)
	request.Content = body
	response, err := a.client.BackflowFeatureConsistencyCheckJobData(request)

	if err != nil {
		return localVarReturnValue, err
	}

	err = json.Unmarshal(response.GetHttpContentBytes(), &localVarReturnValue)
	if err != nil {
		return localVarReturnValue, err
	}

	return localVarReturnValue, nil
}

/*
SyncFeatureConsistencyCheckJobReplayLog send reply log data to pairec config server

@return FeatureConsistencyReplyResponse
*/
func (a *FeatureConsistencyCheckService) SyncFeatureConsistencyCheckJobReplayLog(replyData *model.FeatureConsistencyReplyData) (FeatureConsistencyReplyResponse, error) {
	replyData.InstanceId = a.instanceId
	request := pairecservice.CreateSyncFeatureConsistencyCheckJobReplayLogRequest()
	request.Domain = a.client.GetDomain()
	request.Headers["Content-Type"] = "application/json"
	var (
		localVarReturnValue FeatureConsistencyReplyResponse
	)

	body, _ := json.Marshal(replyData)
	request.Content = body

	response, err := a.client.SyncFeatureConsistencyCheckJobReplayLog(request)
	if err != nil {
		return localVarReturnValue, err
	}

	err = json.Unmarshal(response.GetHttpContentBytes(), &localVarReturnValue)
	if err != nil {
		return localVarReturnValue, err
	}

	return localVarReturnValue, nil
}
