package api

import (
	"context"
	"strconv"

	paifeaturestore "github.com/alibabacloud-go/paifeaturestore-20230621/client"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/constants"
)

// Linger please
var (
	_ context.Context
)

type FeatureViewApiService service

/*
FeatureViewApiService Get FeatureView By ID
  - @param featureViewId

@return GetFeatureViewResponse
*/
func (a *FeatureViewApiService) GetFeatureViewByID(featureViewId string) (GetFeatureViewResponse, error) {
	var (
		localVarReturnValue GetFeatureViewResponse
	)

	response, err := a.client.GetFeatureView(&a.client.instanceId, &featureViewId)
	if err != nil {
		return localVarReturnValue, err
	}

	featureView := FeatureView{
		ProjectName:       *response.Body.ProjectName,
		FeatureEntityName: *response.Body.FeatureEntityName,
		Name:              *response.Body.Name,
		Type:              *response.Body.Type,
		Online:            *response.Body.SyncOnlineTable,
		Ttl:               int(*response.Body.TTL),
		Config:            *response.Body.Config,
	}
	if response.Body.RegisterTable != nil && *response.Body.RegisterTable != "" {
		featureView.RegisterTable = *response.Body.RegisterTable
		featureView.IsRegister = true
	}
	if response.Body.RegisterDatasourceId != nil && *response.Body.RegisterDatasourceId != "" {
		if id, err := strconv.Atoi(*response.Body.RegisterDatasourceId); err == nil {
			featureView.RegisterDatasourceId = id
		}
	}

	if response.Body.LastSyncConfig != nil && *response.Body.LastSyncConfig != "" {
		featureView.LasySyncConfig = *response.Body.LastSyncConfig
	}
	featureView.FeatureViewId, _ = strconv.Atoi(featureViewId)

	if id, err := strconv.Atoi(*response.Body.ProjectId); err == nil {
		featureView.ProjectId = id
	}
	if id, err := strconv.Atoi(*response.Body.FeatureEntityId); err == nil {
		featureView.FeatureEntityId = id
	}

	var fields []*FeatureViewFields
	for i, fieldItem := range response.Body.Fields {
		field := FeatureViewFields{
			Name:     *fieldItem.Name,
			Position: i + 1,
		}

		switch *fieldItem.Type {
		case "INT32":
			field.Type = int32(constants.FS_INT32)
		case "INT64":
			field.Type = int32(constants.FS_INT64)
		case "FLOAT":
			field.Type = int32(constants.FS_FLOAT)
		case "DOUBLE":
			field.Type = int32(constants.FS_DOUBLE)
		case "BOOLEAN":
			field.Type = int32(constants.FS_BOOLEAN)
		case "TIMESTAMP":
			field.Type = int32(constants.FS_TIMESTAMP)
		default:
			field.Type = int32(constants.FS_STRING)
		}

		for _, attr := range fieldItem.Attributes {
			switch *attr {
			case "Partition":
				field.IsPartition = true
			case "PrimaryKey":
				field.IsPrimaryKey = true
			case "EventTime":
				field.IsEventTime = true
			}
		}

		fields = append(fields, &field)
	}

	featureView.Fields = fields
	localVarReturnValue.FeatureView = &featureView
	return localVarReturnValue, nil
}

/*
FeatureViewApiService List FeatureViews
 * @param optional nil or *FeatureViewApiListFeatureViewsOpts - Optional Parameters:
     * @param "Pagesize" (optional.Int32) -
     * @param "Pagenumber" (optional.Int32) -
     * @param "ProjectId" (optional.Int32) -
     * @param "Owner" (optional.String) -
     * @param "Tag" (optional.String) -
     * @param "Feature" (optional.String) -
@return ListFeatureViewsResponse
*/

func (a *FeatureViewApiService) ListFeatureViews(pagesize, pagenumber int32, projectId string) (ListFeatureViewsResponse, error) {
	var (
		localVarReturnValue ListFeatureViewsResponse
	)

	request := paifeaturestore.ListFeatureViewsRequest{}
	request.SetPageSize(pagesize)
	request.SetPageNumber(pagenumber)
	request.SetProjectId(projectId)

	response, err := a.client.ListFeatureViews(&a.client.instanceId, &request)
	if err != nil {
		return localVarReturnValue, nil
	}

	localVarReturnValue.TotalCount = int(*response.Body.TotalCount)
	var featureViews []*FeatureView

	for _, view := range response.Body.FeatureViews {
		if viewId, err := strconv.Atoi(*view.FeatureViewId); err == nil {
			featureView := FeatureView{
				FeatureViewId:     viewId,
				Type:              *view.Type,
				FeatureEntityName: *view.FeatureEntityName,
				ProjectName:       *view.ProjectName,
			}
			if id, err := strconv.Atoi(*view.ProjectId); err == nil {
				featureView.ProjectId = id
			}

			featureViews = append(featureViews, &featureView)
		}
	}

	localVarReturnValue.FeatureViews = featureViews

	return localVarReturnValue, nil
}
