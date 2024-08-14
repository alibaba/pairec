package domain

import (
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/api"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/constants"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/datasource/hologres"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/datasource/igraph"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/datasource/tablestore"
)

type Project struct {
	*api.Project
	OnlineStore      OnlineStore
	FeatureViewMap   map[string]FeatureView
	FeatureEntityMap map[string]*FeatureEntity
	ModelMap         map[string]*Model
}

func NewProject(p *api.Project, isInitClient bool) *Project {
	project := Project{
		Project:          p,
		FeatureViewMap:   make(map[string]FeatureView),
		FeatureEntityMap: make(map[string]*FeatureEntity),
		ModelMap:         make(map[string]*Model),
	}

	switch p.OnlineDatasourceType {
	case constants.Datasource_Type_Hologres:
		onlineStore := &HologresOnlineStore{
			Datasource: p.OnlineDataSource,
		}
		if isInitClient {
			dsn := onlineStore.Datasource.GenerateDSN(constants.Datasource_Type_Hologres)
			hologres.RegisterHologres(onlineStore.Name, dsn)
		}
		project.OnlineStore = onlineStore
	case constants.Datasource_Type_IGraph:
		onlineStore := &IGraphOnlineStore{
			Datasource: p.OnlineDataSource,
		}

		if isInitClient {
			client := igraph.NewGraphClient(p.OnlineDataSource.VpcAddress, p.OnlineDataSource.User, p.OnlineDataSource.Pwd)
			igraph.RegisterGraphClient(onlineStore.Name, client)
		}
		project.OnlineStore = onlineStore
	case constants.Datasource_Type_TableStore:
		onlineStore := &TableStoreOnlineStore{
			Datasource: p.OnlineDataSource,
		}

		if isInitClient {
			client := onlineStore.Datasource.NewTableStoreClient()
			tablestore.RegisterTableStoreClient(onlineStore.Name, client)
		}
		project.OnlineStore = onlineStore
	default:
		panic("not support onlinestore type")
	}

	return &project
}

func (p *Project) GetFeatureView(name string) FeatureView {
	return p.FeatureViewMap[name]
}

func (p *Project) GetFeatureEntity(name string) *FeatureEntity {
	return p.FeatureEntityMap[name]
}

func (p *Project) GetModel(name string) *Model {
	return p.ModelMap[name]
}
func (p *Project) GetModelFeature(name string) *Model {
	return p.ModelMap[name]
}
