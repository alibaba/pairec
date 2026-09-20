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

	user := u.Clone()
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

	if err != nil {
		return &TriggerResult{Err: fmt.Errorf("user embedding inference: %w", err)}
	}
	result, ok := algoRet.([]response.AlgoResponse)
	if !ok || len(result) == 0 {
		return &TriggerResult{Err: fmt.Errorf("user embedding response is empty or invalid")}
	}
	embedding, ok := result[0].(*eas.TorchrecEmbeddingResponse)
	if !ok || embedding == nil {
		return &TriggerResult{Err: fmt.Errorf("unexpected user embedding response type %T", result[0])}
	}
	passThrough := embedding.GetPassThroughData()
	version := strings.TrimSpace(passThrough["version"])
	if version == "" {
		version = strings.TrimSpace(passThrough["model_version"])
	}
	if queries := embedding.GetInterests(); queries != nil {
		return &TriggerResult{Queries: queries, Version: version}
	}
	values := embedding.GetEmbedding()
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.FormatFloat(float64(value), 'f', -1, 32))
	}
	return &TriggerResult{TriggerItem: strings.Join(parts, ","), Version: version}
}
