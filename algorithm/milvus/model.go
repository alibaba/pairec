package milvus

import (
	"context"
	"errors"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/alibaba/pairec/v2/recconf"
)

type MilvusModel struct {
	name   string
	client client.Client
}

func NewMilvusModel(name string) *MilvusModel {
	return &MilvusModel{name: name}
}
func (m *MilvusModel) Init(conf *recconf.AlgoConfig) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*time.Duration(conf.MilvusConf.Timeout))
	defer cancel()
	client, err := client.NewGrpcClient(ctx, conf.MilvusConf.ServerAddress)
	if err != nil {
		return err
	}

	m.client = client
	return nil

}
func (m *MilvusModel) Run(algoData interface{}) (response interface{}, err error) {
	request, ok := algoData.(*MilvusRequest)
	if !ok {
		err = errors.New("requestData is not MilvusRequest type")
		return
	}

	return m.client.Search(context.Background(), request.CollectionName, request.Partitions, request.Expr,
		request.OutputFields, request.Vectors, request.VectorField, request.MetricType, request.TopK, request.SearchParams)
}
