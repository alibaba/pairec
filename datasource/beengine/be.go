package beengine

import (
	"fmt"
	"sync"

	be "github.com/aliyun/aliyun-be-go-sdk"
	"github.com/alibaba/pairec/v2/recconf"
)

type BeClient struct {
	BeClient        *be.Client
	productReleased bool
}

var (
	mu          sync.RWMutex
	beInstances = make(map[string]*BeClient)
)

func GetBeClient(name string) (*BeClient, error) {
	mu.RLock()
	defer mu.RUnlock()
	if _, ok := beInstances[name]; !ok {
		return nil, fmt.Errorf("BeClient not found, name:%s", name)
	}

	return beInstances[name], nil
}
func RegisterBeClient(name string, client *BeClient) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := beInstances[name]; !ok {
		beInstances[name] = client
	}
}

func NewBeClient(username, password, endpoint string) *BeClient {
	p := &BeClient{
		productReleased: false,
	}
	p.BeClient = be.NewClient(endpoint, username, password)
	return p
}

func (d *BeClient) Init() error {

	return nil
}

func (d *BeClient) IsProductReleased() bool {

	return d.productReleased
}

func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.BEConfs {
		if _, ok := beInstances[name]; ok {
			continue
		}
		m := NewBeClient(conf.Username, conf.Password, conf.Endpoint)
		if conf.ReleaseType == "product" {
			m.productReleased = true
		}

		err := m.Init()
		if err != nil {
			panic(err)
		}
		RegisterBeClient(name, m)
	}
}
