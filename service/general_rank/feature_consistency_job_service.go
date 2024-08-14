package general_rank

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/alibaba/pairec/v2/abtest"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/common"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type FeatureConsistencyJobService struct {
}

func (r *FeatureConsistencyJobService) LogRankResult(user *module.User, items []*module.Item, context *context.RecommendContext) {

	if abtest.GetExperimentClient() != nil {
		scene := context.Param.GetParameter("scene").(string)
		jobs := abtest.GetExperimentClient().GetSceneParams(scene).GetFeatureConsistencyJobs()

		for _, job := range jobs {
			if r.checkFeatureConsistencyJobForRunning(job, user, items, context) {
				log.Info(fmt.Sprintf("requestId=%s\tevent=logRankResult\tname=%s", context.RecommendId, job.JobName))
				r.logRankResultToPaiConfigServer(user, items, context, job)
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
			urls := strings.Split(algoConfig.EasConf.Url, "/api/predict/")
			name := urls[1]
			if name == job.EasModelServiceName {
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

func (r *FeatureConsistencyJobService) logRankResultToPaiConfigServer(user *module.User, items []*module.Item, context *context.RecommendContext, job *model.FeatureConsistencyJob) {
	userProperties := user.MakeUserFeatures2()
	userProperties["_module_"] = "general_rank"
	userFeatures := utils.ConvertFeatures(userProperties)
	userData, _ := json.Marshal(userFeatures)
	userDataStr := utils.Byte2string(userData)
	scene := context.Param.GetParameter("scene").(string)

	i := 0

	backflowData := model.FeatureConsistencyBackflowData{}
	backflowData.FeatureConsistencyCheckJobConfigId = strconv.Itoa(job.JobId)
	backflowData.LogRequestId = context.RecommendId
	backflowData.SceneName = scene
	backflowData.LogRequestTime = time.Now().Unix()
	backflowData.LogUserId = string(user.Id)
	backflowData.UserFeatures = userDataStr

	var itemIds []string
	var itemFeatures []string
	var itemScores []string
	for _, item := range items {
		i++
		itemId := string(item.Id)
		itemIds = append(itemIds, itemId)
		j, _ := json.Marshal(utils.ConvertFeatures(item.GetCloneFeatures()))
		itemFeatures = append(itemFeatures, string(j))
		j, _ = json.Marshal(item.CloneAlgoScores())
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
		}
	}
}
