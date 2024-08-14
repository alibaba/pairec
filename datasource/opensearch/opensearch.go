package opensearch

import (
	"fmt"
	"sync"

	openSearchClient "github.com/alibaba/pairec/v2/datasource/opensearch/client"
	"github.com/alibaba/pairec/v2/recconf"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
)

type OpenSearchClient struct {
	OpenSearchClient *openSearchClient.Client
	Runtime          *util.RuntimeOptions
}

var (
	mu                  sync.RWMutex
	opensearchInstances = make(map[string]*OpenSearchClient)
)

func GetOpenSearchClient(name string) (*OpenSearchClient, error) {
	mu.RLock()
	defer mu.RUnlock()
	if _, ok := opensearchInstances[name]; !ok {
		return nil, fmt.Errorf("opensearchClient not found, name:%s", name)
	}

	return opensearchInstances[name], nil
}
func RegisterOpenSearchClient(name string, client *OpenSearchClient) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := opensearchInstances[name]; !ok {
		opensearchInstances[name] = client
	}
}

func NewOpenSearchClient(endpoint, accessId, accessKey string) *OpenSearchClient {
	p := &OpenSearchClient{}
	config := &openSearchClient.Config{
		Endpoint:        tea.String(endpoint),
		AccessKeyId:     tea.String(accessId),
		AccessKeySecret: tea.String(accessKey),
	}
	client, _clientErr := openSearchClient.NewClient(config)
	if _clientErr != nil {
		panic(_clientErr)
	}

	p.OpenSearchClient = client
	p.Runtime = &util.RuntimeOptions{
		ConnectTimeout: tea.Int(2000),
		ReadTimeout:    tea.Int(1000),
		Autoretry:      tea.Bool(false),
		IgnoreSSL:      tea.Bool(false),
		MaxIdleConns:   tea.Int(50),
	}
	return p
}

func (d *OpenSearchClient) Init() error {

	return nil
}

func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.OpenSearchConfs {
		if _, ok := opensearchInstances[name]; ok {
			continue
		}
		m := NewOpenSearchClient(conf.EndPoint, conf.AccessKeyId, conf.AccessKeySecret)

		err := m.Init()
		if err != nil {
			panic(err)
		}
		RegisterOpenSearchClient(name, m)
	}
}
