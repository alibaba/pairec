package recallenginerecall

import (
	gocontext "context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/algorithm/eas"
	"github.com/alibaba/pairec/v2/algorithm/response"
	"github.com/alibaba/pairec/v2/context"
	plog "github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/feature"
	"github.com/alibaba/pairec/v2/service/rank"
)

type UserRealtimeEmbeddingTrigger struct {
	features       []*feature.Feature
	recallAlgo     string
	recallAlgoType string
	/*
		embeddingNum                 int
		datahub                      *datahub.Datahub
		featureConsistencyJobService *rank.FeatureConsistencyJobService
	*/
	useCacheFeatures bool
	debug            bool
}

func NewUserRealtimeEmbeddingTrigger(config *recconf.UserRealtimeEmbeddingTriggerConfig) *UserRealtimeEmbeddingTrigger {
	trigger := &UserRealtimeEmbeddingTrigger{
		recallAlgo:     config.RecallAlgo,
		recallAlgoType: eas.Eas_Processor_EASYREC,
		//embeddingNum:                 config.EmbeddingNum,
		debug: config.Debug,
		//featureConsistencyJobService: new(rank.FeatureConsistencyJobService),
	}
	var features []*feature.Feature
	for _, conf := range config.UserFeatureConfs {
		if conf.FeatureDaoConf.LoadFromCacheFeaturesName != "" {
			trigger.useCacheFeatures = true
		}
		f := feature.LoadWithConfig(conf)
		features = append(features, f)
	}

	trigger.features = features
	/*
		if trigger.debug {
			if datahubclient, err := datahub.GetDatahub(config.DebugLogDatahub); err == nil {
				trigger.datahub = datahubclient
			} else {
				plog.Error(fmt.Sprintf("get datahub error:%v", err))
			}
		}
	*/

	return trigger
}
func (t *UserRealtimeEmbeddingTrigger) loadUserFeatures(user *module.User, context *context.RecommendContext) {
	var wg sync.WaitGroup
	for _, fea := range t.features {
		wg.Add(1)
		go func(fea *feature.Feature) {
			defer wg.Done()
			fea.LoadFeatures(user, nil, context)
		}(fea)
	}

	wg.Wait()

}
func (t *UserRealtimeEmbeddingTrigger) GetTriggerKey(u *module.User, context *context.RecommendContext) *TriggerResult {
	//start := time.Now()
	if t.useCacheFeatures {
		ctx, cancel := gocontext.WithTimeout(gocontext.Background(), 150*time.Millisecond)
		defer cancel()
		select {
		case <-u.FeatureAsyncLoadCh():
		case <-ctx.Done():
			plog.Error(fmt.Sprintf("requestId=%s\tmodule=UserRealtimeEmbeddingTrigger\terror=%v", context.RecommendId, ctx.Err()))
		}
	}

	//user := u.Clone()
	user := u
	t.loadUserFeatures(user, context)
	//plog.Info(fmt.Sprintf("requestId=%s\tmodule=UserRealtimeEmbeddingTrigger_loadfeature\tcost=%v", context.RecommendId, utils.CostTime(start)))

	userFeatures := user.MakeUserFeatures2()
	algoGenerator := rank.CreateAlgoDataGenerator(t.recallAlgoType, nil)
	algoGenerator.SetItemFeatures(nil)
	algoGenerator.AddFeatures(nil, nil, userFeatures)
	//algoData := algoGenerator.GeneratorAlgoDataDebugWithLevel(102, map[string]string{"request_id": context.RecommendId})
	algoData := algoGenerator.GeneratorAlgoData()
	//easyrecRequest := algoData.GetFeatures().(*easyrec.PBRequest)
	//easyrecRequest.FaissNeighNum = int32(t.embeddingNum)
	algoRet, err := algorithm.Run(t.recallAlgo, algoData.GetFeatures())

	//var triggerItem string
	//var triggerItems []string
	var userEmbedding string

	if err != nil {
		plog.Error(fmt.Sprintf("requestId=%s\tmodule=UserRealtimeEmbeddingTrigger\terr=%v", context.RecommendId, err))
	} else {
		// eas model invoke success
		if result, ok := algoRet.([]response.AlgoResponse); ok && len(result) > 0 {
			if embeddingReponse, ok := result[0].(*eas.TorchrecEmbeddingResponse); ok {
				embeddings := embeddingReponse.GetEmbedding()
				embeddingList := make([]string, 0, len(embeddings))
				for _, embedding := range embeddings {
					embeddingList = append(embeddingList, strconv.FormatFloat(float64(embedding), 'f', -1, 32))
				}

				userEmbedding = strings.Join(embeddingList, ",")
			} else {
				plog.Error(fmt.Sprintf("requestId=%s\tmodule=UserRealtimeEmbeddingTrigger\terror=unexpected response type, expected *eas.TorchrecEmbeddingResponse, got %T", context.RecommendId, result[0]))
			}
			/*
				if response, ok := result[0].(*eas.EasyrecUserRealtimeEmbeddingResponse); ok {
					userEmbedding = response.GetUserEmbedding()
					for _, info := range response.GetEmbeddingList() {
						triggerItems = append(triggerItems, fmt.Sprintf("%s:%f", info.ItemId, info.Score))
					}

					triggerItem = strings.Join(triggerItems, ",")
				}
			*/
		}

	}
	if context.Debug {
		plog.Info(fmt.Sprintf("requestId=%s\tmodule=UserRealtimeEmbeddingTrigger\tuserEmbedding=%s", context.RecommendId, userEmbedding))
	}

	//go t.featureConsistencyJobService.LogRecallResult(user, nil, context, "dssm", userEmbedding, triggerItem, t.recallAlgo, t.recallAlgoType, "", "", "")

	triggerResult := &TriggerResult{
		TriggerItem: userEmbedding,
	}
	//plog.Info(fmt.Sprintf("requestId=%s\tmodule=UserRealtimeEmbeddingTrigger\tcost=%v", context.RecommendId, utils.CostTime(start)))
	return triggerResult
}
