package hbase_thrift

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/alibaba/pairec/datasource/hbase_thrift/gen-go/hbase"
	"github.com/alibaba/pairec/recconf"
)

type HBaseThrift struct {
	Client   *hbase.THBaseServiceClient
	Host     string
	User     string
	Password string
}
type HBaseThriftPool struct {
	mu      sync.Mutex
	clients []*HBaseThrift
	client  *HBaseThrift
}

var hbaseThriftInstances = make(map[string]*HBaseThriftPool)

func GetHBaseThrift(name string) (*HBaseThrift, error) {
	pool, ok := hbaseThriftInstances[name]
	if !ok {
		return nil, fmt.Errorf("hbase not found, name:%s", name)
	}

	pool.mu.Lock()
	defer pool.mu.Unlock()
	var client *HBaseThrift
	var err error
	if len(pool.clients) > 0 {
		client = pool.clients[0]
		pool.clients = pool.clients[1:]
	} else {
		client = &HBaseThrift{
			User:     pool.client.User,
			Password: pool.client.Password,
			Host:     pool.client.Host,
		}
		err = client.Init()
	}

	return client, err
}
func PutHBaseThrift(name string, client *HBaseThrift) {
	pool, ok := hbaseThriftInstances[name]
	if !ok {
		return
	}

	pool.mu.Lock()
	pool.clients = append(pool.clients, client)
	pool.mu.Unlock()
}
func (h *HBaseThrift) Init() error {
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   1000 * time.Millisecond, // 1000ms
			KeepAlive: 5 * time.Minute,
		}).DialContext,
		MaxIdleConnsPerHost:   100,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
	}

	thriftHttpClient := &http.Client{Transport: tr}
	trans, err := thrift.NewTHttpClientWithOptions(h.Host, thrift.THttpClientOptions{Client: thriftHttpClient})
	if err != nil {
		return err
	}
	// 设置用户名密码
	httClient := trans.(*thrift.THttpClient)
	httClient.SetHeader("ACCESSKEYID", h.User)
	httClient.SetHeader("ACCESSSIGNATURE", h.Password)
	client := hbase.NewTHBaseServiceClientFactory(trans, protocolFactory)
	if err := trans.Open(); err != nil {
		return err
	}
	h.Client = client
	return nil
}

func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.HBaseThriftConfs {
		if _, ok := hbaseThriftInstances[name]; ok {
			continue
		}
		pool := &HBaseThriftPool{client: &HBaseThrift{
			User:     conf.User,
			Password: conf.Password,
			Host:     conf.Host,
		}}
		for i := 0; i < 10; i++ {
			client := &HBaseThrift{
				User:     conf.User,
				Password: conf.Password,
				Host:     conf.Host,
			}
			if err := client.Init(); err != nil {
				panic(err)
			}
			pool.clients = append(pool.clients, client)

		}
		hbaseThriftInstances[name] = pool
	}
}
