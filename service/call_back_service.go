package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/algorithm/eas"
	"github.com/alibaba/pairec/v2/algorithm/response"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/datahub"
	"github.com/alibaba/pairec/v2/datasource/kafka"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/debug"
	"github.com/alibaba/pairec/v2/service/feature"
	"github.com/alibaba/pairec/v2/service/rank"
	"github.com/alibaba/pairec/v2/utils"
)

// processorWriteCallbackLogDebugLevel is the debug_level that makes the easyrec
// processor write the callback feature log on its own side. A request using it
// must reach the processor, otherwise no log is produced at all.
const processorWriteCallbackLogDebugLevel = 6

type CallBackService struct {
	recallService      *RecallService
	featureService     *feature.FeatureService
	userFeatureService *feature.UserFeatureService
	User               *module.User
	Items              []*module.Item
}

func NewCallBackService() *CallBackService {
	services := CallBackService{
		recallService:      &RecallService{},
		featureService:     feature.DefaultFeatureService(),
		userFeatureService: feature.DefaultUserFeatureService(),
	}
	return &services
}
func (r *CallBackService) LoadUserFeatures(context *context.RecommendContext) {
	//user feature prefetch
	r.userFeatureService.LoadUserFeaturesForCallback(r.User, context)
}
func (r *CallBackService) Recommend(context *context.RecommendContext) {
	scene_name := context.GetParameter("scene").(string)
	if _, ok := context.Config.SceneConfs[scene_name]; ok {
		items := r.recallService.GetItems(r.User, context)
		itemMap := make(map[module.ItemId]*module.Item, len(items))
		for _, item := range items {
			itemMap[item.Id] = item
		}

		for _, item := range r.Items {
			if itemM, ok := itemMap[item.Id]; ok {
				item.RetrieveId = itemM.RetrieveId
			}
		}

		go func() {
			debugService := debug.NewDebugService(r.User, context)
			debugService.WriteRecallLog(r.User, items, context)
		}()
	}
}

func (r *CallBackService) LoadFeatures(context *context.RecommendContext) {
	// load user features
	r.featureService.LoadFeatures(r.User, r.Items, context)
}

func (r *CallBackService) RecordLog(context *context.RecommendContext, msg string) error {
	scene_name := context.GetParameter("scene").(string)
	scene_name = strings.ReplaceAll(scene_name, "_callback", "")
	if callBackConfig, ok := context.Config.CallBackConfs[scene_name]; ok {
		data_source := callBackConfig.DataSource
		if data_source.Type == recconf.DataSource_Type_Kafka {
			return r.RecordToKafka(data_source.Name, msg)
		} else if data_source.Type == recconf.DataSource_Type_Datahub {
			return r.RecordToDatahub(data_source.Name, msg)
		}
	}

	return nil
}

func (r *CallBackService) RecordToKafka(kafka_name string, msg string) error {
	p, error := kafka.GetKafkaProducer(kafka_name)
	if error != nil {
		return error
	}
	p.SendMessage([]byte(msg))
	return nil
}

func (r *CallBackService) RecordToDatahub(name string, msg string) error {
	p, error := datahub.GetDatahub(name)
	if error != nil {
		return error
	}
	p.SendMessage([]map[string]interface{}{{"callback_log": msg}})
	return nil
}

func (r *CallBackService) RecordLogList(context *context.RecommendContext, messages []map[string]interface{}) error {
	scene_name := context.GetParameter("scene").(string)
	scene_name = strings.ReplaceAll(scene_name, "_callback", "")
	if callBackConfig, ok := context.Config.CallBackConfs[scene_name]; ok {
		data_source := callBackConfig.DataSource
		if data_source.Type == recconf.DataSource_Type_Kafka {
			return r.RecordToKafkaList(data_source.Name, messages)
		} else if data_source.Type == recconf.DataSource_Type_Datahub {
			return r.RecordToDatahubList(data_source.Name, messages)
		}
	}

	return nil
}

func (r *CallBackService) RecordToKafkaList(kafka_name string, messages []map[string]interface{}) error {
	p, error := kafka.GetKafkaProducer(kafka_name)
	if error != nil {
		return error
	}
	for _, msg := range messages {
		j, _ := json.Marshal(msg)
		p.SendMessage(j)
	}
	return nil
}

func (r *CallBackService) RecordToDatahubList(name string, messages []map[string]interface{}) error {
	p, error := datahub.GetDatahub(name)
	if error != nil {
		return error
	}

	p.SendMessage(messages)
	return nil
}
func (r *CallBackService) Rank(context *context.RecommendContext) {
	var rankConfig recconf.RankConfig
	scene_name := context.GetParameter("scene").(string)
	scene_name = strings.ReplaceAll(scene_name, "_callback", "")
	callBackConfig, ok := context.Config.CallBackConfs[scene_name]

	if ok {
		rankConfig = callBackConfig.RankConf
	}

	if len(rankConfig.RankAlgoList) == 0 {
		return
	}

	start := time.Now()

	rankItems := r.Items

	algoGenerator := rank.CreateAlgoDataGenerator(rankConfig.Processor, rankConfig.ContextFeatures)

	var userFeatures map[string]interface{}

	if rankConfig.Processor == eas.Eas_Processor_EASYREC {
		userFeatures = r.User.MakeUserFeatures2()
		algoGenerator.SetItemFeatures(rankConfig.ItemFeatures)
		algoGenerator.(*rank.EasyrecAlgoDataGenerator).SetUserId(string(r.User.Id))
	} else {
		userFeatures = r.User.MakeUserFeatures()
	}

	for _, item := range rankItems {
		features := item.GetFeatures()
		algoGenerator.AddFeatures(item, features, userFeatures)
	}

	var algoData rank.IAlgoData
	debugLevel := 3
	if callBackConfig.DebugLevel > 0 {
		debugLevel = callBackConfig.DebugLevel
	}

	writeRawFeatrues := false
	if callBackConfig.RawFeatures && callBackConfig.RawFeaturesRate > 0 {
		if rand.Intn(100) < callBackConfig.RawFeaturesRate {
			debugLevel = 1
			writeRawFeatrues = true
		}
	}

	// carry the request id in the pb meta data, so the processor side log can be
	// joined back to this callback request, same as the rank and recall paths do
	meta := map[string]string{"request_id": context.RecommendId}
	if algoGenerator.HasFeatures() {
		// the processor writes the log itself at this level, so it has to be sent
		// through even for a debug request
		if context.Debug && debugLevel != processorWriteCallbackLogDebugLevel {
			algoData = algoGenerator.GeneratorAlgoDataDebugWithLevel(2, meta)
		} else {
			algoData = algoGenerator.GeneratorAlgoDataDebugWithLevel(debugLevel, meta)
		}
	}

	// Explicit guard: when no item carried features (e.g. r.Items is empty
	// or every item returned an empty feature map), algoGenerator.HasFeatures
	// is false and algoData stays nil. The goroutine loop below dereferences
	// algoData, so without this guard we hit a nil pointer panic at
	// algoData.GetFeatures(). Returning here also avoids spawning goroutines
	// that have no work to do.
	//
	// With the upstream guards in callback_hook.go and web.SendDirect, this
	// branch is only reachable when a future caller bypasses both, i.e.
	// misuse, so log at Warning level for ops triage.
	if algoData == nil {
		log.Warning(fmt.Sprintf("requestId=%s\tmodule=callback\tevent=Rank\tmsg=algoData is nil, skipping rank fan-out", context.RecommendId))
		return
	}

	// 主模型 = RankAlgoList[0]，独占老三字段（raw_features / generate_features /
	// context_features）的写入权。多模型并行打分时，如果每个 goroutine 都往同
	// 名 key 上 AddProperty，Properties[key]=value 会互相整段覆盖，最终哪个
	// 模型胜出取决于 goroutine 完成先后，日志字段甚至可能来自不同模型（例如
	// raw_features 是 A 的、generate_features 是 B 的）。收敛到主模型独占后，
	// 单模型用户 byte-for-byte 零感知，多模型用户日志结果确定。
	//
	// 非主模型的特征以 <algo>_raw_features / <algo>_generate_features /
	// <algo>_context_features 带前缀 key 写入 Properties。web/callback_controller.go
	// 侧不会 delete 这些前缀 key，它们会随 json.Marshal(item.GetFeatures())
	// 自然序列化进已有的 item_features 字段，datahub schema 零变更。
	// 主模型调用失败时采取严格语义：老三字段留空、不用其他模型兜底，
	// 避免重新引入不确定性。
	primaryAlgo := rankConfig.RankAlgoList[0]

	var wg sync.WaitGroup
	for _, algoName := range rankConfig.RankAlgoList {
		wg.Add(1)
		go func(algo string) {
			defer wg.Done()

			newAlgoName := algo + "_callback"
			found := false
			var processor string
			for _, config := range context.Config.AlgoConfs {
				if config.Name == newAlgoName {
					found = true
					processor = config.EasConf.Processor
					break
				}
			}
			if !found {
				var algoConfig recconf.AlgoConfig
				for _, config := range context.Config.AlgoConfs {
					if config.Name == algo {
						algoConfig = config
						processor = config.EasConf.Processor
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
			// run 返回原始的值，然后处理返回数据// 注册配置
			ret, err := algorithm.Run(newAlgoName, algoData.GetFeatures())
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\terror=run algorithm error(%v)", context.RecommendId, err))
				algoData.SetError(err)
			} else {
				if result, ok := ret.([]response.AlgoResponse); ok {
					algoData.SetAlgoResult(algo, result)
					if processor == eas.Eas_Processor_EASYREC {
						itemList := algoData.GetItems()
						isPrimary := algo == primaryAlgo
						for j := 0; j < len(result) && j < len(itemList); j++ {
							switch response := result[j].(type) {
							case *eas.EasyrecResponse:
								writeEasyrecCallbackFeatures(itemList[j], algo, isPrimary, writeRawFeatrues,
									response.RawFeatures, response.GenerateFeatures, response.ContextFeatures)
							case *eas.EasyrecClassificationResponse:
								// EasyrecClassificationResponse 内嵌 EasyrecResponse，
								// 三段特征字段类型完全一致，共用同一个写入函数。
								writeEasyrecCallbackFeatures(itemList[j], algo, isPrimary, writeRawFeatrues,
									response.RawFeatures, response.GenerateFeatures, response.ContextFeatures)
							default:
								log.Error(fmt.Sprintf("requestId=%s\tmodule=callback\terror=algo response type(%T) not supported", context.RecommendId, response))
							}
						}
					}
				}
			}

		}(algoName)

	}

	wg.Wait()

	log.Info(fmt.Sprintf("requestId=%s\tmodule=CallBackRank\tcost=%d", context.RecommendId, utils.CostTime(start)))
}

// writeEasyrecCallbackFeatures 把一个 EasyRec 响应的三段特征写入 item.Properties，
// 服务于 CallBackService.Rank 的多模型并行打分场景。
//
// 主模型（isPrimary=true）写老三个 key：raw_features / generate_features /
// context_features。generate_features 必须继续以 *bytes.Buffer 指针形态存储，
// 因为 web/callback_controller.go 侧依赖 buf.(*bytes.Buffer).String() 断言把它
// 读出来落到 datahub 老三列之一；改成 string 会触发类型断言 panic。
//
// 非主模型写带 <algo>_ 前缀的 key：controller 不 delete 这些前缀 key，它们
// 会随 json.Marshal(item.GetFeatures()) 序列化进已有的 item_features 字段。
// 由于 *bytes.Buffer 未实现 MarshalJSON、内部字段也不导出，直接放进 Properties
// 序列化后是空对象，所以带前缀路径上必须先 .String() 转成 string 再存。
//
// writeRawFeatrues 是 CallBackConfig.RawFeatures + RawFeaturesRate 的采样开关，
// 同时约束老 key 与带前缀 key 的 raw_features 写入，防止采样语义失效导致
// item_features JSON 无脑膨胀。generate_features / context_features 不受此
// 开关约束，保持与升级前一致的行为。
func writeEasyrecCallbackFeatures(item *module.Item, algo string, isPrimary, writeRawFeatrues bool,
	rawFeatures string, generateFeatures *bytes.Buffer, contextFeatures string) {
	if item == nil {
		return
	}
	if isPrimary {
		if writeRawFeatrues {
			item.AddProperty("raw_features", rawFeatures)
		}
		item.AddProperty("generate_features", generateFeatures)
		item.AddProperty("context_features", contextFeatures)
		return
	}
	if writeRawFeatrues {
		item.AddProperty(algo+"_raw_features", rawFeatures)
	}
	genStr := ""
	if generateFeatures != nil {
		genStr = generateFeatures.String()
	}
	item.AddProperty(algo+"_generate_features", genStr)
	item.AddProperty(algo+"_context_features", contextFeatures)
}
