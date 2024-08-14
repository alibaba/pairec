package tablestore

import (
	"fmt"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
)

type TableStoreClient struct {
	client *tablestore.TableStoreClient
}

var (
	tablestoreInstances = make(map[string]*TableStoreClient)
)

func (o *TableStoreClient) Init() error {

	return nil
}

func RegisterTableStoreClient(name string, client *tablestore.TableStoreClient) {
	p := &TableStoreClient{}
	if _, ok := tablestoreInstances[name]; !ok {
		p.client = client
		tablestoreInstances[name] = p
	}
}

func GetTableStoreClient(name string) (*TableStoreClient, error) {
	if _, ok := tablestoreInstances[name]; !ok {
		return nil, fmt.Errorf("TableStoreClient not found, name:%s", name)
	}

	return tablestoreInstances[name], nil
}

func (o *TableStoreClient) GetClient() *tablestore.TableStoreClient {
	return o.client
}
