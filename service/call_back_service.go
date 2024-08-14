package service

import (
	"encoding/json"
	"fmt"
	"github.com/alibaba/pairec/v2/datasource/kafka"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/algorithm/eas"
	"github.com/alibaba/pairec/v2/algorithm/response"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/datahub"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/debug"
	"github.com/alibaba/pairec/v2/service/feature"
	"github.com/alibaba/pairec/v2/service/rank"
	"github.com/alibaba/pairec/v2/utils"
)

type CallBackService struct {
	recallService      *RecallService
	featureService     *feature.FeatureService
	userFeatureService *feature.UserFeatureService
	User               *module.User
	Items              []*module.Item
}

func NewCallBackService() *CallBackService {
	services := CallBackService{
		recallService:      &RecallService{},
		featureService:     feature.DefaultFeatureService(),
		userFeatureService: feature.DefaultUserFeatureService(),
	}
	return &services
}

func (r *CallBackService) LoadUserFeatures(context *context.RecommendContext) {
	//user feature prefetch
	r.userFeatureService.LoadUserFeaturesForCallback(r.User, context)
}

func (r *CallBackService) Recommend(context *context.RecommendContext) {
	scene_name := context.GetParameter("scene").(string)
	if _, ok := context.Config.SceneConfs[scene_name]; ok {
		items := r.recallService.GetItems(r.User, context)
		itemMap := make(map[module.ItemId]*module.Item, len(items))
		for _, item := range items {
			itemMap[item.Id] = item
		}

		for _, item := range r.Items {
			if itemM, ok := itemMap[item.Id]; ok {
				item.RetrieveId = itemM.RetrieveId
			}
		}

		go func() {
			debugService := debug.NewDebugService(r.User, context)
			debugService.WriteRecallLog(r.User, items, context)
		}()
	}
}

func (r *CallBackService) LoadFeatures(context *context.RecommendContext) {
	// load user features
	r.featureService.LoadFeatures(r.User, r.Items, context)
}

func (r *CallBackService) RecordLog(context *context.RecommendContext, msg string) error {
	scene_name := context.GetParameter("scene").(string)
	scene_name = strings.ReplaceAll(scene_name, "_callback", "")
	if callBackConfig, ok := context.Config.CallBackConfs[scene_name]; ok {
		data_source := callBackConfig.DataSource
		if data_source.Type == recconf.DataSource_Type_Kafka {
			return r.RecordToKafka(data_source.Name, msg)
		} else if data_source.Type == recconf.DataSource_Type_Datahub {
			return r.RecordToDatahub(data_source.Name, msg)
		}
	}

	return nil
}

func (r *CallBackService) RecordToKafka(kafka_name string, msg string) error {
	p, error := kafka.GetKafkaProducer(kafka_name)
	if error != nil {
		return error
	}
	p.SendMessage([]byte(msg))
	return nil
}

func (r *CallBackService) RecordToDatahub(name string, msg string) error {
	p, error := datahub.GetDatahub(name)
	if error != nil {
		return error
	}
	p.SendMessage([]map[string]interface{}{{"callback_log": msg}})
	return nil
}

func (r *CallBackService) RecordLogList(context *context.RecommendContext, messages []map[string]interface{}) error {
	scene_name := context.GetParameter("scene").(string)
	scene_name = strings.ReplaceAll(scene_name, "_callback", "")
	if callBackConfig, ok := context.Config.CallBackConfs[scene_name]; ok {
		data_source := callBackConfig.DataSource
		if data_source.Type == recconf.DataSource_Type_Kafka {
			return r.RecordToKafkaList(data_source.Name, messages)
		} else if data_source.Type == recconf.DataSource_Type_Datahub {
			return r.RecordToDatahubList(data_source.Name, messages)
		}
	}

	return nil
}

func (r *CallBackService) RecordToKafkaList(kafka_name string, messages []map[string]interface{}) error {
	p, error := kafka.GetKafkaProducer(kafka_name)
	if error != nil {
		return error
	}
	for _, msg := range messages {
		j, _ := json.Marshal(msg)
		p.SendMessage([]byte(string(j)))
	}
	return nil
}

func (r *CallBackService) RecordToDatahubList(name string, messages []map[string]interface{}) error {
	p, error := datahub.GetDatahub(name)
	if error != nil {
		return error
	}

	p.SendMessage(messages)
	return nil
}
func (r *CallBackService) Rank(context *context.RecommendContext) {
	var rankConfig recconf.RankConfig
	scene_name := context.GetParameter("scene").(string)
	scene_name = strings.ReplaceAll(scene_name, "_callback", "")
	callBackConfig, ok := context.Config.CallBackConfs[scene_name]

	if ok {
		rankConfig = callBackConfig.RankConf
	}

	if len(rankConfig.RankAlgoList) == 0 {
		return
	}

	start := time.Now()

	rankItems := r.Items

	algoGenerator := rank.CreateAlgoDataGenerator(rankConfig.Processor, rankConfig.ContextFeatures)

	var userFeatures map[string]interface{}

	if rankConfig.Processor == eas.Eas_Processor_EASYREC {
		userFeatures = r.User.MakeUserFeatures2()
	} else {
		userFeatures = r.User.MakeUserFeatures()
	}

	for _, item := range rankItems {
		features := item.GetFeatures()
		algoGenerator.AddFeatures(item, features, userFeatures)
	}

	var algoData rank.IAlgoData
	debugLevel := 3

	writeRawFeatrues := false
	if callBackConfig.RawFeatures && callBackConfig.RawFeaturesRate > 0 {
		if rand.Intn(100) < callBackConfig.RawFeaturesRate {
			debugLevel = 1
			writeRawFeatrues = true
		}
	}

	if algoGenerator.HasFeatures() {
		algoData = algoGenerator.GeneratorAlgoDataDebugWithLevel(debugLevel)
	}

	var wg sync.WaitGroup
	for _, algoName := range rankConfig.RankAlgoList {
		wg.Add(1)
		go func(algo string) {
			defer wg.Done()

			newAlgoName := algo + "_callback"
			found := false
			var processor string
			for _, config := range context.Config.AlgoConfs {
				if config.Name == newAlgoName {
					found = true
					processor = config.EasConf.Processor
					break
				}
			}
			if !found {
				var algoConfig recconf.AlgoConfig
				for _, config := range context.Config.AlgoConfs {
					if config.Name == algo {
						algoConfig = config
						processor = config.EasConf.Processor
						// change algoname  and  response function name
						algoConfig.Name = newAlgoName
						if algoConfig.EasConf.ResponseFuncName != "" {
							algoConfig.EasConf.ResponseFuncName += "Debug"
						}
						algorithm.AddAlgo(algoConfig)
						break
					}
				}
			}
			// run 返回原始的值，然后处理返回数据// 注册配置
			ret, err := algorithm.Run(newAlgoName, algoData.GetFeatures())
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\terror=run algorithm error(%v)", context.RecommendId, err))
				algoData.SetError(err)
			} else {
				if result, ok := ret.([]response.AlgoResponse); ok {
					algoData.SetAlgoResult(algo, result)
					if processor == eas.Eas_Processor_EASYREC {
						itemList := algoData.GetItems()
						for j := 0; j < len(result) && j < len(itemList); j++ {
							response, _ := (result[j]).(*eas.EasyrecResponse)
							if writeRawFeatrues {
								itemList[j].AddProperty("raw_features", response.RawFeatures)
							}

							//itemList[j].AddProperty("generate_features", response.GenerateFeatures)
							itemList[j].Properties["generate_features"] = response.GenerateFeatures
							itemList[j].AddProperty("context_features", response.ContextFeatures)
						}
					}
				}
			}

		}(algoName)

	}

	wg.Wait()

	log.Info(fmt.Sprintf("requestId=%s\tmodule=CallBackRank\tcost=%d", context.RecommendId, utils.CostTime(start)))
}
