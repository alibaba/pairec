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

type ParamApiService service

/*
 ParamApiService get params By scene id
  * @param sceneId param of scene Id
  * @param optional nil or *ParamApiGetParamOpts - Optional Parameters:
	  * @param "Environment" (optional.String) -  environment value
	  * @param "ParamId" (optional.Int64) -  param id
	  * @param "ParamName" (optional.String) -  param name
 @return Response
*/

type ParamApiGetParamOpts struct {
	Environment optional.String
	ParamId     optional.Int64
	ParamName   optional.String
}

func (a *ParamApiService) GetParam(sceneId int64, localVarOptionals *ParamApiGetParamOpts) (ListParamsResponse, error) {
	listParamsRequest := pairecservice.CreateListParamsRequest()
	listParamsRequest.InstanceId = a.instanceId
	listParamsRequest.SceneId = strconv.Itoa(int(sceneId))
	listParamsRequest.SetDomain(a.client.GetDomain())
	var (
		localVarReturnValue ListParamsResponse
	)

	if localVarOptionals != nil && localVarOptionals.Environment.IsSet() {
		listParamsRequest.Environment = common.EnvironmentDesc2OpenApiString[localVarOptionals.Environment.Value()]
	}
	response, err := a.client.ListParams(listParamsRequest)
	if err != nil {
		return localVarReturnValue, err
	}
	for _, item := range response.Params {
		if id, err := strconv.Atoi(item.ParamId); err == nil {
			param := model.Param{
				ParamId:     int64(id),
				ParamName:   item.Name,
				ParamValue:  item.Value,
				Environment: int32(common.OpenapiEnvironment2Environment[item.Environment]),
			}

			localVarReturnValue.Params = append(localVarReturnValue.Params, &param)
		}
	}

	return localVarReturnValue, nil
}
