package recallenginerecall

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/recallengine"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/recall/berecall"
	"github.com/alibaba/pairec/v2/utils"
	re "github.com/aliyun/aliyun-pairec-config-go-sdk/v2/recallengine"
)

// sliceToAny converts a typed slice to []any using generics.
func sliceToAny[T any](s []T) []any {
	result := make([]any, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}

// mergeFeaturesToContextParams merges features map into contextParams,
// converting slice types to []any.
func mergeFeaturesToContextParams(features map[string]interface{}, contextParams map[string]interface{}) {
	for k, v := range features {
		switch val := v.(type) {
		case []string:
			contextParams[k] = sliceToAny(val)
		case []int:
			contextParams[k] = sliceToAny(val)
		case []int32:
			contextParams[k] = sliceToAny(val)
		case []int64:
			contextParams[k] = sliceToAny(val)
		case []float32:
			contextParams[k] = sliceToAny(val)
		case []float64:
			contextParams[k] = sliceToAny(val)
		default:
			contextParams[k] = v
		}
	}
}

var _ RecallEngineBaseRecall = (*RecallEngineServiceRecall)(nil)

type RecallEngineServiceRecall struct {
	returnCount       int
	modelName         string // recall engine recall name
	bizName           string
	serviceName       string
	versionName       string
	itemIdName        string
	client            *recallengine.RecallEngineClient
	recallMap         map[string]RecallEngineBaseRecall
	beFilterNames     []string
	beABParams        map[string]interface{}
	recallNameMapping map[string]recconf.RecallNameMappingConfig
}

func NewRecallEngineServiceRecall(client *recallengine.RecallEngineClient, conf recconf.RecallEngineConfig, modelName string) *RecallEngineServiceRecall {
	if len(conf.RecallEngineParams) < 1 {
		return nil
	}

	r := RecallEngineServiceRecall{
		returnCount:       conf.Count,
		modelName:         modelName,
		serviceName:       conf.ServiceName,
		versionName:       conf.VersionName,
		client:            client,
		beFilterNames:     conf.BeFilterNames,
		beABParams:        conf.BeABParams,
		recallMap:         make(map[string]RecallEngineBaseRecall, 8),
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

	for _, param := range conf.RecallEngineParams {
		switch param.RecallType {
		case recconf.RecallEngine_RecallType_X2I:
			recall := &RecallEngineX2IRecall{
				recallName:      param.RecallName,
				returnCount:     param.Count,
				scorerClause:    param.ScorerClause,
				triggerIdName:   param.TriggerIdName,
				recallTableName: param.RecallTableName,
				diversityParam:  param.DiversityParam,
				customParams:    param.CustomParams,
				triggerKey:      NewTriggerKey(&param, client),
				client:          client,
				cloneInstances:  make(map[string]*RecallEngineX2IRecall),
			}

			r.recallMap[param.RecallName] = recall
		case recconf.RecallEngine_RecallType_Random:
			recall := &RecallEngineRandomRecall{
				recallName:      param.RecallName,
				returnCount:     param.Count,
				scorerClause:    param.ScorerClause,
				recallTableName: param.RecallTableName,
				diversityParam:  param.DiversityParam,
				customParams:    param.CustomParams,
				client:          client,
				cloneInstances:  make(map[string]*RecallEngineRandomRecall),
			}
			r.recallMap[param.RecallName] = recall
			/*
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
			*/
		}
	}

	return &r
}
func (r *RecallEngineServiceRecall) getRecalls(user *module.User, context *context.RecommendContext) (recalls []RecallEngineBaseRecall) {
	recallMap := make(map[string]RecallEngineBaseRecall, len(r.recallMap))
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
func (r *RecallEngineServiceRecall) buildRequest(user *module.User, context *context.RecommendContext) *re.RecallRequest {
	recallRequest := re.RecallRequest{
		Recalls:       make(map[string]re.RecallConf),
		ContextParams: make(map[string]interface{}),
		InstanceId:    r.client.InstanceId(),
	}
	recallRequest.RequestId = context.RecommendId
	if context.Debug {
		recallRequest.Debug = true
	}
	recallRequest.Uid = string(user.Id)
	recallRequest.Service = r.serviceName
	if r.versionName != "" {
		recallRequest.Version = r.versionName
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	/*
		beABParams := r.beABParams
		if context.ExperimentResult != nil {
			params := context.ExperimentResult.GetExperimentParams().Get(fmt.Sprintf("recall.%s.beABParams", r.modelName), nil)
			if params != nil {
				if abparams, ok := params.(map[string]interface{}); ok {
					beABParams = abparams
				}
			}
		}
	*/
	recalls := r.getRecalls(user, context)
	for _, recall := range recalls {
		wg.Add(1)
		go func(recall RecallEngineBaseRecall) {
			defer wg.Done()
			recallConf := recall.BuildQueryParams(user, context)
			if recallConf.Count > 0 {
				mu.Lock()
				defer mu.Unlock()
				recallRequest.Recalls[recall.GetRecallName()] = recallConf
			}

		}(recall)
	}

	if len(r.beFilterNames) > 0 {
		for _, name := range r.beFilterNames {
			if filter, err := berecall.GetFilter(name); err == nil {
				wg.Add(1)
				go func(filer berecall.IBeFilter) {
					defer wg.Done()
					//filterParams := filter.BuildQueryParams(user, context)
					mu.Lock()
					defer mu.Unlock()
					/*
						for k, v := range filterParams {
							if r.client.IsProductReleased() {
								params[k] = strings.ReplaceAll(v, ",", "|")
							} else {
								params[k] = v
							}
						}
					*/

				}(filter)
			}
		}
	}
	features := context.GetParameter("features").(map[string]interface{})
	if len(features) > 0 {
		mergeFeaturesToContextParams(features, recallRequest.ContextParams)
	}
	wg.Wait()

	/*
		for k, v := range beABParams {
			params[k] = utils.ToString(v, "")
		}
	*/

	if context.Debug {
		log.Info(fmt.Sprintf("requestId=%s\tname=%s\tserviceName=%s\tversionName=%srequest=%v", context.RecommendId,
			r.modelName, r.serviceName, r.versionName, recallRequest))
	}
	return &recallRequest
}

func (r *RecallEngineServiceRecall) GetItems(user *module.User, context *context.RecommendContext) (ret []*module.Item, err error) {
	start := time.Now()
	recallRequest := r.buildRequest(user, context)

	response, err := r.client.GetRecallEngineClient().Recall(recallRequest)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=RecallEngineRecall\tname=%s\tserviceName=%s\tversionName=%s\trequest=%v\terror=%v",
			context.RecommendId, r.modelName, r.serviceName, r.versionName, *recallRequest, err))
		return
	}

	if response != nil && response.Result != nil {
		record := response.Result
		ret = make([]*module.Item, record.Size())
		fieldNames := record.FieldNames()
		tableIndex := record.TableIndex()
		size := record.Size()

		itemIdColumn := record.GetColumn(REItemIdFieldName)
		scoreColumn := record.GetColumn(REScoreFieldName)
		recallNameColumn := record.GetColumn(RERecallName)
		if itemIdColumn == nil {
			return nil, fmt.Errorf("item_id column not found")
		}

		for i := 0; i < size; i++ {
			index := tableIndex.GetIndex(i)
			if v, err := itemIdColumn.Get(index); err == nil {
				if itemId := utils.ToString(v, ""); itemId != "" {
					ret[i] = module.NewItem(itemId)
					if scoreColumn != nil {
						if v, err := scoreColumn.Get(index); err == nil {
							ret[i].Score = utils.ToFloat(v, 0)
						}
					}
					if recallNameColumn != nil {
						if v, err := recallNameColumn.Get(index); err == nil {
							ret[i].RetrieveId = utils.ToString(v, "")
						}
					}
				}
			}
		}
		//  slices.DeleteFunc (Go 1.21+)
		fieldNames = slices.DeleteFunc(fieldNames, func(name string) bool {
			return name == REItemIdFieldName || name == REScoreFieldName || name == RERecallName
		})
		for i := 0; i < size; i++ {
			properties := make(map[string]interface{}, len(fieldNames))
			for _, name := range fieldNames {
				column := record.GetColumn(name)
				if column == nil {
					continue
				}
				if v, err := column.Get(tableIndex.GetIndex(i)); err == nil {
					properties[name] = v
				}
			}
			item := ret[i]
			if item != nil {
				item.AddProperties(properties)
			}

		}
		ret = slices.DeleteFunc(ret, func(item *module.Item) bool {
			return item == nil
		})
	}

	log.Info(fmt.Sprintf("requestId=%s\ttmodule=RecallEngineRecall\tname=%s\tcount=%d\tcost=%d",
		context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return
}

func (r *RecallEngineServiceRecall) BuildQueryParams(user *module.User, context *context.RecommendContext) (ret re.RecallConf) {
	return
}

func (r *RecallEngineServiceRecall) CloneWithConfig(params map[string]interface{}) RecallEngineBaseRecall {
	return r
}
func (r *RecallEngineServiceRecall) GetRecallName() string {
	return r.modelName
}
