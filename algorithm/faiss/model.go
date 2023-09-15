package faiss

import (
	"time"

	"github.com/alibaba/pairec/recconf"
)

const ()

type FaissModel struct {
	name   string
	client *VectorClient
}

func NewFaissModel(name string) *FaissModel {
	return &FaissModel{name: name}
}
func (m *FaissModel) Init(conf *recconf.AlgoConfig) error {
	client, err := NewVectorClient(conf.VectorConf.ServerAddress, time.Millisecond*time.Duration(conf.VectorConf.Timeout))
	if err != nil {
		return err
	}

	m.client = client
	return nil

}
func (m *FaissModel) Run(algoData interface{}) (interface{}, error) {
	return m.client.Search(algoData)
}
