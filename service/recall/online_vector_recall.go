package recall

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goburrow/cache"

	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/algorithm/eas"
	"github.com/alibaba/pairec/v2/algorithm/response"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/feature"
	"github.com/alibaba/pairec/v2/service/rank"
	"github.com/alibaba/pairec/v2/utils"
)

const (
	VectorAlgoType_EasyRec      = "easyrec"
	VectorAlgoType_TorchRec_TDM = "torchrec_tdm"
	VectorAlgoType_TorchRec     = "torchrec_vector"
)

type OnlineVectorRecall struct {
	*BaseRecall
	features        []*feature.Feature
	recallAlgoType  string
	vectorAlgoType  string
	userVectorCache cache.Cache
}

func NewOnlineVectorRecall(config recconf.RecallConfig) *OnlineVectorRecall {
	recall := &OnlineVectorRecall{
		BaseRecall:     NewBaseRecall(config),
		recallAlgoType: eas.Eas_Processor_EASYREC,
		vectorAlgoType: config.VectorAlgoType,
	}

	if recall.cacheTime <= 0 && recall.cache != nil {
		recall.cacheTime = 1800
	}

	var features []*feature.Feature
	for _, conf := range config.UserFeatureConfs {
		f := feature.LoadWithConfig(conf)
		features = append(features, f)
	}

	recall.features = features
	return recall
}

func (r *OnlineVectorRecall) loadUserFeatures(user *module.User, context *context.RecommendContext) {
	var wg sync.WaitGroup
	for _, fea := range r.features {
		wg.Add(1)
		go func(fea *feature.Feature) {
			defer wg.Done()
			fea.LoadFeatures(user, nil, context)
		}(fea)
	}

	wg.Wait()

}
func (r *OnlineVectorRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()
	if r.cache != nil {
		key := r.cachePrefix + string(user.Id)
		cacheRet := r.cache.Get(key)
		if itemStr, ok := cacheRet.([]uint8); ok {
			itemIds := strings.Split(string(itemStr), ",")
			for _, id := range itemIds {
				var item *module.Item
				if strings.Contains(id, ":") {
					vars := strings.Split(id, ":")
					item = module.NewItem(vars[0])
					f, _ := strconv.ParseFloat(vars[1], 64)
					item.Score = f
				} else {
					item = module.NewItem(id)
				}

				item.RetrieveId = r.modelName
				ret = append(ret, item)
			}
			context.LogInfo(fmt.Sprintf("module=OnlineVectorRecall\tname=%s\thit cache\tcount=%d\tcost=%d",
				r.modelName, len(ret), utils.CostTime(start)))
			return
		}
	}

	r.loadUserFeatures(user, context)
	// second invoke eas model
	algoGenerator := rank.CreateAlgoDataGenerator(r.recallAlgoType, nil)
	algoGenerator.SetItemFeatures(nil)
	algoGenerator.AddFeatures(nil, nil, user.MakeUserFeatures2())
	algoData := algoGenerator.GeneratorAlgoData()
	algoRet, err := algorithm.Run(r.recallAlgo, algoData.GetFeatures())
	if err != nil {
		context.LogError(fmt.Sprintf("requestId=%s\tmodule=OnlineVectorRecall\tname=%s\terr=%v", context.RecommendId, r.modelName, err))
	} else {
		// eas model invoke success
		if result, ok := algoRet.([]response.AlgoResponse); ok && len(result) > 0 {
			if r.vectorAlgoType == VectorAlgoType_TorchRec_TDM || r.vectorAlgoType == VectorAlgoType_TorchRec {
				if userEmbResponse, ok := result[0].(*eas.TorchrecEmbeddingItemsResponse); ok {
					embeddingInfos := userEmbResponse.GetEmbeddingItems()
					ret = make([]*module.Item, 0, len(embeddingInfos))

					for _, info := range embeddingInfos {
						item := module.NewItem(info.ItemId)
						item.Score = info.Score
						item.RetrieveId = r.modelName
						ret = append(ret, item)
					}
				}
			}

		}
	}

	if r.cache != nil && len(ret) > 0 {
		go func() {
			key := r.cachePrefix + string(user.Id)
			var itemIds string
			for _, item := range ret {
				itemIds += fmt.Sprintf("%s:%v", string(item.Id), item.Score) + ","
			}
			itemIds = itemIds[:len(itemIds)-1]
			if err2 := r.cache.Put(key, itemIds, time.Duration(r.cacheTime)*time.Second); err2 != nil {
				context.LogError(fmt.Sprintf("requestId=%s\tmodule=OnlineVectorRecall\terror=%v", context.RecommendId, err2))
			}
		}()
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=OnlineVectorRecall\tname=%s\tcount=%d\tcost=%d",
		context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return
}
