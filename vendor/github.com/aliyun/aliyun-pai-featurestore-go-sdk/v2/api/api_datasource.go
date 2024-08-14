package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/constants"
)

type DatasourceApiService service

/*
DatasourceApiService Get datasource By datasource_id
  - @param datasourceId

@return GetDatasourceResponse
*/
func (a *DatasourceApiService) DatasourceDatasourceIdGet(datasourceId int) (GetDatasourceResponse, error) {

	var (
		localVarReturnValue GetDatasourceResponse
	)

	datasourceIdStr := strconv.Itoa(datasourceId)
	response, err := a.client.GetDatasource(&a.client.instanceId, &datasourceIdStr)
	if err != nil {
		return localVarReturnValue, err
	}

	datasource := Datasource{
		DatasourceId: datasourceId,
		Type:         *response.Body.Type,
		Name:         *response.Body.Name,
	}
	switch *response.Body.Type {
	case "Hologres":
		datasource.Type = constants.Datasource_Type_Hologres
		uris := strings.Split(*response.Body.Uri, "/")
		datasource.Database = uris[1]
		datasource.VpcAddress = fmt.Sprintf("%s-%s-vpc-st.hologres.aliyuncs.com:80", uris[0], a.client.cfg.regionId)
		//datasource.VpcAddress = fmt.Sprintf("%s-%s.hologres.aliyuncs.com:80", uris[0], a.client.cfg.regionId)
	case "GraphCompute":
		datasource.Type = constants.Datasource_Type_IGraph
		var config map[string]string
		if err := json.Unmarshal([]byte(*response.Body.Config), &config); err == nil {
			datasource.VpcAddress = config["address"]
			datasource.User = config["username"]
			datasource.Pwd = config["password"]
		}
		datasource.RdsInstanceId = *response.Body.Uri
	case "Tablestore":
		datasource.Type = constants.Datasource_Type_TableStore
		datasource.VpcAddress = fmt.Sprintf("https://%s.%s.vpc.tablestore.aliyuncs.com", *response.Body.Uri, a.client.cfg.regionId)
		datasource.RdsInstanceId = *response.Body.Uri
	case "MaxCompute":
		datasource.Type = constants.Datasource_Type_MaxCompute
		datasource.Project = *response.Body.Uri
	}

	localVarReturnValue.Datasource = &datasource

	return localVarReturnValue, nil
}
