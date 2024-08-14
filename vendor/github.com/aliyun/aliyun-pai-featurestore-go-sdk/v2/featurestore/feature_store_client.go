package featurestore

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/api"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
)

type ClientOption func(c *FeatureStoreClient)

func WithLogger(l Logger) ClientOption {
	return func(e *FeatureStoreClient) {
		e.Logger = l
	}
}

func WithErrorLogger(l Logger) ClientOption {
	return func(e *FeatureStoreClient) {
		e.ErrorLogger = l
	}
}

// WithDomain set custom domain
func WithDomain(domian string) ClientOption {
	return func(e *FeatureStoreClient) {
		e.domain = domian
	}
}

func WithLoopData(loopLoad bool) ClientOption {
	return func(e *FeatureStoreClient) {
		e.loopLoadData = loopLoad
	}
}

func WithNoDatasourceInitClient() ClientOption {
	return func(e *FeatureStoreClient) {
		e.datasourceInitClient = false
	}

}

type FeatureStoreClient struct {
	// loopLoadData flag to invoke loopLoadProjectData  function
	loopLoadData bool

	// datasourceInitClient flag to init onlinestore  client
	datasourceInitClient bool

	domain string

	client *api.APIClient

	projectMap map[string]*domain.Project

	// Logger specifies a logger used to report internal changes within the writer
	Logger Logger

	// ErrorLogger is the logger to report errors
	ErrorLogger Logger
}

func NewFeatureStoreClient(regionId, accessKeyId, accessKeySecret, projectName string, opts ...ClientOption) (*FeatureStoreClient, error) {
	client := FeatureStoreClient{
		projectMap:           make(map[string]*domain.Project, 0),
		loopLoadData:         true,
		datasourceInitClient: true,
	}

	for _, opt := range opts {
		opt(&client)
	}

	cfg := api.NewConfiguration(regionId, accessKeyId, accessKeySecret, projectName)
	if client.domain != "" {
		cfg.SetDomain(client.domain)
	}

	apiClient, err := api.NewAPIClient(cfg)
	if err != nil {
		return nil, err
	}

	client.client = apiClient

	if err := client.Validate(); err != nil {
		return nil, err
	}

	client.LoadProjectData()

	if client.loopLoadData {
		go client.loopLoadProjectData()
	}

	return &client, nil
}

// Validate check the  FeatureStoreClient value
func (e *FeatureStoreClient) Validate() error {
	// check instance
	if err := e.client.InstanceApi.GetInstance(); err != nil {
		return err
	}

	return nil
}

func (c *FeatureStoreClient) GetProject(name string) (*domain.Project, error) {
	project, ok := c.projectMap[name]
	if ok {
		return project, nil
	}

	return nil, fmt.Errorf("not found project, name:%s", name)
}

func (c *FeatureStoreClient) logError(err error) {
	if c.ErrorLogger != nil {
		c.ErrorLogger.Printf(err.Error())
		return
	}

	if c.Logger != nil {
		c.Logger.Printf(err.Error())
	}
}

// LoadProjectData specifies a function to load data from featurestore server
func (c *FeatureStoreClient) LoadProjectData() {
	ak := api.Ak{
		AccesskeyId:     c.client.GetConfig().AccessKeyId,
		AccesskeySecret: c.client.GetConfig().AccessKeySecret,
	}
	projectData := make(map[string]*domain.Project, 0)

	listProjectsResponse, err := c.client.FsProjectApi.ListProjects()
	if err != nil {
		c.logError(fmt.Errorf("list projects error, err=%v", err))
		return
	}

	for _, p := range listProjectsResponse.Projects {
		// get datasource
		getDataSourceResponse, err := c.client.DatasourceApi.DatasourceDatasourceIdGet(p.OnlineDatasourceId)
		if err != nil {
			c.logError(fmt.Errorf("get datasource error, err=%v", err))
			continue
		}

		p.OnlineDataSource = getDataSourceResponse.Datasource
		p.OnlineDataSource.Ak = ak

		getDataSourceResponse, err = c.client.DatasourceApi.DatasourceDatasourceIdGet(p.OfflineDatasourceId)
		if err != nil {
			c.logError(fmt.Errorf("get datasource error, err=%v", err))
			continue
		}

		p.OfflineDataSource = getDataSourceResponse.Datasource
		p.OfflineDataSource.Ak = ak

		project := domain.NewProject(p, c.datasourceInitClient)
		projectData[project.ProjectName] = project

		// get feature entities
		listFeatureEntitiesResponse, err := c.client.FeatureEntityApi.ListFeatureEntities(strconv.Itoa(p.ProjectId))
		if err != nil {
			c.logError(fmt.Errorf("list feature entities error, err=%v", err))
			continue
		}

		for _, entity := range listFeatureEntitiesResponse.FeatureEntities {
			if entity.ProjectId == project.ProjectId {
				project.FeatureEntityMap[entity.FeatureEntityName] = domain.NewFeatureEntity(entity)
			}
		}

		var (
			pagesize   = 100
			pagenumber = 1
		)
		// get feature views
		for {
			listFeatureViews, err := c.client.FeatureViewApi.ListFeatureViews(int32(pagesize), int32(pagenumber), strconv.Itoa(p.ProjectId))
			if err != nil {
				c.logError(fmt.Errorf("list feature views error, err=%v", err))
				continue
			}

			for _, view := range listFeatureViews.FeatureViews {
				getFeatureViewResponse, err := c.client.FeatureViewApi.GetFeatureViewByID(strconv.Itoa(int(view.FeatureViewId)))
				if err != nil {
					c.logError(fmt.Errorf("get feature view error, err=%v", err))
					continue
				}
				featureView := getFeatureViewResponse.FeatureView
				if featureView.RegisterDatasourceId > 0 {
					getDataSourceResponse, err := c.client.DatasourceApi.DatasourceDatasourceIdGet(featureView.RegisterDatasourceId)
					if err != nil {
						c.logError(fmt.Errorf("get datasource error, err=%v", err))
						continue
					}
					featureView.RegisterDataSource = getDataSourceResponse.Datasource
				}

				featureViewDomain := domain.NewFeatureView(featureView, project, project.FeatureEntityMap[featureView.FeatureEntityName])
				project.FeatureViewMap[featureView.Name] = featureViewDomain

			}

			if len(listFeatureViews.FeatureViews) == 0 || pagesize*pagenumber > listFeatureViews.TotalCount {
				break
			}

			pagenumber++

		}

		pagenumber = 1
		// get model
		for {
			listModelsResponse, err := c.client.FsModelApi.ListModels(pagesize, pagenumber, strconv.Itoa(project.ProjectId))
			if err != nil {
				c.logError(fmt.Errorf("list models error, err=%v", err))
				continue
			}

			for _, m := range listModelsResponse.Models {
				getModelResponse, err := c.client.FsModelApi.GetModelByID(strconv.Itoa(m.ModelId))
				if err != nil {
					c.logError(fmt.Errorf("get model error, err=%v", err))
					continue
				}
				model := getModelResponse.Model
				modelDomain := domain.NewModel(model, project)
				project.ModelMap[model.Name] = modelDomain

			}

			if len(listModelsResponse.Models) == 0 || pagenumber*pagesize > int(listModelsResponse.TotalCount) {
				break
			}

			pagenumber++

		}

	}

	if len(projectData) > 0 {
		c.projectMap = projectData
	}
}

func (c *FeatureStoreClient) loopLoadProjectData() {
	for {
		time.Sleep(time.Minute)
		c.LoadProjectData()
	}
}
