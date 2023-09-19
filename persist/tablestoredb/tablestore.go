package tablestoredb

import (
	"fmt"
	"time"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/alibaba/pairec/v2/recconf"
)

type TableStore struct {
	EndPoint        string
	InstanceName    string
	accessKeyId     string
	accessKeySecret string
	RoleArn         string
	Client          *tablestore.TableStoreClient
	Expiration      time.Time
}

var tablestoreInstances = make(map[string]*TableStore)

func GetTableStore(name string) (*TableStore, error) {
	if _, ok := tablestoreInstances[name]; !ok {
		return nil, fmt.Errorf("TableStore not found, name:%s", name)
	}

	return tablestoreInstances[name], nil
}
func (m *TableStore) SetAccessKeyId(id string) {
	m.accessKeyId = id
}

func (m *TableStore) SetAccessKeySecret(secret string) {
	m.accessKeySecret = secret
}

func (m *TableStore) Init() error {
	akId := m.accessKeyId
	akSecret := m.accessKeySecret
	securityToken := ""

	client := tablestore.NewClientWithConfig(m.EndPoint, m.InstanceName, akId, akSecret, securityToken, nil)
	m.Client = client
	return nil
}

func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.TableStoreConfs {
		if _, ok := tablestoreInstances[name]; ok {
			continue
		}
		t := &TableStore{
			EndPoint:        conf.EndPoint,
			InstanceName:    conf.InstanceName,
			accessKeyId:     conf.AccessKeyId,
			accessKeySecret: conf.AccessKeySecret,
			RoleArn:         conf.RoleArn,
		}
		err := t.Init()
		if err != nil {
			panic(err)
		}
		tablestoreInstances[name] = t
	}
}
