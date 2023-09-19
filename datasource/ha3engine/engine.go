package ha3engine

import (
	"fmt"
	"sync"

	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/alibaba/pairec/v2/datasource/ha3engine/ha3client"
	"github.com/alibaba/pairec/v2/recconf"
)

type Ha3EngineClient struct {
	Ha3Client *ha3client.Client
	runtime   *util.RuntimeOptions
}

var (
	mu           sync.RWMutex
	ha3Instances = make(map[string]*Ha3EngineClient)
)

func GetHa3EngineClient(name string) (*Ha3EngineClient, error) {
	mu.RLock()
	defer mu.RUnlock()
	if _, ok := ha3Instances[name]; !ok {
		return nil, fmt.Errorf("ha3EngineClient not found, name:%s", name)
	}

	return ha3Instances[name], nil
}
func RegisterHa3EngineClient(name string, client *Ha3EngineClient) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := ha3Instances[name]; !ok {
		ha3Instances[name] = client
	}
}

func NewHa3EngineClient(username, password, endpoint, instanceId string) *Ha3EngineClient {
	p := &Ha3EngineClient{}
	config := &ha3client.Config{
		Endpoint:       tea.String(endpoint),
		InstanceId:     tea.String(instanceId),
		AccessUserName: tea.String(username),
		AccessPassWord: tea.String(password),
	}
	client, _clientErr := ha3client.NewClient(config)
	if _clientErr != nil {
		panic(_clientErr)
	}

	p.Ha3Client = client
	p.runtime = &util.RuntimeOptions{
		ConnectTimeout: tea.Int(5000),
		ReadTimeout:    tea.Int(200),
		Autoretry:      tea.Bool(false),
		IgnoreSSL:      tea.Bool(false),
		MaxIdleConns:   tea.Int(50),
		//HttpProxy:      tea.String("http://116.*.*.187:8088"),
	}
	return p
}

func (d *Ha3EngineClient) Init() error {

	return nil
}

func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.Ha3EngineConfs {
		if _, ok := ha3Instances[name]; ok {
			continue
		}
		m := NewHa3EngineClient(conf.Username, conf.Password, conf.Endpoint, conf.InstanceId)

		err := m.Init()
		if err != nil {
			panic(err)
		}
		RegisterHa3EngineClient(name, m)
	}
}
