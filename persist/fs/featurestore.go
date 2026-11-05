package fs

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/featurestore"
)

var (
	fsInstancesMu sync.RWMutex
	fsLoadMu      sync.Mutex
	fsInstances   = make(map[string]*FSClient)
)

func GetFeatureStoreClient(name string) (*FSClient, error) {
	fsInstancesMu.RLock()
	defer fsInstancesMu.RUnlock()
	if client, ok := fsInstances[name]; ok {
		return client, nil
	}
	return nil, fmt.Errorf("feature store client not found, name:%s", name)
}

type FSClient struct {
	client      *featurestore.FeatureStoreClient
	projectMu   sync.RWMutex
	project     *domain.Project
	projectName string
}

func (fs *FSClient) GetProject() *domain.Project {
	fs.projectMu.RLock()
	defer fs.projectMu.RUnlock()
	return fs.project
}

func (fs *FSClient) GetLLMConfig(name string) (*domain.LLMConfig, error) {
	return fs.client.GetLLMConfig(name)
}

func (fs *FSClient) ReloadProject() {
	if p, err := fs.client.GetProject(fs.projectName); err == nil {
		fs.setProject(p)
	} else {
		log.Error(fmt.Sprintf("get project failed, projectName:%s, err:%v", fs.projectName, err))
	}
}

func (fs *FSClient) setProject(project *domain.Project) {
	fs.projectMu.Lock()
	defer fs.projectMu.Unlock()
	fs.project = project
}

func Load(config *recconf.RecommendConfig) {
	fsLoadMu.Lock()
	defer fsLoadMu.Unlock()

	// The feature store SDK caches connection pools (hologres/featuredb/...) in
	// process-level registries keyed by server-side datasource name with
	// first-registration-wins semantics. This loop iterates the config map in its
	// natural (random) order, so when several FeatureStoreConfs share one project
	// but differ, which config takes effect is non-deterministic across restarts.
	// Warn on that conflict so it can be converged instead of silently picking one.
	seenProjects := make(map[string]string, len(config.FeatureStoreConfs))
	for name, conf := range config.FeatureStoreConfs {
		if firstName, ok := seenProjects[conf.ProjectName]; ok {
			if diffFields := diffFeatureStoreConfs(config.FeatureStoreConfs[firstName], conf); len(diffFields) > 0 {
				log.Warning(fmt.Sprintf("module=FeatureStore\tname=%q\tmsg=project %s is registered by multiple FeatureStoreConfs (%q and %q) with differing fields, the effective config is non-deterministic, please converge them",
					name, conf.ProjectName, firstName, name))
			}
		} else {
			seenProjects[conf.ProjectName] = name
		}

		fsInstancesMu.RLock()
		fs, ok := fsInstances[name]
		fsInstancesMu.RUnlock()
		if ok {
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
		fsInstancesMu.Lock()
		fsInstances[name] = m
		fsInstancesMu.Unlock()
	}
}

// diffFeatureStoreConfs returns the field names that differ between two
// FeatureStoreConfig entries of the same project. Values are never included
// because the configs may contain credentials.
func diffFeatureStoreConfs(first, other recconf.FeatureStoreConfig) []string {
	var diffFields []string
	firstValue := reflect.ValueOf(first)
	otherValue := reflect.ValueOf(other)
	for i := 0; i < firstValue.NumField(); i++ {
		if !reflect.DeepEqual(firstValue.Field(i).Interface(), otherValue.Field(i).Interface()) {
			diffFields = append(diffFields, firstValue.Type().Field(i).Name)
		}
	}
	return diffFields
}
