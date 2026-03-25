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

type RecallEngineVectorRecall struct {
	returnCount  int
	bizName      string
	recallName   string
	serviceName  string
	scorerClause string
	//itemIdName      string
	//recallTableName string
	diversityParam string
	triggerKey     TriggerKey
	client         *recallengine.RecallEngineClient
	mu             sync.RWMutex
	cloneInstances map[string]*RecallEngineVectorRecall
}

func NewRecallEngineVectorRecall(client *recallengine.RecallEngineClient, conf recconf.RecallEngineConfig) *RecallEngineVectorRecall {
	if len(conf.RecallEngineParams) != 1 {
		return nil
	}

	r := RecallEngineVectorRecall{
		serviceName:  conf.ServiceName,
		returnCount:  conf.RecallEngineParams[0].Count,
		scorerClause: conf.RecallEngineParams[0].ScorerClause,
		recallName:   conf.RecallEngineParams[0].RecallName,
		//recallTableName: conf.RecallEngineParams[0].RecallTableName,
		diversityParam: conf.RecallEngineParams[0].DiversityParam,
		triggerKey:     NewTriggerKey(&conf.RecallEngineParams[0], nil),
		client:         client,
		cloneInstances: make(map[string]*RecallEngineVectorRecall),
	}

	return &r
}
func (r *RecallEngineVectorRecall) GetRecallName() string {
	return r.recallName
}

func (r *RecallEngineVectorRecall) GetItems(user *module.User, context *context.RecommendContext) (ret []*module.Item, err error) {

	return
}

func (r *RecallEngineVectorRecall) BuildQueryParams(user *module.User, context *context.RecommendContext) (ret re.RecallConf) {
	triggerResult := r.triggerKey.GetTriggerKey(user, context)
	if triggerResult.TriggerItem == "" {
		return
	}
	ret.Trigger = triggerResult.TriggerItem
	ret.Count = r.returnCount
	return

}
func (r *RecallEngineVectorRecall) CloneWithConfig(params map[string]interface{}) RecallEngineBaseRecall {
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

	recall = &RecallEngineVectorRecall{
		serviceName:    r.serviceName,
		client:         r.client,
		returnCount:    recallParams.Count,
		recallName:     r.recallName,
		diversityParam: recallParams.DiversityParam,
		triggerKey:     NewTriggerKey(&recallParams, r.client),
	}

	r.cloneInstances[md5] = recall
	return recall
}
