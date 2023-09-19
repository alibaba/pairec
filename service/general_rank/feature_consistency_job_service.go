package general_rank

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

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

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type FeatureConsistencyJobService struct {
	rateLimit uint64
	mutex     sync.Mutex
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
				}
			}
		}
	}
}

func (s *FeatureConsistencyJobService) checkFeatureConsistencyJobForRunning(job *model.FeatureConsistencyJob, user *module.User, items []*module.Item, context *context.RecommendContext) bool {

	scene := context.Param.GetParameter("scene").(string)
	if job.Status != common.Feature_Consistency_Job_State_RUNNING {
		return false
	}
	currTime := time.Now().Unix()
	if currTime >= job.StartTime && currTime <= job.EndTime {
		rankAlgoNames := s.findRankAlgoNames(scene, context)

		var easModelAlgoNames []string
		for _, algoConfig := range context.Config.AlgoConfs {
			if algoConfig.EasConf.Url == job.EasModelUrl {
				easModelAlgoNames = append(easModelAlgoNames, algoConfig.Name)
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

			if found {
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
func (s *FeatureConsistencyJobService) findRankAlgoNames(scene string, context *context.RecommendContext) []string {
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

	return rankConfig.RankConf.RankAlgoList
}

func (r *FeatureConsistencyJobService) logRankResultToDatahub(user *module.User, items []*module.Item, context *context.RecommendContext, job *model.FeatureConsistencyJob) {
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

	userProperties := user.MakeUserFeatures2()
	userProperties["_module_"] = "general_rank"
	userFeatures := utils.ConvertFeatures(userProperties)
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
