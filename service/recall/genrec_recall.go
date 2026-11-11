package recall

import (
	"fmt"
	"time"

	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/algorithm/eas"
	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/rank"
	"github.com/alibaba/pairec/v2/utils"
)

// GenRecRecall calls the EAS GenRec generation API, which returns SID beam
// results with mapped item ids directly, so no vector retrieval is needed.
type GenRecRecall struct {
	*BaseRecall
	genRecConf recconf.GenRecConfig
}

func NewGenRecRecall(config recconf.RecallConfig) *GenRecRecall {
	return &GenRecRecall{
		BaseRecall: NewBaseRecall(config),
		genRecConf: config.GenRecConf,
	}
}

func (r *GenRecRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()

	// Features come from FeatureStore as typed slices, so the standard EasyRec
	// builder encodes them (arrays become PB lists) the same way as rank.
	algoGenerator := rank.CreateAlgoDataGenerator(eas.Eas_Processor_EASYREC, nil)
	algoGenerator.SetItemFeatures(nil)
	algoGenerator.AddFeatures(nil, nil, user.MakeUserFeatures2())
	algoData := algoGenerator.GeneratorAlgoData()
	request := algoData.GetFeatures().(*easyrec.PBRequest)

	algoRet, err := algorithm.Run(r.recallAlgo, &eas.GenRecRequestData{
		Request:      request,
		BeamWidth:    r.genRecConf.BeamWidth,
		SidCount:     r.genRecConf.SidCount,
		MaxNewTokens: r.genRecConf.MaxNewTokens,
		Rid:          context.RecommendId,
	})
	if err != nil {
		context.LogError(fmt.Sprintf("requestId=%s\tmodule=GenRecRecall\tname=%s\terr=%v", context.RecommendId, r.modelName, err))
		return
	}

	response, ok := algoRet.(*eas.GenRecResponse)
	if !ok {
		context.LogError(fmt.Sprintf("requestId=%s\tmodule=GenRecRecall\tname=%s\terr=unexpected response type %T", context.RecommendId, r.modelName, algoRet))
		return
	}

	// Beams arrive sorted by score. One SID may map to multiple items sharing
	// the same score; items repeated across beams keep their first occurrence.
	seen := make(map[module.ItemId]bool)
	for _, beam := range response.MetaInfo.BeamResults {
		for _, itemId := range beam.ItemIds {
			if itemId == "" {
				continue
			}
			id := module.ItemId(itemId)
			if seen[id] {
				continue
			}
			seen[id] = true
			item := module.NewItem(itemId)
			item.Score = beam.MetaInfo.SequenceScore
			item.ItemType = r.itemType
			item.RetrieveId = r.modelName
			ret = append(ret, item)
		}
	}

	if r.recallCount > 0 && len(ret) > r.recallCount {
		ret = ret[:r.recallCount]
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=GenRecRecall\tname=%s\tcount=%d\tcost=%d",
		context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return
}
