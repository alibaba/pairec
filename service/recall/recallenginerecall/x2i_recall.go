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

type RecallEngineX2IRecall struct {
	returnCount     int
	bizName         string
	serviceName     string
	recallName      string
	scorerClause    string
	triggerIdName   string
	recallTableName string
	diversityParam  string
	customParams    map[string]interface{}
	triggerKey      TriggerKey
	client          *recallengine.RecallEngineClient
	mu              sync.RWMutex
	cloneInstances  map[string]*RecallEngineX2IRecall
}

func NewRecallEngineX2IRecall(client *recallengine.RecallEngineClient, conf recconf.RecallEngineConfig) *RecallEngineX2IRecall {
	if len(conf.RecallEngineParams) != 1 {
		return nil
	}

	r := RecallEngineX2IRecall{
		serviceName:     conf.ServiceName,
		returnCount:     conf.RecallEngineParams[0].Count,
		scorerClause:    conf.RecallEngineParams[0].ScorerClause,
		recallName:      conf.RecallEngineParams[0].RecallName,
		triggerIdName:   conf.RecallEngineParams[0].TriggerIdName,
		recallTableName: conf.RecallEngineParams[0].RecallTableName,
		diversityParam:  conf.RecallEngineParams[0].DiversityParam,
		customParams:    conf.RecallEngineParams[0].CustomParams,
		triggerKey:      NewTriggerKey(&conf.RecallEngineParams[0], nil),
		client:          client,
		cloneInstances:  make(map[string]*RecallEngineX2IRecall),
	}

	return &r
}
func (r *RecallEngineX2IRecall) GetRecallName() string {
	return r.recallName
}

func (r *RecallEngineX2IRecall) GetItems(user *module.User, context *context.RecommendContext) (ret []*module.Item, err error) {
	return
}

func (r *RecallEngineX2IRecall) BuildQueryParams(user *module.User, context *context.RecommendContext) (ret re.RecallConf) {
	triggerResult := r.triggerKey.GetTriggerKey(user, context)
	if triggerResult.TriggerItem == "" {
		return
	}

	ret.Trigger = triggerResult.TriggerItem
	ret.Count = r.returnCount
	return

	/*
		if _, ok := r.triggerKey.(*berecall.UserRealtimeEmbeddingMindTrigger); ok {
			triggerItems := strings.Split(triggerResult.TriggerItem, "|")
			if r.client.IsProductReleased() {
				var items []string
				for i, trigger := range triggerItems {
					itemIdScores := strings.Split(trigger, ",")
					for _, item := range itemIdScores {
						items = append(items, fmt.Sprintf("%s:%d", item, i))
					}
				}
				ret[fmt.Sprintf("%s_list", r.recallName)] = strings.Join(items, ",")

			} else {
				for i, trigger := range triggerItems {
					ret[fmt.Sprintf("%s_%d_list", r.recallName, i)] = trigger
				}
			}
			ret["mind_embedding_return_count"] = strconv.Itoa(r.returnCount)
			ret[fmt.Sprintf("%s_return_count", r.recallName)] = strconv.Itoa(r.returnCount)
			if triggerResult.DistinctParam != "" && triggerResult.DistinctParamName != "" {
				ret[triggerResult.DistinctParamName] = triggerResult.DistinctParam
			}
		} else if _, ok := r.triggerKey.(*berecall.UserEmbeddingDssmO2OTrigger); ok {
			ret[fmt.Sprintf("%s_qinfo", r.recallName)] = triggerResult.TriggerItem
			ret[fmt.Sprintf("%s_return_count", r.recallName)] = strconv.Itoa(r.returnCount)
		} else {
			ret[fmt.Sprintf("%s_list", r.recallName)] = triggerResult.TriggerItem
			ret[fmt.Sprintf("%s_return_count", r.recallName)] = strconv.Itoa(r.returnCount)
		}
		//ret[fmt.Sprintf("%s_return_count", r.recallName)] = strconv.Itoa(r.returnCount)
		if r.diversityParam != "" {
			ret[fmt.Sprintf("%s_diversity_param", r.recallName)] = r.diversityParam
		} else if r.triggerIdName != "" && triggerResult.DistinctParam != "" {
			ret[fmt.Sprintf("%s_distinct_param", r.recallName)] = fmt.Sprintf("%s:%s", r.triggerIdName, triggerResult.DistinctParam)
		}

		if r.recallTableName != "" {
			ret[fmt.Sprintf("%s_table", r.recallName)] = r.recallTableName
		}

		for k, v := range r.customParams {
			ret[k] = utils.ToString(v, "")
		}
	*/

}
func (r *RecallEngineX2IRecall) CloneWithConfig(params map[string]interface{}) RecallEngineBaseRecall {
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

	recall = &RecallEngineX2IRecall{
		serviceName:     r.serviceName,
		client:          r.client,
		returnCount:     recallParams.Count,
		recallName:      r.recallName,
		triggerIdName:   recallParams.TriggerIdName,
		recallTableName: recallParams.RecallTableName,
		diversityParam:  recallParams.DiversityParam,
		customParams:    recallParams.CustomParams,
		triggerKey:      NewTriggerKey(&recallParams, r.client),
	}

	r.cloneInstances[md5] = recall
	return recall
}
