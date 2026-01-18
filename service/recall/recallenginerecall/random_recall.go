package recallenginerecall

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/recallengine"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	re "github.com/aliyun/aliyun-pairec-config-go-sdk/v2/recallengine"
)

type RecallEngineRandomRecall struct {
	returnCount     int
	bizName         string
	serviceName     string
	recallName      string
	scorerClause    string
	recallTableName string
	diversityParam  string
	customParams    map[string]interface{}
	beFilterNames   []string
	client          *recallengine.RecallEngineClient
	mu              sync.RWMutex
	cloneInstances  map[string]*RecallEngineRandomRecall
}

func NewRecallEngineRandomRecall(client *recallengine.RecallEngineClient, conf recconf.RecallEngineConfig) *RecallEngineRandomRecall {
	if len(conf.RecallEngineParams) != 1 {
		return nil
	}

	//beClient := client.BeClient
	r := RecallEngineRandomRecall{
		serviceName:     conf.ServiceName,
		returnCount:     conf.RecallEngineParams[0].Count,
		scorerClause:    conf.RecallEngineParams[0].ScorerClause,
		recallName:      conf.RecallEngineParams[0].RecallName,
		recallTableName: conf.RecallEngineParams[0].RecallTableName,
		diversityParam:  conf.RecallEngineParams[0].DiversityParam,
		customParams:    conf.RecallEngineParams[0].CustomParams,
		beFilterNames:   conf.BeFilterNames,
		client:          client,
		cloneInstances:  make(map[string]*RecallEngineRandomRecall),
	}

	return &r
}
func (r *RecallEngineRandomRecall) GetRecallName() string {
	return r.recallName
}

func (r *RecallEngineRandomRecall) GetItems(user *module.User, context *context.RecommendContext) (ret []*module.Item, err error) {
	return
}

func (r *RecallEngineRandomRecall) BuildQueryParams(user *module.User, context *context.RecommendContext) (ret re.RecallConf) {
	ret.Count = r.returnCount
	return

}
func (r *RecallEngineRandomRecall) CloneWithConfig(params map[string]interface{}) RecallEngineBaseRecall {
	j, err := json.Marshal(params)
	if err != nil {
		log.Error(fmt.Sprintf("event=CloneWithConfig\terror=%v", err))
		return r
	}

	recallParams := recconf.RecallEngineParam{}
	if err := json.Unmarshal(j, &recallParams); err != nil {
		log.Error(fmt.Sprintf("event=CloneWithConfig\terror=%v", err))
		return r
	}

	d, _ := json.Marshal(recallParams)
	md5 := utils.Md5(string(d))
	r.mu.RLock()
	recall, ok := r.cloneInstances[md5]
	if ok {
		r.mu.RUnlock()
		return recall
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if recall, ok := r.cloneInstances[md5]; ok {
		return recall
	}

	recall = &RecallEngineRandomRecall{
		serviceName:     r.serviceName,
		client:          r.client,
		beFilterNames:   r.beFilterNames,
		returnCount:     recallParams.Count,
		recallName:      r.recallName,
		recallTableName: recallParams.RecallTableName,
		diversityParam:  recallParams.DiversityParam,
		customParams:    recallParams.CustomParams,
	}

	r.cloneInstances[md5] = recall
	return recall
}
