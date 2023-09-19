package graph

import (
	"fmt"
	igraph "github.com/aliyun/aliyun-igraph-go-sdk"
	"github.com/alibaba/pairec/v2/recconf"
	"sync"
)

type GraphClient struct {
	GraphClient *igraph.Client
}

var (
	mu             sync.RWMutex
	graphInstances = make(map[string]*GraphClient)
)

func GetGraphClient(name string) (*GraphClient, error) {
	mu.RLock()
	defer mu.RUnlock()
	if _, ok := graphInstances[name]; !ok {
		return nil, fmt.Errorf("GraphClient not found, name:%s", name)
	}

	return graphInstances[name], nil
}
func RegisterGraphClient(name string, client *GraphClient) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := graphInstances[name]; !ok {
		graphInstances[name] = client
	}
}

func NewGraphClient(host, userName, passwd string) *GraphClient {
	p := &GraphClient{}
	p.GraphClient = igraph.NewClient(host, userName, passwd, "pairec")
	return p
}

func (d *GraphClient) Init() error {

	return nil
}

func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.GraphConfs {
		if _, ok := graphInstances[name]; ok {
			continue
		}
		m := NewGraphClient(conf.Host, conf.UserName, conf.Passwd)

		err := m.Init()
		if err != nil {
			panic(err)
		}
		RegisterGraphClient(name, m)
	}
}
