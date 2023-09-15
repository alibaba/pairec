package fs

import (
	"fmt"

	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/domain"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/featurestore"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/recconf"
)

var fsInstances = make(map[string]*FSClient)

func GetFeatureStoreClient(name string) (*FSClient, error) {
	if _, ok := fsInstances[name]; !ok {
		return nil, fmt.Errorf("feature store client not found, name:%s", name)
	}

	return fsInstances[name], nil
}

type FSClient struct {
	client  *featurestore.FeatureStoreClient
	project *domain.Project
}

func (fs *FSClient) GetProject() *domain.Project {
	return fs.project
}

func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.FeatureStoreConfs {
		if _, ok := fsInstances[name]; ok {
			continue
		}

		l := log.FeatureStoreLogger{}
		client, err := featurestore.NewFeatureStoreClient(conf.Host,
			featurestore.WithLogger(featurestore.LoggerFunc(l.Infof)),
			featurestore.WithErrorLogger(featurestore.LoggerFunc(l.Errorf)),
			featurestore.WithToken(conf.Token),
		)

		if err != nil {
			panic(err)
		}

		p, err := client.GetProject(conf.ProjectName)

		if err != nil {
			panic(err)
		}

		m := &FSClient{
			client:  client,
			project: p,
		}
		fsInstances[name] = m
	}
}
