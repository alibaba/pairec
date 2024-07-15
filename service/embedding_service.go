package service

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/algorithm/eas"
	"github.com/alibaba/pairec/v2/algorithm/response"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/datahub"
	"github.com/alibaba/pairec/v2/log"
	plog "github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/rank"
	"github.com/alibaba/pairec/v2/utils"
)

var (
	User_Embedding_Module = "user"
	Item_Embedding_Module = "item"
)

type EmbeddingService struct {
	RecommendService
	module          string // user or item
	contextFeatures []string
}

func NewEmbeddingService() *EmbeddingService {
	service := EmbeddingService{}
	return &service
}

func (r *EmbeddingService) Recommend(context *context.RecommendContext) ([]float32, error) {
	var (
		user  *module.User
		items []*module.Item
	)

	uid := context.GetParameter("uid")
	if uid == nil || uid.(string) == "" {
		user = module.NewUserWithContext("default_user", context)
	} else {
		userId := r.GetUID(context)
		user = module.NewUserWithContext(userId, context)
		features := context.GetParameter("user_features")
		if features != nil {
			user.AddProperties(features.(map[string]any))
		}
		r.module = User_Embedding_Module
	}

	item_id := context.GetParameter("item_id")
	if item_id != nil && item_id.(string) != "" {
		item := module.NewItem(item_id.(string))
		features := context.GetParameter("item_features")
		if features != nil {
			item.AddProperties(features.(map[string]any))
		}
		for k := range features.(map[string]any) {
			r.contextFeatures = append(r.contextFeatures, k)
		}
		items = append(items, item)
		r.module = Item_Embedding_Module
	}

	embeddings, err := r.Rank(user, items, context)
	go r.recordLog(user, items, context, embeddings)

	return embeddings, err
}

func (r *EmbeddingService) Rank(user *module.User, items []*module.Item, context *context.RecommendContext) (embeddings []float32, err error) {
	start := time.Now()
	if context.Debug {
		data, _ := json.Marshal(user)
		fmt.Printf("requestId=%s\tuser=%s\n", context.RecommendId, string(data))
		data, _ = json.Marshal(items)
		fmt.Printf("requestId=%s\titems=%s\n", context.RecommendId, string(data))
	}

	var rankConfig recconf.RankConfig
	scene_name := context.GetParameter("scene").(string)
	embeddingConfig, ok := context.Config.EmbeddingConfs[scene_name]

	if ok {
		rankConfig = embeddingConfig.RankConf
	}

	if len(rankConfig.RankAlgoList) == 0 {
		return
	}

	algoGenerator := rank.CreateAlgoDataGenerator(rankConfig.Processor, r.contextFeatures)

	if r.module == User_Embedding_Module {
		userFeatures := user.MakeUserFeatures2()
		algoGenerator.AddFeatures(nil, nil, userFeatures)
	} else if r.module == Item_Embedding_Module {
		for _, item := range items {
			features := item.GetFeatures()
			algoGenerator.AddFeatures(item, features, nil)
		}
	}

	algoData := algoGenerator.GeneratorAlgoData()

	var wg sync.WaitGroup
	for _, algoName := range rankConfig.RankAlgoList {
		wg.Add(1)
		go func(algo string) {
			defer wg.Done()

			// run 返回原始的值，然后处理返回数据// 注册配置
			ret, err := algorithm.Run(algo, algoData.GetFeatures())
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\terror=run algorithm error(%v)", context.RecommendId, err))
				algoData.SetError(err)
			} else {
				if result, ok := ret.([]response.AlgoResponse); ok {
					algoData.SetAlgoResult(algo, result)
				}
			}

		}(algoName)

	}

	wg.Wait()
	if algoData.Error() != nil {
		return nil, algoData.Error()
	}

	for _, algoResults := range algoData.GetAlgoResult() {
		if len(algoResults) > 0 {
			if embeddingReponse, ok := algoResults[0].(*eas.TorchrecEmbeddingResponse); ok {
				embeddings = embeddingReponse.GetEmbedding()
			}
		}
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("embeddings is empty")
	}

	log.Info(fmt.Sprintf("requestId=%s\tmodule=EmbeddingRank\tcost=%d", context.RecommendId, utils.CostTime(start)))
	return
}

func (r *EmbeddingService) recordLog(user *module.User, items []*module.Item, context *context.RecommendContext, embeddings []float32) {
	scene_name := context.GetParameter("scene").(string)
	embeddingConfig, ok := context.Config.EmbeddingConfs[scene_name]
	if !ok {
		return
	}
	if embeddingConfig.DataSource.Name == "" || embeddingConfig.DataSource.Type == "" {
		return
	}

	log := make(map[string]any)
	log["request_id"] = context.RecommendId
	log["scene"] = scene_name
	log["request_time"] = time.Now().Unix()
	log["module"] = r.module

	if r.module == User_Embedding_Module {
		log["user_id"] = string(user.Id)
		features := user.MakeUserFeatures2()
		j, _ := json.Marshal(features)
		log["user_features"] = string(j)
	} else if r.module == Item_Embedding_Module {
		log["item_id"] = string(items[0].Id)

		features := items[0].GetFeatures()
		j, _ := json.Marshal(features)
		log["item_features"] = string(j)
	}

	j, _ := json.Marshal(embeddings)
	log["embeddings"] = string(j)
	var err error
	if embeddingConfig.DataSource.Type == recconf.DataSource_Type_Datahub {
		err = r.recordToDatahub(embeddingConfig.DataSource.Name, []map[string]any{log})
	}
	if err != nil {
		plog.Error(fmt.Sprintf("requestId=%s\tmodule=recordLog\terror=%v", context.RecommendId, err))
	}
}
func (r *EmbeddingService) recordToDatahub(name string, messages []map[string]interface{}) error {
	p, error := datahub.GetDatahub(name)
	if error != nil {
		return error
	}

	p.SendMessage(messages)
	return nil
}
