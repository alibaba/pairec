package fs

import (
	"fmt"
	"reflect"
	"sort"
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
	// first-registration-wins semantics. Iterating the config map in random
	// order would make the winning config of a same-project registration
	// non-deterministic across restarts, so configs are loaded in sorted key
	// order to make the selection stable and reproducible.
	names := make([]string, 0, len(config.FeatureStoreConfs))
	for name := range config.FeatureStoreConfs {
		names = append(names, name)
	}
	sort.Strings(names)

	// firstConfs records, per project, the config that registers the shared
	// connection pools first (and therefore actually takes effect).
	firstConfs := make(map[string]string, len(names))
	for _, name := range names {
		conf := config.FeatureStoreConfs[name]
		if firstName, ok := firstConfs[conf.ProjectName]; ok {
			if diffFields := diffFeatureStoreConfs(config.FeatureStoreConfs[firstName], conf); len(diffFields) > 0 {
				log.Warning(fmt.Sprintf("module=FeatureStore\tname=%q\tmsg=project %s is registered by multiple FeatureStoreConfs, the config of %q is used",
					name, conf.ProjectName, firstName))
			}
		} else {
			firstConfs[conf.ProjectName] = name
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
