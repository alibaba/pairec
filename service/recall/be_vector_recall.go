package recall

import (
	"encoding/json"
	"fmt"
	"strconv"
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

type BeVectorRecall struct {
	returnCount     int
	bizName         string
	recallName      string
	scorerClause    string
	itemIdName      string
	recallTableName string
	diversityParam  string
	triggerKey      berecall.TriggerKey
	beFilterNames   []string
	beClient        *be.Client
	client          *beengine.BeClient
	cloneInstances  map[string]*BeVectorRecall
}

func NewBeVectorRecall(client *beengine.BeClient, conf recconf.BeConfig) *BeVectorRecall {
	if len(conf.BeRecallParams) != 1 {
		return nil
	}

	beClient := client.BeClient
	r := BeVectorRecall{
		bizName:         conf.BizName,
		returnCount:     conf.BeRecallParams[0].Count,
		scorerClause:    conf.BeRecallParams[0].ScorerClause,
		itemIdName:      conf.BeRecallParams[0].ItemIdName,
		recallName:      conf.BeRecallParams[0].RecallName,
		recallTableName: conf.BeRecallParams[0].RecallTableName,
		diversityParam:  conf.BeRecallParams[0].DiversityParam,
		beFilterNames:   conf.BeFilterNames,
		triggerKey:      berecall.NewTriggerKey(&conf.BeRecallParams[0], client),
		beClient:        beClient,
		client:          client,
		cloneInstances:  make(map[string]*BeVectorRecall),
	}

	return &r
}
func (r *BeVectorRecall) GetItems(user *module.User, context *context.RecommendContext) (ret []*module.Item, err error) {

	params := r.BuildQueryParams(user, context)
	params["user_id"] = string(user.Id)

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

	vectorReadRequest := be.NewReadRequest(r.bizName, r.returnCount)
	vectorReadRequest.IsRawRequest = true
	vectorReadRequest.SetQueryParams(params)

	if context.Debug {
		uri := vectorReadRequest.BuildUri()
		log.Info(fmt.Sprintf("requestId=%s\tbizName=%s\turl=%s", context.RecommendId, r.bizName, uri.RequestURI()))
	}

	vectorReadResponse, err := r.beClient.Read(*vectorReadRequest)
	if err != nil {
		return
	}

	matchItems := vectorReadResponse.Result.MatchItems
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

/*
	func (r *BeVectorRecall) BuildRecallParam(user *module.User, context *context.RecommendContext) *be.RecallParam {
		triggerResult := r.triggerKey.GetTriggerKey(user, context)
		if triggerResult.TriggerItem == "" {
			return nil
		}
		vectorRecallParams := be.NewRecallParam().
			SetTriggerItems([]string{triggerResult.TriggerItem}).
			SetRecallType(be.RecallTypeVector)
		vectorRecallParams.ReturnCount = r.returnCount

		if r.scorerClause != "" {
			vectorRecallParams.SetScorerClause(be.NewScorerClause(r.scorerClause))
		}

		if r.recallName != "" {
			vectorRecallParams.SetRecallName(r.recallName)
		}

		return vectorRecallParams
	}
*/
func (r *BeVectorRecall) BuildQueryParams(user *module.User, context *context.RecommendContext) (ret map[string]string) {
	ret = make(map[string]string)
	triggerResult := r.triggerKey.GetTriggerKey(user, context)
	if triggerResult.TriggerItem == "" {
		return
	}

	ret[fmt.Sprintf("%s_list", r.recallName)] = triggerResult.TriggerItem
	ret[fmt.Sprintf("%s_return_count", r.recallName)] = strconv.Itoa(r.returnCount)

	if r.recallTableName != "" {
		ret[fmt.Sprintf("%s_table", r.recallName)] = r.recallTableName
	}

	if r.diversityParam != "" {
		ret[fmt.Sprintf("%s_diversity_param", r.recallName)] = r.diversityParam
	}

	return
}
func (r *BeVectorRecall) CloneWithConfig(params map[string]interface{}) BeBaseRecall {
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

	recall := &BeVectorRecall{
		bizName:         r.bizName,
		beClient:        r.beClient,
		client:          r.client,
		beFilterNames:   r.beFilterNames,
		returnCount:     recallParams.Count,
		itemIdName:      recallParams.ItemIdName,
		recallName:      r.recallName,
		recallTableName: recallParams.RecallTableName,
		diversityParam:  recallParams.DiversityParam,
		triggerKey:      berecall.NewTriggerKey(&recallParams, r.client),
	}

	r.cloneInstances[md5] = recall
	return recall
}
