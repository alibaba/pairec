package api

import (
	"context"
	"strconv"

	paifeaturestore "github.com/alibabacloud-go/paifeaturestore-20230621/client"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/constants"
	"github.com/antihax/optional"
)

// Linger please
var (
	_ context.Context
)

type FsModelApiService service

/*
FsModelApiService Get Model By ID
  - @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
  - @param modelId

@return InlineResponse20086
*/
func (a *FsModelApiService) GetModelByID(modelId string) (GetModelResponse, error) {
	var (
		localVarReturnValue GetModelResponse
	)

	response, err := a.client.GetModelFeature(&a.client.instanceId, &modelId)
	if err != nil {
		return localVarReturnValue, err
	}

	mid, _ := strconv.Atoi(modelId)
	model := Model{
		ModelId:     mid,
		ProjectName: *response.Body.ProjectName,
		Name:        *response.Body.Name,
	}
	if id, err := strconv.Atoi(*response.Body.ProjectId); err == nil {
		model.ProjectId = id
	}

	var features []*ModelFeatures
	for _, featureItem := range response.Body.Features {
		feature := ModelFeatures{
			FeatureViewName: *featureItem.FeatureViewName,
			Name:            *featureItem.Name,
		}
		if featureItem.AliasName != nil && *featureItem.AliasName != "" && *featureItem.AliasName != feature.Name {
			feature.AliasName = *featureItem.AliasName
		}
		if id, err := strconv.Atoi(*featureItem.FeatureViewId); err == nil {
			feature.FeatureViewId = id
		}
		switch *featureItem.Type {
		case "INT32":
			feature.Type = int32(constants.FS_INT32)
		case "INT64":
			feature.Type = int32(constants.FS_INT64)
		case "FLOAT":
			feature.Type = int32(constants.FS_FLOAT)
		case "DOUBLE":
			feature.Type = int32(constants.FS_DOUBLE)
		case "BOOLEAN":
			feature.Type = int32(constants.FS_BOOLEAN)
		case "TIMESTAMP":
			feature.Type = int32(constants.FS_TIMESTAMP)
		default:
			feature.Type = int32(constants.FS_STRING)
		}

		features = append(features, &feature)
	}

	model.Features = features
	localVarReturnValue.Model = &model
	return localVarReturnValue, nil
}

/*
FsModelApiService List Models
 * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param optional nil or *FsModelApiListModelsOpts - Optional Parameters:
     * @param "Pagesize" (optional.Int32) -
     * @param "Pagenumber" (optional.Int32) -
     * @param "ProjectId" (optional.Int32) -
@return InlineResponse20085
*/

type FsModelApiListModelsOpts struct {
	Pagesize   optional.Int32
	Pagenumber optional.Int32
	ProjectId  optional.Int32
}

func (a *FsModelApiService) ListModels(pagesize, pagenumber int, projectId string) (ListModelsResponse, error) {
	var (
		localVarReturnValue ListModelsResponse
	)
	request := paifeaturestore.ListModelFeaturesRequest{}
	request.SetPageSize(strconv.Itoa(pagesize))
	request.SetPageNumber(strconv.Itoa(pagenumber))
	request.SetProjectId(projectId)

	response, err := a.client.ListModelFeatures(&a.client.instanceId, &request)
	if err != nil {
		return localVarReturnValue, err
	}

	localVarReturnValue.TotalCount = int(*response.Body.TotalCount)
	var models []*Model
	for _, modelFeature := range response.Body.ModelFeatures {
		if id, err := strconv.Atoi(*modelFeature.ModelFeatureId); err == nil {
			model := Model{
				ModelId:     id,
				Name:        *modelFeature.Name,
				ProjectName: *modelFeature.ProjectName,
			}
			if id, err := strconv.Atoi(*modelFeature.ProjectId); err == nil {
				model.ProjectId = id
			}

			models = append(models, &model)
		}
	}

	localVarReturnValue.Models = models
	return localVarReturnValue, nil
}
