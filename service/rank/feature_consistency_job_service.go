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
	"google.golang.org/protobuf/proto"

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

// FeatureConsistencySkipReplyDebugLevel is the special debug level carried by the
// online scoring request when a skip_feature_reply job is hit. The model service
// recognizes it and writes the generated features to OSS CSV directly, so the
// /api/feature_reply relay is skipped. The value is fixed on the engine side and
// must stay aligned with the model service and the overall consistency scheme.
const FeatureConsistencySkipReplyDebugLevel = 103

// featureConsistencySkipJobIdKey is the recommend context param key used to carry
// the set of sampled skip_feature_reply job ids (map[int]bool) from the synchronous
// scoring path to the asynchronous backflow path, so both share one sampling decision.
const featureConsistencySkipJobIdKey = "_fcj_skip_job_id_"

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

// MatchSkipFeatureReplyJob resolves the skip_feature_reply consistency jobs for this
// scoring call, keyed by the algo whose request should carry debug_level=103. The
// scoring fan-out already sends one request per algo, so decorating a matched algo's
// request adds no extra EAS call. Sampling is per job and independent: a job is served
// at exactly its own SampleRate, never conditioned on another job's sampling outcome.
// Jobs targeting different algos never compete; only jobs sharing the same algo contend
// for that algo's single request slot (and single task_id), resolved by one weighted
// draw. Sampling happens here exactly once; the matched job ids are stored in the
// recommend context so the async backflow path reuses the same decision.
func MatchSkipFeatureReplyJob(user *module.User, ctx *context.RecommendContext, rankAlgoNames []string) map[string]*model.FeatureConsistencyJob {
	client := abtest.GetExperimentClient()
	if client == nil || ctx == nil || user == nil {
		return nil
	}
	scene, _ := ctx.GetParameter("scene").(string)
	if scene == "" {
		return nil
	}
	jobs := client.GetSceneParams(scene).GetFeatureConsistencyJobs()
	currTime := time.Now().Unix()

	// Group the structurally-matched jobs by the algo they target, keeping first-seen
	// order so same-algo contention is resolved deterministically.
	jobsByAlgo := make(map[string][]*model.FeatureConsistencyJob)
	var algoOrder []string
	for _, job := range jobs {
		if job == nil || !job.SkipFeatureReply {
			continue
		}
		if job.Status != common.Feature_Consistency_Job_State_RUNNING {
			continue
		}
		if currTime < job.StartTime || currTime > job.EndTime {
			continue
		}
		algoName := matchEasModelAlgoName(job, ctx, rankAlgoNames)
		if algoName == "" {
			continue
		}
		if _, ok := jobsByAlgo[algoName]; !ok {
			algoOrder = append(algoOrder, algoName)
		}
		jobsByAlgo[algoName] = append(jobsByAlgo[algoName], job)
	}

	matched := make(map[string]*model.FeatureConsistencyJob)
	hitJobIds := make(map[int]bool)
	for _, algoName := range algoOrder {
		if job := sampleSkipJobGroup(ctx, scene, algoName, jobsByAlgo[algoName]); job != nil {
			matched[algoName] = job
			hitJobIds[job.JobId] = true
		}
	}
	if len(hitJobIds) == 0 {
		return nil
	}
	ctx.AddContextParam(featureConsistencySkipJobIdKey, hitJobIds)
	return matched
}

// sampleSkipJobGroup picks at most one job from a group of skip jobs that all target
// the same algo, so they contend for that algo's single request slot and single
// task_id. A lone job is sampled plainly at its own rate. Several jobs are resolved by
// one weighted draw over [0,100): each job is selected with probability SampleRate/100
// and the leftover selects none, keeping every rate exact and the selections mutually
// exclusive. When the rates oversubscribe (>100) exact rates for all are impossible, so
// the first job is sampled at its own rate, the rest are not served, and a warning marks
// the conflict.
func sampleSkipJobGroup(ctx *context.RecommendContext, scene, algoName string, group []*model.FeatureConsistencyJob) *model.FeatureConsistencyJob {
	if len(group) == 1 {
		if sampleHit(group[0].SampleRate) {
			return group[0]
		}
		return nil
	}

	total := 0
	for _, job := range group {
		total += sampleWeight(job.SampleRate)
	}

	if total > 100 {
		names := make([]string, 0, len(group))
		for _, job := range group {
			names = append(names, job.JobName)
		}
		log.Warning(fmt.Sprintf("requestId=%s\tevent=matchSkipFeatureReplyJobOversubscribed\tscene=%s\talgo=%s\tjobs=%s\tmsg=same-algo skip job sample rates exceed 100, serving only the first at its own rate", ctx.RecommendId, scene, algoName, strings.Join(names, ",")))
		if sampleHit(group[0].SampleRate) {
			return group[0]
		}
		return nil
	}

	r := rand.Intn(100)
	acc := 0
	for _, job := range group {
		acc += sampleWeight(job.SampleRate)
		if r < acc {
			return job
		}
	}
	return nil
}

// sampleWeight clamps a job sample rate into [0,100] for weighted selection.
func sampleWeight(rate int) int {
	if rate >= 100 {
		return 100
	}
	if rate <= 0 {
		return 0
	}
	return rate
}

// sampleHit reports whether a single job passes its own sample rate check.
func sampleHit(rate int) bool {
	if rate >= 100 {
		return true
	}
	if rate <= 0 {
		return false
	}
	return rand.Intn(100) < rate
}

// matchEasModelAlgoName resolves the algo name whose EAS url targets the job's
// model service and that is also present in rankAlgoNames. It returns "" if none
// matches. The url split is guarded to avoid a panic on malformed EAS urls.
func matchEasModelAlgoName(job *model.FeatureConsistencyJob, ctx *context.RecommendContext, rankAlgoNames []string) string {
	if ctx.Config == nil {
		return ""
	}
	for _, algoConfig := range ctx.Config.AlgoConfs {
		urls := strings.Split(algoConfig.EasConf.Url, "/api/predict/")
		if len(urls) < 2 {
			continue
		}
		if urls[1] != job.EasModelServiceName {
			continue
		}
		for _, rankAlgoName := range rankAlgoNames {
			if rankAlgoName == algoConfig.Name {
				return algoConfig.Name
			}
		}
	}
	return ""
}

// SkipFeatureReplyMeta builds the meta_data carried by the online scoring request
// when skip_feature_reply is hit. The model service uses these fields as OSS CSV
// columns, so they must stay aligned with the compare script's table schema.
// task_id is the current run's history id (not JobId): the model service lands CSV
// under {log_oss_path}/{task_id}/{uuid}.csv, and a per-run unique history id keeps
// each run's CSVs isolated instead of piling up under a stable JobId directory.
func SkipFeatureReplyMeta(user *module.User, ctx *context.RecommendContext, job *model.FeatureConsistencyJob) map[string]string {
	scene, _ := ctx.GetParameter("scene").(string)
	return map[string]string{
		"request_id":   ctx.RecommendId,
		"user_id":      string(user.Id),
		"scene":        scene,
		"request_time": strconv.FormatInt(time.Now().UnixMilli(), 10),
		"task_id":      strconv.Itoa(job.LatestHistoryJobId),
	}
}

// CloneEasyrecRequestWithDebug returns a deep copy of req carrying the given debug
// level and merged meta_data, leaving the original request untouched so other
// algos sharing the same batch request are not affected.
func CloneEasyrecRequestWithDebug(req *easyrec.PBRequest, level int, meta map[string]string) *easyrec.PBRequest {
	if req == nil {
		return nil
	}
	clone, ok := proto.Clone(req).(*easyrec.PBRequest)
	if !ok || clone == nil {
		return req
	}
	clone.DebugLevel = int32(level)
	if clone.MetaData == nil {
		clone.MetaData = make(map[string]string, len(meta))
	}
	for k, v := range meta {
		clone.MetaData[k] = v
	}
	return clone
}

// SkipFeatureReplyJobHit reports whether the given job was the one sampled in the
// synchronous scoring path, so the async backflow reuses the same decision instead
// of re-sampling. It is exported for reuse by the general_rank package.
func SkipFeatureReplyJobHit(ctx *context.RecommendContext, jobId int) bool {
	if ctx == nil {
		return false
	}
	if hitJobIds, ok := ctx.GetContextParam(featureConsistencySkipJobIdKey).(map[int]bool); ok {
		return hitJobIds[jobId]
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

			if job.SkipFeatureReply {
				// The online scoring request already carried debug_level=103, so the
				// model service writes OSS CSV directly. Reuse the sampling decision
				// made in the scoring path instead of re-sampling, and only backflow
				// the online-side data. No /api/feature_reply relay, no replay log sync.
				if SkipFeatureReplyJobHit(context, job.JobId) {
					log.Info(fmt.Sprintf("requestId=%s\tevent=logRankResultSkipFeatureReply\tname=%s", context.RecommendId, job.JobName))
					r.logRankResultToPaiConfigServer(user, items, context, job)
				}
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
				log.Info(fmt.Sprintf("requestId=%s\tevent=LogSampleResult\tname=%s", context.RecommendId, job.JobName))
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

		if i%50 == 0 {
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
				job.FeatureBackflowQueueDatahubProject, job.FeatureBackflowQueueDatahubTopic, "lz4", nil)
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
		algoData := algoGenerator.GeneratorAlgoDataDebugWithLevel(1, map[string]string{"request_id": context.RecommendId})
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
