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

type FsProjectApiService service

/*
FsProjectApiService List Projects

@return ListProjectsResponse
*/
func (a *FsProjectApiService) ListProjects() (ListProjectsResponse, error) {
	request := paifeaturestore.ListProjectsRequest{}
	request.SetName(a.client.cfg.projectName)

	response, err := a.client.ListProjects(&a.client.instanceId, &request)
	var (
		localVarReturnValue ListProjectsResponse
	)

	if err != nil {
		return localVarReturnValue, err
	}

	var projects []*Project
	for _, projectItem := range response.Body.Projects {
		if id, err := strconv.Atoi(*projectItem.ProjectId); err == nil {
			project := Project{
				ProjectId:   id,
				ProjectName: *projectItem.Name,
			}
			if id, err := strconv.Atoi(*projectItem.OfflineDatasourceId); err == nil {
				project.OfflineDatasourceId = id
			}
			if id, err := strconv.Atoi(*projectItem.OnlineDatasourceId); err == nil {
				project.OnlineDatasourceId = id
			}

			switch *projectItem.OfflineDatasourceType {
			case "MaxCompute":
				project.OfflineDatasourceType = constants.Datasource_Type_MaxCompute
			}

			switch *projectItem.OnlineDatasourceType {
			case "Hologres":
				project.OnlineDatasourceType = constants.Datasource_Type_Hologres
			case "GraphCompute":
				project.OnlineDatasourceType = constants.Datasource_Type_IGraph
			case "Tablestore":
				project.OnlineDatasourceType = constants.Datasource_Type_TableStore
			}

			projects = append(projects, &project)
		}
	}

	localVarReturnValue.Projects = projects

	return localVarReturnValue, nil
}
