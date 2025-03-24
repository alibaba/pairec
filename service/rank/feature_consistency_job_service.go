package rank

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/datasource/beengine"
	be "github.com/aliyun/aliyun-be-go-sdk"

	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/algorithm/eas"
	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
	"github.com/alibaba/pairec/v2/algorithm/response"

	jsoniter "github.com/json-iterator/go"

	"github.com/alibaba/pairec/v2/abtest"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/datahub"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/common"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
)

const (
	DssmO2o = "dssm_o2o"
	MindO2o = "mind_o2o"
)

var serviceName string

func init() {
	serviceName = os.Getenv("SERVICE_NAME")

	// serviceName: name@region => name
	ss := strings.Split(serviceName, "@")
	if len(ss) > 1 {
		serviceName = strings.Join(ss[:len(ss)-1], "")
	}
}

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
			if job.ModelType == "rank_sample" {
				continue
			}

			if r.checkFeatureConsistencyJobForRunning(job, user, items, context) {
				log.Info(fmt.Sprintf("requestId=%s\tevent=logRankResult\tname=%s", context.RecommendId, job.JobName))
				r.logRankResultToPaiConfigServer(user, items, context, job)
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
			urls := strings.Split(algoConfig.EasConf.Url, "/api/predict/")
			name := urls[1]
			if name == job.EasModelServiceName {
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

func (r *FeatureConsistencyJobService) LogSampleResult(user *module.User, items []*module.Item, context *context.RecommendContext) {
	if abtest.GetExperimentClient() != nil {
		scene := context.Param.GetParameter("scene").(string)
		jobs := abtest.GetExperimentClient().GetSceneParams(scene).GetFeatureConsistencyJobs()

		for _, job := range jobs {
			if job.ModelType != "rank_sample" {
				continue
			}

			if r.checkFeatureConsistencyJobForRunning(job, user, items, context) {
				log.Info(fmt.Sprintf("requestId=%s\tevent=logRankResult\tname=%s", context.RecommendId, job.JobName))
				r.logRankResultToPaiConfigServer(user, items, context, job)
			}
		}
	}
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

func (r *FeatureConsistencyJobService) logRankResultToPaiConfigServer(user *module.User, items []*module.Item, context *context.RecommendContext, job *model.FeatureConsistencyJob) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	userFeatures := utils.ConvertFeatures(user.MakeUserFeatures2())
	userData, _ := json.Marshal(userFeatures)
	userDataStr := utils.Byte2string(userData)
	scene := context.Param.GetParameter("scene").(string)
	backflowData := model.FeatureConsistencyBackflowData{}
	backflowData.FeatureConsistencyCheckJobConfigId = strconv.Itoa(job.JobId)
	backflowData.LogRequestId = context.RecommendId
	backflowData.SceneName = scene
	backflowData.LogRequestTime = time.Now().UnixMilli()
	backflowData.LogUserId = string(user.Id)
	backflowData.UserFeatures = userDataStr

	if job.ModelType == "rank_sample" {
		backflowData.ServiceName = serviceName
	}

	i := 0
	var itemIds []string
	var itemFeatures []string
	var itemScores []string

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
			backflowData.LogItemId = string(j)
			j, _ = json.Marshal(itemFeatures)
			backflowData.ItemFeatures = string(j)
			j, _ = json.Marshal(itemScores)
			backflowData.Scores = string(j)

			resp, err := abtest.GetExperimentClient().BackflowFeatureConsistencyCheckJobData(&backflowData)
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tevent=logRankResultToPaiConfigServer\tresponse=%v\terror=%v", context.RecommendId, resp, err))
				continue
			}

			itemIds = itemIds[:0]
			itemFeatures = itemFeatures[:0]
			itemScores = itemScores[:0]
			i = 0
		}

	}
	if i > 0 {
		j, _ := json.Marshal(itemIds)
		backflowData.LogItemId = string(j)
		j, _ = json.Marshal(itemFeatures)
		backflowData.ItemFeatures = string(j)
		j, _ = json.Marshal(itemScores)
		backflowData.Scores = string(j)
		resp, err := abtest.GetExperimentClient().BackflowFeatureConsistencyCheckJobData(&backflowData)
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tevent=logRankResultToPaiConfigServer\tresponse=%v\terror=%v", context.RecommendId, resp, err))
			return
		}

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
		algoGenerator.SetItemFeatures(nil)
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
