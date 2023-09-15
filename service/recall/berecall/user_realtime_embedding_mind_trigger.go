package berecall

import (
	gocontext "context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/algorithm"
	"github.com/alibaba/pairec/algorithm/eas"
	"github.com/alibaba/pairec/algorithm/eas/easyrec"
	"github.com/alibaba/pairec/algorithm/response"
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/datasource/datahub"
	plog "github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/service/feature"
	"github.com/alibaba/pairec/service/rank"
)

type UserRealtimeEmbeddingMindTrigger struct {
	debug                        bool
	features                     []*feature.Feature
	recallAlgo                   string
	recallAlgoType               string
	distinctParamName            string
	distinctParamValue           string
	embeddingNum                 int
	datahub                      *datahub.Datahub
	useCacheFeatures             bool
	featureConsistencyJobService *rank.FeatureConsistencyJobService
}

func NewUserRealtimeEmbeddingMindTrigger(config *recconf.UserRealtimeEmbeddingTriggerConfig) *UserRealtimeEmbeddingMindTrigger {
	trigger := &UserRealtimeEmbeddingMindTrigger{
		recallAlgo:                   config.RecallAlgo,
		recallAlgoType:               eas.Eas_Processor_EASYREC,
		embeddingNum:                 config.EmbeddingNum,
		debug:                        config.Debug,
		featureConsistencyJobService: new(rank.FeatureConsistencyJobService),
		distinctParamValue:           config.DistinctParamValue,
		distinctParamName:            config.DistinctParamName,
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

	if trigger.debug {
		if datahubclient, err := datahub.GetDatahub(config.DebugLogDatahub); err == nil {
			trigger.datahub = datahubclient
		} else {
			plog.Error(fmt.Sprintf("get datahub error:%v", err))
		}
	}

	return trigger
}
func (t *UserRealtimeEmbeddingMindTrigger) loadUserFeatures(user *module.User, context *context.RecommendContext) {
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
func (t *UserRealtimeEmbeddingMindTrigger) GetTriggerKey(u *module.User, context *context.RecommendContext) *TriggerResult {
	//start := time.Now()
	if t.useCacheFeatures {
		ctx, cancel := gocontext.WithTimeout(gocontext.Background(), 150*time.Millisecond)
		defer cancel()
		select {
		case <-u.FeatureAsyncLoadCh():
		case <-ctx.Done():
			plog.Error(fmt.Sprintf("requestId=%s\tmodule=UserRealtimeEmbeddingMindTrigger\terror=%v", context.RecommendId, ctx.Err()))
		}
	}
	user := u.Clone()
	t.loadUserFeatures(user, context)
	//plog.Info(fmt.Sprintf("requestId=%s\tmodule=UserRealtimeEmbeddingMindTrigger_loadfeature\tcost=%v", context.RecommendId, utils.CostTime(start)))

	algoGenerator := rank.CreateAlgoDataGenerator(t.recallAlgoType, nil)
	algoGenerator.AddFeatures(nil, nil, user.MakeUserFeatures2())
	algoData := algoGenerator.GeneratorAlgoDataDebugWithLevel(102)
	easyrecRequest := algoData.GetFeatures().(*easyrec.PBRequest)
	easyrecRequest.FaissNeighNum = int32(t.embeddingNum)
	algoRet, err := algorithm.Run(t.recallAlgo, easyrecRequest)

	var triggerItem string
	var triggerItems []string
	var userEmbedding string
	if err != nil {
		plog.Error(fmt.Sprintf("requestId=%s\tmodule=UserRealtimeEmbeddingMindTrigger\terr=%v", context.RecommendId, err))
	} else {
		// eas model invoke success
		if result, ok := algoRet.([]response.AlgoResponse); ok && len(result) > 0 {
			if response, ok := result[0].(*eas.EasyrecUserRealtimeEmbeddingMindResponse); ok {
				userEmbedding = response.GetUserEmbedding()
				triggerItems = make([]string, 0, response.GetDimSize())
				for i, info := range response.GetEmbeddingList() {
					triggerItems = append(triggerItems, fmt.Sprintf("%s:%f", info.ItemId, info.Score))
					if (i+1)%response.GetDimSize() == 0 {
						triggerItem += strings.Join(triggerItems, ",") + "|"
						triggerItems = triggerItems[:0]
					}
				}

				if len(triggerItems) > 0 {
					triggerItem += strings.Join(triggerItems, ",") + "|"
				}

				triggerItem = triggerItem[:len(triggerItem)-1]
			}
		}

	}
	go t.featureConsistencyJobService.LogRecallResult(user, nil, context, "mind", userEmbedding, triggerItem, t.recallAlgo, t.recallAlgoType, "", "", "")
	triggerResult := &TriggerResult{
		TriggerItem:       triggerItem,
		DistinctParam:     t.distinctParamValue,
		DistinctParamName: t.distinctParamName,
	}
	//plog.Info(fmt.Sprintf("requestId=%s\tmodule=UserRealtimeEmbeddingMindTrigger\tcost=%v", context.RecommendId, utils.CostTime(start)))
	return triggerResult
}
