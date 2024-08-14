package module

import (
	"fmt"

	be "github.com/aliyun/aliyun-be-go-sdk"
	"github.com/alibaba/pairec/v2/datasource/beengine"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
)

type VectorBeDao struct {
	beClient       *be.Client
	bizName        string
	embeddingField string
	keyField       string
}

func NewVectorBeDao(config recconf.RecallConfig) *VectorBeDao {
	client, err := beengine.GetBeClient(config.VectorDaoConf.BeName)
	if err != nil {
		log.Error(fmt.Sprintf("get beclient error:%v", err))
		return nil
	}

	dao := &VectorBeDao{
		beClient:       client.BeClient,
		bizName:        config.VectorDaoConf.BizName,
		embeddingField: config.VectorDaoConf.EmbeddingField,
		keyField:       config.VectorDaoConf.KeyField,
	}

	return dao
}

func (d *VectorBeDao) VectorString(id string) (string, error) {

	x2iReadRequest := be.NewReadRequest(d.bizName, 1)
	x2iRecallParams := be.NewRecallParam().
		SetTriggerItems([]string{id}).
		SetRecallType(be.RecallTypeX2I)
	x2iReadRequest.AddRecallParam(x2iRecallParams)

	x2iReadResponse, err := d.beClient.Read(*x2iReadRequest)
	if err != nil {
		return "", err
	}

	matchItems := x2iReadResponse.Result.MatchItems
	if matchItems == nil || len(matchItems.FieldValues) == 0 {
		return "", VectoryEmptyError
	}

	embedding := matchItems.FieldValues[0][0].(string)

	if embedding == "" {
		return embedding, VectoryEmptyError
	}

	return embedding, nil
}
