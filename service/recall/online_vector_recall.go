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
	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
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
	userFeatures := user.MakeUserFeatures2()
	algoGenerator.AddFeatures(nil, nil, userFeatures)
	algoData := algoGenerator.GeneratorAlgoData()
	easyrecRequest := algoData.GetFeatures().(*easyrec.PBRequest)
	easyrecRequest.FaissNeighNum = int32(r.recallCount)
	algoRet, err := algorithm.Run(r.recallAlgo, easyrecRequest)
	if context.Debug {
		go r.debugFeature(userFeatures, context)
	}
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

					if r.recallCount > 0 && len(ret) > r.recallCount {
						ret = ret[:r.recallCount]
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

func (r *OnlineVectorRecall) debugFeature(userFeatures map[string]any, context *context.RecommendContext) {
	newAlgoName := r.recallAlgo + "_debug_vector_algo"

	printFunc := func(data, featureType string) {
		size := len(data)
		for i := 0; i < size; {
			end := i + 4096
			if end >= size {
				end = size
			} else {
				for end > i {
					if data[end] == '|' {
						end++
						break
					}
					end--
				}
				if end == i {
					end = i + 4096
				}
			}
			log.Info(fmt.Sprintf("requestId=%s\t%s=%s", context.RecommendId, featureType, string(data[i:end])))
			i = end
		}

	}
	found := false
	for _, config := range context.Config.AlgoConfs {
		if config.Name == newAlgoName {
			found = true
			break
		}
	}
	if !found {
		var algoConfig recconf.AlgoConfig
		for _, config := range context.Config.AlgoConfs {
			if config.Name == r.recallAlgo {
				algoConfig = config
				// change algoname  and  response function name
				algoConfig.Name = newAlgoName
				if algoConfig.EasConf.ResponseFuncName != "" {
					algoConfig.EasConf.ResponseFuncName += "Debug"
				}
				algorithm.AddAlgoWithSign(algoConfig)
				break
			}
		}

	}
	algoGenerator := rank.CreateAlgoDataGenerator(r.recallAlgoType, nil)
	algoGenerator.SetItemFeatures(nil)
	algoGenerator.AddFeatures(nil, nil, userFeatures)
	algoData := algoGenerator.GeneratorAlgoData()
	easyrecRequest := algoData.GetFeatures().(*easyrec.PBRequest)
	easyrecRequest.DebugLevel = 1
	algoRet, err := algorithm.Run(newAlgoName, easyrecRequest)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=OnlineVectorRecall\tname=%s\terr=%v", context.RecommendId, r.modelName, err))
	} else {
		if result, ok := algoRet.([]response.AlgoResponse); ok && len(result) > 0 {
			if easyResponse, ok := result[0].(*eas.EasyrecResponse); ok {
				printFunc(easyResponse.RawFeatures, "RawFeatures")
				printFunc(easyResponse.GenerateFeatures.String(), "GenerateFeatures")
			}
		}
	}

}
