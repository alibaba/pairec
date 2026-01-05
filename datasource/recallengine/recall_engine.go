package recallengine

import (
	"fmt"
	"sync"

	"github.com/alibaba/pairec/v2/recconf"
)

type RecallEngineClient struct {
}

var (
	mu          sync.RWMutex
	reInstances = make(map[string]*RecallEngineClient)
)

func GetRecallEngineClient(name string) (*RecallEngineClient, error) {
	mu.RLock()
	defer mu.RUnlock()
	if _, ok := reInstances[name]; !ok {
		return nil, fmt.Errorf("RecallEngineClient not found, name:%s", name)
	}

	return reInstances[name], nil
}
func RegisterRecallEngineClient(name string, client *RecallEngineClient) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := reInstances[name]; !ok {
		reInstances[name] = client
	}
}

func NewRecallEngineClient(username, password string) *RecallEngineClient {
	p := &RecallEngineClient{
		//productReleased: false,
	}
	//p.RecallEngineClient = be.NewClient(endpoint, username, password)
	return p
}

func (d *RecallEngineClient) Init() error {

	return nil
}

/*
func (d *RecallEngineClient) IsProductReleased() bool {

	return d.productReleased
}
*/

func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.RecallEngineConfs {
		if _, ok := reInstances[name]; ok {
			continue
		}
		m := NewRecallEngineClient(conf.Username, conf.Password)
		/*
			if conf.ReleaseType == "product" {
				m.productReleased = true
			}
		*/

		err := m.Init()
		if err != nil {
			panic(err)
		}
		RegisterRecallEngineClient(name, m)
	}
}
