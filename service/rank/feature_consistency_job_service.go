package rank

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	be "github.com/aliyun/aliyun-be-go-sdk"
	"github.com/alibaba/pairec/datasource/beengine"

	"github.com/alibaba/pairec/algorithm"
	"github.com/alibaba/pairec/algorithm/eas"
	"github.com/alibaba/pairec/algorithm/eas/easyrec"
	"github.com/alibaba/pairec/algorithm/response"

	jsoniter "github.com/json-iterator/go"

	"github.com/aliyun/aliyun-pairec-config-go-sdk/common"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/model"
	"github.com/alibaba/pairec/abtest"
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/datasource/datahub"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/utils"
)

const (
	DssmO2o = "dssm_o2o"
	MindO2o = "mind_o2o"
)

type FeatureConsistencyJobService struct {
	mutex  sync.Mutex
	client *beengine.BeClient
}

func (r *FeatureConsistencyJobService) LogRecallResult(user *module.User, items []*module.Item, context *context.RecommendContext, triggerType, userEmbedding, triggerItem, recallAlgo, recallAlgoType, beName, bizName, recallName string) {
	if abtest.GetExperimentClient() != nil {
		scene := context.Param.GetParameter("scene").(string)
		jobs := abtest.GetExperimentClient().GetSceneParams(scene).GetFeatureConsistencyJobs()

		for _, job := range jobs {
			if r.checkRecallFeatureConsistencyJobForRunning(job, user, items, context, triggerType, recallAlgo) {
				log.Info(fmt.Sprintf("requestId=%s\tevent=logRecallResult\tname=%s", context.RecommendId, job.JobName))
				if job.FeatureBackflowQueueType == "datahub" {
					r.logRecallResultToDatahub(user, items, context, triggerType, job, userEmbedding, triggerItem, recallAlgo, recallAlgoType, beName, bizName, recallName)
				}
			}
		}
	}
}
func (r *FeatureConsistencyJobService) checkRecallFeatureConsistencyJobForRunning(job *model.FeatureConsistencyJob, user *module.User, items []*module.Item, context *context.RecommendContext, triggerType, recallAlgo string) bool {

	if job.Status != common.Feature_Consistency_Job_State_RUNNING {
		return false
	}

	if triggerType == DssmO2o || triggerType == MindO2o {
		return true
	}

	currTime := time.Now().Unix()
	if currTime >= job.StartTime && currTime <= job.EndTime {
		var easModelAlgoNames []string
		for _, algoConfig := range context.Config.AlgoConfs {
			if algoConfig.EasConf.Url == job.EasModelUrl {
				easModelAlgoNames = append(easModelAlgoNames, algoConfig.Name)
				user.AddProperty("_algo_", algoConfig.Name)
				break
			}
		}

		found := false

		for _, name := range easModelAlgoNames {
			if name == recallAlgo {
				found = true
				break
			}
		}

		if !found {
			return false
		}

		// sample rate check
		if job.SampleRate >= 100 {
			return true
		}

		if rand.Intn(100) < job.SampleRate {
			return true
		}

	}

	return false
}

func (r *FeatureConsistencyJobService) LogRankResult(user *module.User, items []*module.Item, context *context.RecommendContext) {

	if abtest.GetExperimentClient() != nil {
		scene := context.Param.GetParameter("scene").(string)
		jobs := abtest.GetExperimentClient().GetSceneParams(scene).GetFeatureConsistencyJobs()

		for _, job := range jobs {
			if r.checkFeatureConsistencyJobForRunning(job, user, items, context) {
				log.Info(fmt.Sprintf("requestId=%s\tevent=logRankResult\tname=%s", context.RecommendId, job.JobName))
				if job.FeatureBackflowQueueType == "datahub" {
					r.logRankResultToDatahub(user, items, context, job)
				} else if job.FeatureBackflowQueueType == "eas" {
					r.logRankResultToPaiConfigServer(user, items, context, job)
				}
			}
		}
	}
}

func (r *FeatureConsistencyJobService) checkFeatureConsistencyJobForRunning(job *model.FeatureConsistencyJob, user *module.User, items []*module.Item, context *context.RecommendContext) bool {

	scene := context.Param.GetParameter("scene").(string)
	if job.Status != common.Feature_Consistency_Job_State_RUNNING {
		return false
	}
	currTime := time.Now().Unix()
	if currTime >= job.StartTime && currTime <= job.EndTime {
		rankAlgoNames := r.findRankAlgoNames(scene, context)

		var easModelAlgoNames []string
		for _, algoConfig := range context.Config.AlgoConfs {
			if algoConfig.EasConf.Url == job.EasModelUrl {
				easModelAlgoNames = append(easModelAlgoNames, algoConfig.Name)
				user.AddProperty("_algo_", algoConfig.Name)
				break
			}
		}

		found := false

		for _, name := range easModelAlgoNames {
			for _, rankAlgoName := range rankAlgoNames {
				if name == rankAlgoName {
					found = true
					break
				}
			}
		}

		if !found {
			return false
		}

		// sample rate check
		if job.SampleRate >= 100 {
			return true
		}

		if rand.Intn(100) < job.SampleRate {
			return true
		}

	}

	return false
}

func (r *FeatureConsistencyJobService) findRankAlgoNames(scene string, context *context.RecommendContext) []string {
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

	return rankConfig.RankAlgoList
}

func (r *FeatureConsistencyJobService) logRankResultToDatahub(user *module.User, items []*module.Item, context *context.RecommendContext, job *model.FeatureConsistencyJob) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	name := fmt.Sprintf("%s-%s-%s-%s-%s", job.FeatureBackflowQueueDatahubAccessId, job.FeatureBackflowQueueDatahubAccessKey, job.FeatureBackflowQueueDatahubEndpoint,
		job.FeatureBackflowQueueDatahubProject, job.FeatureBackflowQueueDatahubTopic)
	var dh *datahub.Datahub
	dh, err := datahub.GetDatahub(name)
	if err != nil {
		r.mutex.Lock()
		dh, err = datahub.GetDatahub(name)
		if err != nil {
			dh = datahub.NewDatahub(job.FeatureBackflowQueueDatahubAccessId, job.FeatureBackflowQueueDatahubAccessKey, job.FeatureBackflowQueueDatahubEndpoint,
				job.FeatureBackflowQueueDatahubProject, job.FeatureBackflowQueueDatahubTopic, nil)
			err = dh.Init()
			datahub.RegisterDatahub(name, dh)
		}
		r.mutex.Unlock()
	}

	if dh == nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=logRankResultToDatahub\tmsg=create datahub error\terror=%v", context.RecommendId, err))
		return
	}

	userFeatures := utils.ConvertFeatures(user.MakeUserFeatures2())
	userData, _ := json.Marshal(userFeatures)
	userDataStr := utils.Byte2string(userData)
	scene := context.Param.GetParameter("scene").(string)
	var data []map[string]interface{}
	i := 0
	message := make(map[string]interface{})
	message["request_id"] = context.RecommendId
	message["scene"] = scene
	message["request_time"] = time.Now().Unix()
	message["user_id"] = string(user.Id)
	message["user_features"] = userDataStr
	var itemIds []string
	var itemFeatures []string
	var itemScores []string
	for _, item := range items {
		if i%10 == 0 {
			j, _ := json.Marshal(itemIds)
			message["item_id"] = string(j)
			j, _ = json.Marshal(itemFeatures)
			message["item_features"] = string(j)
			j, _ = json.Marshal(itemScores)
			message["scores"] = string(j)
			data = append(data, message)
			dh.SendMessage(data)
			data = data[:0]

			itemIds = itemIds[:0]
			itemFeatures = itemFeatures[:0]
			itemScores = itemScores[:0]
			i = 0
		}

		i++
		itemId := string(item.Id)
		itemIds = append(itemIds, itemId)
		j, _ := json.Marshal(utils.ConvertFeatures(item.GetCloneFeatures()))
		itemFeatures = append(itemFeatures, string(j))
		j, _ = json.Marshal(item.CloneAlgoScores())
		itemScores = append(itemScores, string(j))

	}
	if i > 0 {
		j, _ := json.Marshal(itemIds)
		message["item_id"] = string(j)
		j, _ = json.Marshal(itemFeatures)
		message["item_features"] = string(j)
		j, _ = json.Marshal(itemScores)
		message["scores"] = string(j)
		data = append(data, message)
		dh.SendMessage(data)
	}
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
func (r *FeatureConsistencyJobService) logRankResultToPaiConfigServer(user *module.User, items []*module.Item, context *context.RecommendContext, job *model.FeatureConsistencyJob) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	var url string
	if job.FeatureBackflowEASVpcAddress[len(job.FeatureBackflowEASVpcAddress)-1] == '/' {
		url = job.FeatureBackflowEASVpcAddress + "v1/feature_consistency_backflow"
	} else {
		url = job.FeatureBackflowEASVpcAddress + "/v1/feature_consistency_backflow"
	}

	userFeatures := utils.ConvertFeatures(user.MakeUserFeatures2())
	userData, _ := json.Marshal(userFeatures)
	userDataStr := utils.Byte2string(userData)
	scene := context.Param.GetParameter("scene").(string)
	message := make(map[string]interface{})
	message["job_id"] = job.JobId
	message["request_id"] = context.RecommendId
	message["scene"] = scene
	message["request_time"] = time.Now().UnixMilli()
	message["user_id"] = string(user.Id)
	message["user_features"] = userDataStr
	fmt.Println("logRankResultToPaiConfigServer", len(items))

	i := 0
	var itemIds []string
	var itemFeatures []string
	var itemScores []string
	data := make(map[string]any)
	for k, v := range message {
		data[k] = v
	}
	for _, item := range items {

		i++
		itemId := string(item.Id)
		itemIds = append(itemIds, itemId)
		j, _ := json.Marshal(utils.ConvertFeatures(item.GetCloneFeatures()))
		itemFeatures = append(itemFeatures, string(j))
		j, _ = json.Marshal(item.GetAlgoScores())
		itemScores = append(itemScores, string(j))

		if i%20 == 0 {
			j, _ := json.Marshal(itemIds)
			data["item_id"] = string(j)
			j, _ = json.Marshal(itemFeatures)
			data["item_features"] = string(j)
			j, _ = json.Marshal(itemScores)
			data["scores"] = string(j)
			data["request_time"] = time.Now().UnixMilli()
			j, _ = json.Marshal(data)
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tevent=logRankResultToPaiConfigServer\terror=%v", context.RecommendId, err))
				continue
			}

			headers := map[string][]string{
				"Authorization": {job.FeatureBackflowEASToken},
				"Content-Type":  {"application/json"},
			}
			req.Header = headers

			resp, err := pairecConfigClient.Do(req)
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tevent=logRankResultToPaiConfigServer\terror=%v", context.RecommendId, err))
				continue
			}

			resp.Body.Close()

			itemIds = itemIds[:0]
			itemFeatures = itemFeatures[:0]
			itemScores = itemScores[:0]
			i = 0
		}

	}
	if i > 0 {
		j, _ := json.Marshal(itemIds)
		data["item_id"] = string(j)
		j, _ = json.Marshal(itemFeatures)
		data["item_features"] = string(j)
		j, _ = json.Marshal(itemScores)
		data["scores"] = string(j)
		data["request_time"] = time.Now().UnixMilli()
		j, _ = json.Marshal(data)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tevent=logRankResultToPaiConfigServer\terror=%v", context.RecommendId, err))
			return
		}

		headers := map[string][]string{
			"Authorization": {job.FeatureBackflowEASToken},
			"Content-Type":  {"application/json"},
		}
		req.Header = headers

		resp, err := pairecConfigClient.Do(req)
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tevent=logRankResultToPaiConfigServer\terror=%v", context.RecommendId, err))
			return
		}

		resp.Body.Close()

	}
}

func (r *FeatureConsistencyJobService) logRecallResultToDatahub(user *module.User, items []*module.Item, context *context.RecommendContext, triggerType string, job *model.FeatureConsistencyJob, userEmbedding, triggerItem, recallAlgo, recallAlgoType, beName, bizName, recallName string) {
	name := fmt.Sprintf("%s-%s-%s-%s-%s", job.FeatureBackflowQueueDatahubAccessId, job.FeatureBackflowQueueDatahubAccessKey, job.FeatureBackflowQueueDatahubEndpoint,
		job.FeatureBackflowQueueDatahubProject, job.FeatureBackflowQueueDatahubTopic)
	var dh *datahub.Datahub
	dh, err := datahub.GetDatahub(name)
	if err != nil {
		r.mutex.Lock()
		dh, err = datahub.GetDatahub(name)
		if err != nil {
			dh = datahub.NewDatahub(job.FeatureBackflowQueueDatahubAccessId, job.FeatureBackflowQueueDatahubAccessKey, job.FeatureBackflowQueueDatahubEndpoint,
				job.FeatureBackflowQueueDatahubProject, job.FeatureBackflowQueueDatahubTopic, nil)
			err = dh.Init()
			datahub.RegisterDatahub(name, dh)
		}
		r.mutex.Unlock()
	}

	if dh == nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=logRecallResultToDatahub\tmsg=create datahub error\terror=%v", context.RecommendId, err))
		return
	}

	message := make(map[string]interface{})

	message["request_id"] = context.RecommendId
	message["uid"] = string(user.Id)

	features := user.MakeUserFeatures2()
	j, _ := json.Marshal(utils.ConvertFeatures(features))
	message["user_features"] = string(j)
	message["trigger"] = triggerItem
	message["module"] = triggerType

	if triggerType == DssmO2o || triggerType == MindO2o {
		readRequest := be.NewReadRequest(bizName, 10)
		readRequest.IsRawRequest = true
		readRequest.IsPost = true

		params := make(map[string]string)
		params[fmt.Sprintf("%s_qinfo", recallName)] = triggerItem
		params[fmt.Sprintf("%s_return_count", recallName)] = "10"
		params["trace"] = "ALL"
		readRequest.SetQueryParams(params)

		if context.Debug {
			uri := readRequest.BuildUri()
			log.Info(fmt.Sprintf("requestId=%s\tevent=FeatureConsistencyJobService\tbizName=%s\turl=%s", context.RecommendId, bizName, uri.RequestURI()))
		}
		if r.client == nil {
			client, err := beengine.GetBeClient(beName)
			if err != nil {
				log.Error(fmt.Sprintf("error=%v", err))
				return
			}
			r.client = client
		}
		readResponse, err := r.client.BeClient.Read(*readRequest)
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tevent=FeatureConsistencyJobService\terror=be error(%v)", context.RecommendId, err))
		} else {
			traceInfo := *readResponse.Result.TraceInfo
			if traceInfo != nil {
				traceInfoMap := traceInfo.(map[string]interface{})
				valueMap := traceInfoMap["0[]"].(map[string]interface{})
				key := fmt.Sprintf("ExtractTensorFromRawTensorOpV2[recall/%s/extract_%s_user_embedding]", recallName, recallName)
				userEmbeddingArr := valueMap[key].(map[string]interface{})
				arr := userEmbeddingArr["__arr"].([]interface{})
				v := arr[0].(string)
				splitArr := strings.Split(v, "[[")
				result := strings.TrimRight(splitArr[len(splitArr)-1], "]]")
				embeddingArr := strings.Split(result, " ")
				userEmbedding = strings.Join(embeddingArr, ",")
			}
			message["user_embedding"] = userEmbedding
			message["generate_features"] = ""
			dh.SendMessage([]map[string]interface{}{message})
		}
	} else {
		message["user_embedding"] = userEmbedding
		algoGenerator := CreateAlgoDataGenerator(recallAlgoType, nil)
		item := module.NewItem("1")
		algoGenerator.AddFeatures(item, nil, features)
		algoData := algoGenerator.GeneratorAlgoDataDebugWithLevel(1)
		easyrecRequest := algoData.GetFeatures().(*easyrec.PBRequest)
		algoRet, err := algorithm.Run(recallAlgo, easyrecRequest)

		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tevent=logRecallResultToDatahub\terr=%v", context.RecommendId, err))
		} else {
			// eas model invoke success
			if triggerType == "dssm" {
				// eas model invoke success
				if result, ok := algoRet.([]response.AlgoResponse); ok && len(result) > 0 {
					if embeddingResponse, ok := result[0].(*eas.EasyrecUserRealtimeEmbeddingResponse); ok {
						if embeddingResponse.GenerateFeatures != nil {
							message["generate_features"] = embeddingResponse.GenerateFeatures.String()
						}
					}
				}
			} else if triggerType == "mind" {
				if result, ok := algoRet.([]response.AlgoResponse); ok && len(result) > 0 {
					if embeddingMindResponse, ok := result[0].(*eas.EasyrecUserRealtimeEmbeddingMindResponse); ok {
						if embeddingMindResponse.GenerateFeatures != nil {
							message["generate_features"] = embeddingMindResponse.GenerateFeatures.String()
						}
					}
				}

			}
		}
		dh.SendMessage([]map[string]interface{}{message})
	}
}
