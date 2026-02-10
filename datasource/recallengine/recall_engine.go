package recallengine

import (
	"fmt"
	"os"
	"sync"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	re "github.com/aliyun/aliyun-pairec-config-go-sdk/v2/recallengine"
)

type RecallEngineClient struct {
	client     *re.Client
	instanceId string
}

func (r RecallEngineClient) GetRecallEngineClient() *re.Client {
	return r.client
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

func NewRecallEngineClient(username, password, endpoint, authorization string) *RecallEngineClient {
	p := &RecallEngineClient{
		//productReleased: false,
	}
	logger := log.RecallEngineLogger{}
	if endpoint != "" {
		var client *re.Client
		if authorization == "" {
			client = re.NewClient(endpoint, username, password, re.WithRetryTimes(2), re.WithLogger(logger))
		} else {
			client = re.NewClient(endpoint, username, password, re.WithRetryTimes(2), re.WithLogger(logger),
				re.WithRequestHeader("Authorization", authorization))
		}
		p.client = client
	} else {
		region := os.Getenv("REGION")
		instanceId := os.Getenv("INSTANCE_ID")
		accessId := os.Getenv("AccessKey")
		accessSecret := os.Getenv("AccessSecret")
		if region == "" {
			panic("env REGION empty")
		}
		if instanceId == "" {
			panic("env INSTANCE_ID empty")
		}

		endpoint, err := re.GetRecallEngineEndpoint(instanceId, region, &re.GetRecallEngineEndpointOption{
			AccessKeyId:     accessId,
			AccessKeySecret: accessSecret,
		})
		if err != nil {
			panic(err)
		}
		client := re.NewClient(endpoint, username, password, re.WithRetryTimes(2), re.WithLogger(logger), re.WithEndpointSchema("http"))
		p.client = client
		p.instanceId = instanceId
	}

	return p
}

func (d *RecallEngineClient) Init() error {

	return nil
}

func (d *RecallEngineClient) InstanceId() string {
	return d.instanceId
}

func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.RecallEngineConfs {
		if _, ok := reInstances[name]; ok {
			continue
		}
		m := NewRecallEngineClient(conf.Username, conf.Password, conf.Endpoint, conf.Authorization)

		err := m.Init()
		if err != nil {
			panic(err)
		}
		if conf.InstanceId != "" {
			m.instanceId = conf.InstanceId
		}

		RegisterRecallEngineClient(name, m)
	}
}
