package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/alibaba/pairec/algorithm"
	"github.com/alibaba/pairec/algorithm/eas"
	"github.com/alibaba/pairec/algorithm/response"
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/datasource/datahub"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/service/rank"
	"github.com/alibaba/pairec/utils"
)

type FeatureReplyService struct {
	RecommendService
}

func NewFeatureReplyService() *FeatureReplyService {
	service := FeatureReplyService{}
	return &service
}

func (r *FeatureReplyService) FeatureReply(userFeatures string, itemFeatures, itemids []string, context *context.RecommendContext) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	userId := r.GetUID(context)
	user := module.NewUserWithContext(userId, context)
	userProperties := make(map[string]interface{}, 0)
	if userFeatures != "" {
		features := make(map[string]*utils.FeatureInfo, 0)
		if err := json.Unmarshal([]byte(userFeatures), &features); err != nil {
			log.Info(fmt.Sprintf("requestId=%s\tmodule=FeatureReply\terror=%v", context.RecommendId, err))
			return
		}

		for name, feature := range features {
			userProperties[name] = utils.GetValueByType(feature.Value, feature.Type)
		}
	}

	user.SetProperties(userProperties)
	items := make([]*module.Item, 0, len(itemids))

	for i, itemFeature := range itemFeatures {
		itemProperties := make(map[string]interface{}, 0)
		if itemFeature != "" {
			features := make(map[string]*utils.FeatureInfo, 0)
			if err := json.Unmarshal([]byte(itemFeature), &features); err != nil {
				log.Info(fmt.Sprintf("requestId=%s\tmodule=FeatureReply\terror=%v", context.RecommendId, err))
				return
			}

			for name, feature := range features {
				itemProperties[name] = utils.GetValueByType(feature.Value, feature.Type)
			}
		}

		item := module.NewItemWithProperty(itemids[i], itemProperties)
		items = append(items, item)

	}

	if module, ok := userProperties["_module_"]; ok && module.(string) == "general_rank" {
		r.generalRank(user, items, context)
	} else {
		r.rank(user, items, context)
	}
}
func (r *FeatureReplyService) generalRank(user *module.User, items []*module.Item, context *context.RecommendContext) {
	start := time.Now()

	rankItems := items
	scene := context.GetParameter("scene").(string)

	// find rank config
	var rankConfig recconf.GeneralRankConfig
	found := false
	if context.ExperimentResult != nil {
		rankconf := context.ExperimentResult.GetExperimentParams().Get("generalRankConf", "")
		if rankconf != "" {
			d, _ := json.Marshal(rankconf)
			if err := json.Unmarshal(d, &rankConfig); err == nil {
				found = true
			}
		}
	}
	if !found {
		if rankConfigs, ok := recconf.Config.GeneralRankConfs[scene]; ok {
			rankConfig = rankConfigs
		}
	}

	algoGenerator := rank.CreateAlgoDataGenerator(rankConfig.RankConf.Processor, rankConfig.RankConf.ContextFeatures)

	var userFeatures map[string]interface{}

	if rankConfig.RankConf.Processor == eas.Eas_Processor_EASYREC {
		userFeatures = user.MakeUserFeatures2()
	} else {
		userFeatures = user.MakeUserFeatures()
	}

	for _, item := range rankItems {
		features := item.GetFeatures()
		algoGenerator.AddFeatures(item, features, userFeatures)
	}

	var algoData rank.IAlgoData
	if algoGenerator.HasFeatures() {
		algoData = algoGenerator.GeneratorAlgoDataDebugWithLevel(1)
	}

	var wg sync.WaitGroup
	for _, algoName := range rankConfig.RankConf.RankAlgoList {
		wg.Add(1)
		go func(algo string) {
			defer wg.Done()

			newAlgoName := algo + "_feature_reply"
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
							itemList[j].AddProperty("raw_features", response.RawFeatures)
							itemList[j].AddProperty("generate_features", response.GenerateFeatures.String())
							itemList[j].AddProperty("context_features", response.ContextFeatures)

						}
					}
				}
			}

		}(algoName)

	}

	wg.Wait()
	if algoData.Error() == nil && algoData.GetAlgoResult() != nil {
		go r.logFeatureReplyResult(user, items, context)
	}

	log.Info(fmt.Sprintf("requestId=%s\tmodule=general_rank\tcost=%d", context.RecommendId, utils.CostTime(start)))
}

func (r *FeatureReplyService) rank(user *module.User, items []*module.Item, context *context.RecommendContext) {
	start := time.Now()

	rankItems := items
	scene := context.GetParameter("scene").(string)

	// find rank config
	var rankConfig recconf.RankConfig
	found := false
	if context.ExperimentResult != nil {
		rankconf := context.ExperimentResult.GetExperimentParams().Get("rankconf", "")
		if rankconf != "" {
			d, _ := json.Marshal(rankconf)
			if err := json.Unmarshal(d, &rankConfig); err == nil {
				found = true
			}
		}
	}
	if !found {
		if rankConfigs, ok := recconf.Config.RankConf[scene]; ok {
			rankConfig = rankConfigs
		}
	}

	algoGenerator := rank.CreateAlgoDataGenerator(rankConfig.Processor, rankConfig.ContextFeatures)

	var userFeatures map[string]interface{}

	if rankConfig.Processor == eas.Eas_Processor_EASYREC {
		userFeatures = user.MakeUserFeatures2()
	} else {
		userFeatures = user.MakeUserFeatures()
	}

	for _, item := range rankItems {
		features := item.GetFeatures()
		algoGenerator.AddFeatures(item, features, userFeatures)
	}

	var algoData rank.IAlgoData
	if algoGenerator.HasFeatures() {
		algoData = algoGenerator.GeneratorAlgoDataDebugWithLevel(1)
	}

	var wg sync.WaitGroup
	for _, algoName := range rankConfig.RankAlgoList {
		wg.Add(1)
		go func(algo string) {
			defer wg.Done()

			userAlgo := user.StringProperty("_algo_")
			// algo name not equal
			if userAlgo != "" && userAlgo != algo {
				return
			}

			newAlgoName := algo + "_feature_reply"
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
							itemList[j].AddProperty("raw_features", response.RawFeatures)
							itemList[j].AddProperty("generate_features", response.GenerateFeatures.String())
							itemList[j].AddProperty("context_features", response.ContextFeatures)

						}
					}
				}
			}

		}(algoName)

	}

	wg.Wait()
	if algoData.Error() == nil && algoData.GetAlgoResult() != nil {
		/**
		for name, algoResult := range algoData.GetAlgoResult() {
			itemList := algoData.GetItems()
			for j := 0; j < len(algoResult) && j < len(itemList); j++ {
				if algoResult[j].GetModuleType() {
					arr_score := algoResult[j].GetScoreMap()
					for k, v := range arr_score {
						itemList[j].AddAlgoScore(name+"_"+k, v)
					}
				} else {
					itemList[j].AddAlgoScore(name, algoResult[j].GetScore())
				}
			}
		}
		**/
		go r.logFeatureReplyResult(user, items, context)
	}

	log.Info(fmt.Sprintf("requestId=%s\tmodule=rank\tcost=%d", context.RecommendId, utils.CostTime(start)))
}

func (r *FeatureReplyService) logFeatureReplyResult(user *module.User, items []*module.Item, context *context.RecommendContext) {
	datasourceType := context.GetParameter("datasource_type").(string)
	if datasourceType == "datahub" {
		r.logReatureReplyResultToDatahub(user, items, context)
	} else if datasourceType == "eas" {
		r.logReatureReplyResultToPairecConfigServer(user, items, context)
	}
}
func (r *FeatureReplyService) logReatureReplyResultToDatahub(user *module.User, items []*module.Item, context *context.RecommendContext) {
	accessId := context.GetParameter("access_id").(string)
	accessKey := context.GetParameter("access_key").(string)
	endpoint := context.GetParameter("endpoint").(string)
	project := context.GetParameter("project").(string)
	topic := context.GetParameter("topic").(string)

	name := fmt.Sprintf("%s-%s-%s-%s-%s", accessId, accessKey, endpoint, project, topic)

	var dh *datahub.Datahub
	dh, err := datahub.GetDatahub(name)
	if err != nil {
		dh = datahub.NewDatahub(accessId, accessKey, endpoint, project, topic, nil)
		err = dh.Init()
		datahub.RegisterDatahub(name, dh)
	}

	if dh == nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=logFeatureReplyResultToDatahub\tmsg=create datahub error\terror=%v", context.RecommendId, err))
		return
	}

	scene := context.Param.GetParameter("scene").(string)
	var data []map[string]interface{}
	for _, item := range items {
		message := make(map[string]interface{})
		message["request_id"] = context.RecommendId
		message["request_time"] = time.Now().Unix()
		message["scene"] = scene
		message["user_id"] = string(user.Id)
		message["item_id"] = string(item.Id)
		message["raw_features"] = item.StringProperty("raw_features")
		message["generate_features"] = item.StringProperty("generate_features")
		message["context_features"] = item.StringProperty("context_features")
		// j, _ := json.Marshal(item.GetAlgoScores())
		// message["scores"] = string(j)
		data = append(data, message)
	}
	dh.SendMessage(data)
}

var pairecConfigClient *http.Client

func init() {
	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   1000 * time.Millisecond, // 1000ms
			KeepAlive: 5 * time.Minute,
		}).DialContext,
		MaxIdleConnsPerHost:   200,
		MaxIdleConns:          200,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
	}

	pairecConfigClient = &http.Client{Transport: tr}
}

func (r *FeatureReplyService) logReatureReplyResultToPairecConfigServer(user *module.User, items []*module.Item, context *context.RecommendContext) {
	endpoint := context.GetParameter("vpc_address").(string)
	token := context.GetParameter("token").(string)
	jobId := context.GetParameter("job_id")

	var url string
	if endpoint[len(endpoint)-1] == '/' {
		url = endpoint + "v1/feature_consistency_reply"
	} else {
		url = endpoint + "/v1/feature_consistency_reply"
	}

	scene := context.Param.GetParameter("scene").(string)
	for _, item := range items {
		message := make(map[string]interface{})
		message["request_id"] = context.RecommendId
		message["request_time"] = time.Now().UnixMilli()
		message["scene"] = scene
		message["user_id"] = string(user.Id)
		message["item_id"] = string(item.Id)
		message["raw_features"] = item.StringProperty("raw_features")
		message["generate_features"] = item.StringProperty("generate_features")
		message["context_features"] = item.StringProperty("context_features")
		message["job_id"] = jobId
		// j, _ := json.Marshal(item.GetAlgoScores())
		// message["scores"] = string(j)
		j, _ := json.Marshal(message)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tevent=logReatureReplyResultToPairecConfigServer\terror=%v", context.RecommendId, err))
			continue
		}

		headers := map[string][]string{
			"Authorization": {token},
			"Content-Type":  {"application/json"},
		}
		req.Header = headers

		resp, err := pairecConfigClient.Do(req)
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tevent=logReatureReplyResultToPairecConfigServer\terror=%v", context.RecommendId, err))
			continue
		}

		resp.Body.Close()
	}
}
