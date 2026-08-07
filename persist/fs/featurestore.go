package fs

import (
	"fmt"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/featurestore"
)

var fsInstances = make(map[string]*FSClient)

func GetFeatureStoreClient(name string) (*FSClient, error) {
	if client, ok := fsInstances[name]; ok {
		return client, nil
	}
	var matched *FSClient
	for _, client := range fsInstances {
		if client.projectName != name {
			continue
		}
		if matched != nil {
			return nil, fmt.Errorf("multiple feature store clients found for project:%s", name)
		}
		matched = client
	}
	if matched == nil {
		return nil, fmt.Errorf("feature store client not found, name or project:%s", name)
	}
	return matched, nil
}

type FSClient struct {
	client      *featurestore.FeatureStoreClient
	project     *domain.Project
	projectName string
}

func (fs *FSClient) GetProject() *domain.Project {
	return fs.project
}

func (fs *FSClient) GetLLMConfig(name string) (*domain.LLMConfig, error) {
	return fs.client.GetLLMConfig(name)
}

func (fs *FSClient) ReloadProject() {
	if p, err := fs.client.GetProject(fs.projectName); err == nil {
		fs.project = p
	} else {
		log.Error(fmt.Sprintf("get project failed, projectName:%s, err:%v", fs.projectName, err))
	}
}

func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.FeatureStoreConfs {
		if fs, ok := fsInstances[name]; ok {
			fs.ReloadProject()
			continue
		}

		hologresPort := 80
		if conf.HologresPort > 0 {
			hologresPort = conf.HologresPort
		}
		var (
			client *featurestore.FeatureStoreClient
			err    error
		)
		l := log.FeatureStoreLogger{}
		options := []featurestore.ClientOption{
			featurestore.WithLogger(featurestore.LoggerFunc(l.Infof)),
			featurestore.WithErrorLogger(featurestore.LoggerFunc(l.Errorf)),
			featurestore.WithFeatureDBLogin(conf.FeatureDBUsername, conf.FeatureDBPassword),
			featurestore.WithHologresPort(hologresPort),
		}
		if conf.Domain != "" {
			options = append(options, featurestore.WithDomain(conf.Domain))
		}
		if conf.TestMode {
			options = append(options, featurestore.WithTestMode())
		}
		if conf.HologresUsername == "" || conf.HologresPassword == "" {
			client, err = featurestore.NewFeatureStoreClient(conf.RegionId, conf.AccessId, conf.AccessKey, conf.ProjectName,
				options...,
			)
		} else {
			options = append(options, featurestore.WithHologresLogin(conf.HologresUsername, conf.HologresPassword))
			client, err = featurestore.NewFeatureStoreClient(conf.RegionId, conf.AccessId, conf.AccessKey, conf.ProjectName,
				options...,
			)
		}

		if err != nil {
			panic(err)
		}

		p, err := client.GetProject(conf.ProjectName)

		if err != nil {
			panic(err)
		}

		m := &FSClient{
			client:      client,
			project:     p,
			projectName: conf.ProjectName,
		}
		fsInstances[name] = m
	}
}
