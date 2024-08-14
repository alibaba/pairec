package api

import (
	"encoding/json"
	"strconv"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/pairecservice"
)

type CrowdApiService service

/*
CrowdApiService Get Crowd users By crowd ID
Get Crowd users By crowd ID

@return ListCrowdUsersResponse
*/
func (a *CrowdApiService) GetCrowdUsersById(crowdId int64) (ListCrowdUsersResponse, error) {
	listCrowdUsersRequest := pairecservice.CreateListCrowdUsersRequest()
	listCrowdUsersRequest.InstanceId = a.instanceId
	listCrowdUsersRequest.Domain = a.client.GetDomain()
	listCrowdUsersRequest.CrowdId = strconv.Itoa(int(crowdId))
	var (
		localVarReturnValue ListCrowdUsersResponse
	)

	response, err := a.client.ListCrowdUsers(listCrowdUsersRequest)
	if err != nil {
		return localVarReturnValue, err
	}

	err = json.Unmarshal(response.GetHttpContentBytes(), &localVarReturnValue)
	if err != nil {
		return localVarReturnValue, err
	}

	localVarReturnValue.Users = localVarReturnValue.CrowdUsers
	return localVarReturnValue, nil
}
