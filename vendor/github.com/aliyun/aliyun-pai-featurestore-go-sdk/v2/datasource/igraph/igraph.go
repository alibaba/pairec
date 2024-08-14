package igraph

import (
	"fmt"
	"strings"

	aligraph "github.com/aliyun/aliyun-igraph-go-sdk"
)

type GraphClient struct {
	GraphClient *aligraph.Client
}

var (
	graphInstances = make(map[string]*GraphClient)
)

func GetGraphClient(name string) (*GraphClient, error) {
	if _, ok := graphInstances[name]; !ok {
		return nil, fmt.Errorf("GraphClient not found, name:%s", name)
	}

	return graphInstances[name], nil
}
func RegisterGraphClient(name string, client *GraphClient) {
	if _, ok := graphInstances[name]; !ok {
		graphInstances[name] = client
	}
}

func NewGraphClient(host, userName, passwd string) *GraphClient {
	p := &GraphClient{}
	if !strings.HasPrefix(host, "http://") {
		host = "http://" + host
	}

	p.GraphClient = aligraph.NewClient(host, userName, passwd, "featurestore-sdk-go")
	return p
}

func (d *GraphClient) Init() error {

	return nil
}
