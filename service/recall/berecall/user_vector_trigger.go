package berecall

import (
	"fmt"
	"sync"
	"time"

	"github.com/goburrow/cache"
	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/algorithm/eas"
	"github.com/alibaba/pairec/v2/algorithm/response"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/feature"
	"github.com/alibaba/pairec/v2/service/rank"
)

type UserVectorTrigger struct {
	features        []*feature.Feature
	recallAlgo      string
	recallAlgoType  string
	cachePrefix     string
	userVectorCache cache.Cache
}

func NewUserVectorTrigger(config *recconf.UserVectorTriggerConfig) *UserVectorTrigger {
	trigger := &UserVectorTrigger{
		recallAlgo:     config.RecallAlgo,
		recallAlgoType: eas.Eas_Processor_EASYREC,
		cachePrefix:    config.CachePrefix,
	}
	var features []*feature.Feature
	for _, conf := range config.UserFeatureConfs {
		f := feature.LoadWithConfig(conf)
		features = append(features, f)
	}

	trigger.features = features

	trigger.userVectorCache = cache.New(
		cache.WithMaximumSize(10000),
		cache.WithExpireAfterAccess(time.Duration(config.CacheTime+10)*time.Second),
	)

	return trigger
}
func (t *UserVectorTrigger) loadUserFeatures(user *module.User, context *context.RecommendContext) {
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
func (t *UserVectorTrigger) GetTriggerKey(user *module.User, context *context.RecommendContext) *TriggerResult {
	var userEmbedding string
	userEmbKey := t.cachePrefix + string(user.Id)
	if value, ok := t.userVectorCache.GetIfPresent(userEmbKey); ok {
		userEmbedding = value.(string)
		//user.AddProperty(r.modelName+"_embedding", userEmbedding)
	} else {
		t.loadUserFeatures(user, context)
		// second invoke eas model
		algoGenerator := rank.CreateAlgoDataGenerator(t.recallAlgoType, nil)
		algoGenerator.AddFeatures(nil, nil, user.MakeUserFeatures2())
		algoData := algoGenerator.GeneratorAlgoData()
		algoRet, err := algorithm.Run(t.recallAlgo, algoData.GetFeatures())
		if err != nil {
			context.LogError(fmt.Sprintf("requestId=%s\tmodule=UserVectorTrigger\terr=%v", context.RecommendId, err))
		} else {
			// eas model invoke success
			if result, ok := algoRet.([]response.AlgoResponse); ok && len(result) > 0 {
				if userEmbResponse, ok := result[0].(*eas.EasyrecUserEmbResponse); ok {
					userEmbedding = userEmbResponse.GetUserEmb()
					// user embedding put cache
					t.userVectorCache.Put(userEmbKey, userEmbedding)
				}
			}
		}
	}

	triggerResult := &TriggerResult{
		TriggerItem: userEmbedding,
	}
	return triggerResult
}
