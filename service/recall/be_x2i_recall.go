package recall

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	be "github.com/aliyun/aliyun-be-go-sdk"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/beengine"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/recall/berecall"
	"github.com/alibaba/pairec/v2/utils"
)

type BeX2IRecall struct {
	returnCount     int
	bizName         string
	recallName      string
	scorerClause    string
	itemIdName      string
	triggerIdName   string
	recallTableName string
	diversityParam  string
	customParams    map[string]interface{}
	triggerKey      berecall.TriggerKey
	beFilterNames   []string
	beClient        *be.Client
	client          *beengine.BeClient
	cloneInstances  map[string]*BeX2IRecall
}

func NewBeX2IRecall(client *beengine.BeClient, conf recconf.BeConfig) *BeX2IRecall {
	if len(conf.BeRecallParams) != 1 {
		return nil
	}

	beClient := client.BeClient
	r := BeX2IRecall{
		bizName:         conf.BizName,
		returnCount:     conf.BeRecallParams[0].Count,
		scorerClause:    conf.BeRecallParams[0].ScorerClause,
		itemIdName:      conf.BeRecallParams[0].ItemIdName,
		recallName:      conf.BeRecallParams[0].RecallName,
		triggerIdName:   conf.BeRecallParams[0].TriggerIdName,
		recallTableName: conf.BeRecallParams[0].RecallTableName,
		diversityParam:  conf.BeRecallParams[0].DiversityParam,
		customParams:    conf.BeRecallParams[0].CustomParams,
		beFilterNames:   conf.BeFilterNames,
		triggerKey:      berecall.NewTriggerKey(&conf.BeRecallParams[0], client),
		beClient:        beClient,
		client:          client,
		cloneInstances:  make(map[string]*BeX2IRecall),
	}

	return &r
}
func (r *BeX2IRecall) buildRequest(user *module.User, context *context.RecommendContext) *be.ReadRequest {
	x2iReadRequest := be.NewReadRequest(r.bizName, r.returnCount)
	x2iReadRequest.IsRawRequest = true
	params := r.BuildQueryParams(user, context)
	params["user_id"] = string(user.Id)

	// trigger_list
	triggerKey := fmt.Sprintf("%s_list", r.recallName)
	triggerValues, _ := params[triggerKey]
	params["trigger_list"] = triggerValues
	delete(params, triggerKey)

	// return_count
	countKey := fmt.Sprintf("%s_return_count", r.recallName)
	countValue, _ := params[countKey]
	params["return_count"] = countValue
	delete(params, countKey)

	if len(r.beFilterNames) > 0 {
		if len(r.beFilterNames) == 1 {
			if filter, err := berecall.GetFilter(r.beFilterNames[0]); err == nil {
				filterParams := filter.BuildQueryParams(user, context)
				for k, v := range filterParams {
					params[k] = v
				}
			}
		} else {
			var wg sync.WaitGroup
			var mu sync.Mutex
			for _, name := range r.beFilterNames {
				if filter, err := berecall.GetFilter(name); err == nil {
					wg.Add(1)
					go func(filer berecall.IBeFilter) {
						defer wg.Done()
						filterParams := filter.BuildQueryParams(user, context)
						mu.Lock()
						defer mu.Unlock()
						for k, v := range filterParams {
							params[k] = v
						}

					}(filter)
				}
			}
			wg.Wait()
		}
	}

	x2iReadRequest.SetQueryParams(params)

	if context.Debug {
		uri := x2iReadRequest.BuildUri()
		log.Info(fmt.Sprintf("requestId=%s\tbizName=%s\turl=%s", context.RecommendId, r.bizName, uri.RequestURI()))
	}

	return x2iReadRequest
}
func (r *BeX2IRecall) GetItems(user *module.User, context *context.RecommendContext) (ret []*module.Item, err error) {
	x2iReadRequest := r.buildRequest(user, context)
	x2iReadResponse, err := r.beClient.Read(*x2iReadRequest)
	if err != nil {
		uri := x2iReadRequest.BuildUri()
		log.Error(fmt.Sprintf("requestId=%s\tbizName=%s\turl=%s", context.RecommendId, r.bizName, uri.RequestURI()))
		return
	}

	matchItems := x2iReadResponse.Result.MatchItems
	if matchItems == nil || len(matchItems.FieldValues) == 0 {
		return
	}

	itemIndex := -1
	scoreIndex := -1
	for i, name := range matchItems.FieldNames {
		if name == r.itemIdName {
			itemIndex = i
		}
		if name == "__score__" {
			scoreIndex = i
		}

		if itemIndex != -1 && scoreIndex != -1 {
			break
		}
	}

	if itemIndex >= 0 && scoreIndex >= 0 {
		var (
			itemId string
			score  float64
		)

		for _, values := range matchItems.FieldValues {
			properties := make(map[string]interface{})

			for i, value := range values {
				if i == itemIndex {
					itemId = utils.ToString(value, "")
				} else if i == scoreIndex {
					score = value.(float64)
				} else {
					properties[matchItems.FieldNames[i]] = value
				}
			}

			item := module.NewItem(itemId)
			item.Score = score
			item.AddProperties(properties)
			ret = append(ret, item)
		}
	}
	return
}

func (r *BeX2IRecall) BuildQueryParams(user *module.User, context *context.RecommendContext) (ret map[string]string) {
	ret = make(map[string]string)
	triggerResult := r.triggerKey.GetTriggerKey(user, context)
	if triggerResult.TriggerItem == "" {
		return
	}

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

	return
}
func (r *BeX2IRecall) CloneWithConfig(params map[string]interface{}) BeBaseRecall {
	j, err := json.Marshal(params)
	if err != nil {
		log.Error(fmt.Sprintf("event=CloneWithConfig\terror=%v", err))
		return r
	}

	recallParams := recconf.BeRecallParam{}
	if err := json.Unmarshal(j, &recallParams); err != nil {
		log.Error(fmt.Sprintf("event=CloneWithConfig\terror=%v", err))
		return r
	}

	d, _ := json.Marshal(recallParams)
	md5 := utils.Md5(string(d))
	if recall, ok := r.cloneInstances[md5]; ok {
		return recall
	}

	recall := &BeX2IRecall{
		bizName:         r.bizName,
		beClient:        r.beClient,
		client:          r.client,
		beFilterNames:   r.beFilterNames,
		returnCount:     recallParams.Count,
		itemIdName:      recallParams.ItemIdName,
		recallName:      r.recallName,
		triggerIdName:   recallParams.TriggerIdName,
		recallTableName: recallParams.RecallTableName,
		diversityParam:  recallParams.DiversityParam,
		customParams:    recallParams.CustomParams,
		triggerKey:      berecall.NewTriggerKey(&recallParams, r.client),
	}

	r.cloneInstances[md5] = recall
	return recall
}
