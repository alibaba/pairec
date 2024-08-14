package recall

import (
	"fmt"
	"strings"
	"sync"
	"time"

	be "github.com/aliyun/aliyun-be-go-sdk"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/beengine"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/recall/berecall"
	"github.com/alibaba/pairec/v2/utils"
)

type BeMultiBizRecall struct {
	returnCount       int
	modelName         string
	bizName           string
	itemIdName        string
	beClient          *be.Client
	client            *beengine.BeClient
	recallMap         map[string]BeBaseRecall
	beFilterNames     []string
	beABParams        map[string]interface{}
	recallNameMapping map[string]recconf.RecallNameMappingConfig
}

func NewBeMultiBizRecall(client *beengine.BeClient, conf recconf.BeConfig, modelName string) *BeMultiBizRecall {
	if len(conf.BeRecallParams) < 1 {
		return nil
	}

	beClient := client.BeClient
	r := BeMultiBizRecall{
		returnCount:       conf.Count,
		modelName:         modelName,
		bizName:           conf.BizName,
		beClient:          beClient,
		client:            client,
		itemIdName:        conf.BeRecallParams[0].ItemIdName,
		beFilterNames:     conf.BeFilterNames,
		beABParams:        conf.BeABParams,
		recallMap:         make(map[string]BeBaseRecall, 8),
		recallNameMapping: make(map[string]recconf.RecallNameMappingConfig),
	}
	for name, config := range conf.RecallNameMapping {
		recallNameMappingConfig := recconf.RecallNameMappingConfig{
			Format: config.Format,
			Fields: make([]string, len(config.Fields)),
		}
		copy(recallNameMappingConfig.Fields, config.Fields)

		r.recallNameMapping[name] = recallNameMappingConfig
	}

	for _, param := range conf.BeRecallParams {
		switch param.RecallType {
		case recconf.BE_RecallType_X2I:
			recall := &BeX2IRecall{
				recallName:      param.RecallName,
				returnCount:     param.Count,
				scorerClause:    param.ScorerClause,
				itemIdName:      param.ItemIdName,
				triggerIdName:   param.TriggerIdName,
				recallTableName: param.RecallTableName,
				diversityParam:  param.DiversityParam,
				customParams:    param.CustomParams,
				triggerKey:      berecall.NewTriggerKey(&param, client),
				beClient:        beClient,
				client:          client,
				cloneInstances:  make(map[string]*BeX2IRecall),
			}

			r.recallMap[param.RecallName] = recall
		case recconf.BE_RecallType_Vector:
			recall := &BeVectorRecall{
				recallName:      param.RecallName,
				returnCount:     param.Count,
				scorerClause:    param.ScorerClause,
				itemIdName:      param.ItemIdName,
				recallTableName: param.RecallTableName,
				diversityParam:  param.DiversityParam,
				triggerKey:      berecall.NewTriggerKey(&param, client),
				beClient:        beClient,
				client:          client,
				cloneInstances:  make(map[string]*BeVectorRecall),
			}

			r.recallMap[param.RecallName] = recall
		}
	}

	return &r
}
func (r *BeMultiBizRecall) getRecalls(user *module.User, context *context.RecommendContext) (recalls []BeBaseRecall) {
	recallMap := make(map[string]BeBaseRecall, len(r.recallMap))
	for k, v := range r.recallMap {
		recallMap[k] = v
	}
	if context.ExperimentResult != nil {
		includeRecalls := context.ExperimentResult.GetExperimentParams().Get(fmt.Sprintf("recall.%s.includeRecalls", r.modelName), nil)
		if includeRecalls != nil {
			if includeRecallNames, ok := includeRecalls.([]interface{}); ok {
				found := false
				for recallName := range recallMap {
					found = false
					for _, name := range includeRecallNames {
						if recallName == name.(string) {
							found = true
							break
						}
					}
					if !found {
						recallMap[recallName] = nil
						if _, exist := r.recallMap[recallName]; !exist {
							log.Error(fmt.Sprintf("requestId=%s\trecall_name=%s\terror=recall name not found config", context.RecommendId, recallName))
						}
					}
				}
			}
		}

		excludeRecalls := context.ExperimentResult.GetExperimentParams().Get(fmt.Sprintf("recall.%s.excludeRecalls", r.modelName), nil)
		if excludeRecalls != nil {
			if excludeRecallNames, ok := excludeRecalls.([]interface{}); ok {
				for _, name := range excludeRecallNames {
					if _, ok := recallMap[name.(string)]; ok {
						recallMap[name.(string)] = nil
					}
				}
			}
		}
		for name, recall := range recallMap {
			if recall != nil {
				recallConfig := context.ExperimentResult.GetExperimentParams().Get(fmt.Sprintf("recall.%s.%s", r.modelName, name), nil)
				if recallConfig != nil {
					if recallConfigMap, ok := recallConfig.(map[string]interface{}); ok {
						recallMap[name] = recall.CloneWithConfig(recallConfigMap)
					}
				}
			}
		}
	}

	var recallNames []string
	for name, recall := range recallMap {
		if recall != nil {
			recallNames = append(recallNames, name)
			recalls = append(recalls, recall)
		}
	}

	log.Info(fmt.Sprintf("requestId=%s\tbizName=%s\trecall_names=%s", context.RecommendId, r.bizName, strings.Join(recallNames, ",")))
	return
}
func (r *BeMultiBizRecall) buildRequest(user *module.User, context *context.RecommendContext) *be.ReadRequest {
	multiReadRequest := be.NewReadRequest(r.bizName, r.returnCount)
	multiReadRequest.IsRawRequest = true
	multiReadRequest.IsPost = true
	params := make(map[string]string, 16)
	params["user_id"] = string(user.Id)
	var wg sync.WaitGroup
	var mu sync.Mutex
	beABParams := r.beABParams
	if context.ExperimentResult != nil {
		params := context.ExperimentResult.GetExperimentParams().Get(fmt.Sprintf("recall.%s.beABParams", r.modelName), nil)
		if params != nil {
			if abparams, ok := params.(map[string]interface{}); ok {
				beABParams = abparams
			}
		}
	}
	recalls := r.getRecalls(user, context)
	for _, recall := range recalls {
		wg.Add(1)
		go func(beRecall BeBaseRecall) {
			defer wg.Done()
			recallParams := beRecall.BuildQueryParams(user, context)
			mu.Lock()
			defer mu.Unlock()
			for k, v := range recallParams {
				params[k] = v
			}

		}(recall)
	}

	if len(r.beFilterNames) > 0 {
		for _, name := range r.beFilterNames {
			if filter, err := berecall.GetFilter(name); err == nil {
				wg.Add(1)
				go func(filer berecall.IBeFilter) {
					defer wg.Done()
					filterParams := filter.BuildQueryParams(user, context)
					mu.Lock()
					defer mu.Unlock()
					for k, v := range filterParams {
						if r.client.IsProductReleased() {
							params[k] = strings.ReplaceAll(v, ",", "|")
						} else {
							params[k] = v
						}
					}

				}(filter)
			}
		}
	}
	wg.Wait()

	for k, v := range beABParams {
		params[k] = utils.ToString(v, "")
	}

	multiReadRequest.SetQueryParams(params)
	if context.Debug {
		uri := multiReadRequest.BuildParams()
		log.Info(fmt.Sprintf("requestId=%s\tbizName=%s\turl=%s", context.RecommendId, r.bizName, uri))
	}
	return multiReadRequest
}

func (r *BeMultiBizRecall) GetItems(user *module.User, context *context.RecommendContext) (ret []*module.Item, err error) {
	multiReadRequest := r.buildRequest(user, context)

	start := time.Now()
	multiReadResponse, err := r.beClient.Read(*multiReadRequest)
	log.Info(fmt.Sprintf("requestId=%s\tbizName=%s\tcost=%d", context.RecommendId, r.bizName, utils.CostTime(start)))
	if err != nil {
		uri := multiReadRequest.BuildParams()
		log.Error(fmt.Sprintf("requestId=%s\tbizName=%s\turl=%serror=%s", context.RecommendId, r.bizName, uri, err.Error()))
		return
	}

	matchItems := multiReadResponse.Result.MatchItems
	if matchItems == nil || len(matchItems.FieldValues) == 0 {
		return
	}

	fieldNames := matchItems.FieldNames
	var (
		itemId string
		score  float64
		//matchType int
		recallName string
	)
	for _, values := range matchItems.FieldValues {
		properties := make(map[string]interface{})

		for i, value := range values {
			if fieldNames[i] == r.itemIdName {
				itemId = utils.ToString(value, "")
			} else if fieldNames[i] == beScoreFieldName {
				score = value.(float64)
			} else if fieldNames[i] == beMatchTypeFieldName {
				continue
			} else if fieldNames[i] == beRecallName {
				recallName = value.(string)
			} else if fieldNames[i] == beRecallNameV2 {
				recallName = value.(string)
			} else {
				properties[matchItems.FieldNames[i]] = value
			}
		}

		if itemId != "" {
			item := module.NewItem(itemId)
			item.Score = score
			item.AddProperties(properties)
			if config, exist := r.recallNameMapping[recallName]; exist {
				var values []any
				for _, field := range config.Fields {
					if field == beRecallNameV2 {
						values = append(values, recallName)
					} else {
						values = append(values, properties[field])
					}
				}

				item.RetrieveId = fmt.Sprintf(config.Format, values...)

			} else {
				item.RetrieveId = recallName
			}

			ret = append(ret, item)
		}
	}
	return
}

/**
func (r *BeMultiBizRecall) BuildRecallParam(user *module.User, context *context.RecommendContext) *be.RecallParam {
	return nil
}
**/

func (r *BeMultiBizRecall) BuildQueryParams(user *module.User, context *context.RecommendContext) (ret map[string]string) {
	return
}

func (r *BeMultiBizRecall) CloneWithConfig(params map[string]interface{}) BeBaseRecall {
	return r
}
