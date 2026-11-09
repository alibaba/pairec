package recall

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/algorithm/eas"
	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

// genRecMultiValueSeparator joins multi-value features (CHR(29)); hist uses ';'.
const genRecMultiValueSeparator = "\x1d"

// GenRecRecall calls the EAS GenRec generation API, which returns SID beam
// results with mapped item ids directly, so no vector retrieval is needed.
type GenRecRecall struct {
	*BaseRecall
	genRecConf recconf.GenRecConfig
}

func NewGenRecRecall(config recconf.RecallConfig) *GenRecRecall {
	recall := &GenRecRecall{
		BaseRecall: NewBaseRecall(config),
		genRecConf: config.GenRecConf,
	}
	if recall.genRecConf.HistFeatureName == "" {
		recall.genRecConf.HistFeatureName = "hist"
	}
	return recall
}

func (r *GenRecRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()

	algoRet, err := algorithm.Run(r.recallAlgo, &eas.GenRecRequestData{
		Request: &easyrec.PBRequest{
			UserFeatures: r.encodeUserFeatures(user.MakeUserFeatures2(), context),
		},
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

// encodeUserFeatures converts user properties to PB features. Every feature is
// written as string_feature as required by the GenRec processor. With a
// whitelist, only the listed fields are encoded and missing ones are filled
// with empty strings; without it, all properties are encoded.
func (r *GenRecRecall) encodeUserFeatures(features map[string]interface{}, context *context.RecommendContext) map[string]*easyrec.PBFeature {
	userFeatures := make(map[string]*easyrec.PBFeature, len(features))
	if len(r.genRecConf.FeatureNames) > 0 {
		for _, name := range r.genRecConf.FeatureNames {
			value, ok := features[name]
			if !ok {
				userFeatures[name] = genRecStringFeature("")
				log.Warning(fmt.Sprintf("requestId=%s\tmodule=GenRecRecall\tname=%s\tmsg=feature missing, fill empty\tfield=%s",
					context.RecommendId, r.modelName, name))
				continue
			}
			r.encodeFeature(userFeatures, name, value, context)
		}
		return userFeatures
	}

	for name, value := range features {
		r.encodeFeature(userFeatures, name, value, context)
	}
	return userFeatures
}

// encodeFeature encodes one feature; values of unsupported types are skipped
// with a warning instead of being replaced by fabricated values.
func (r *GenRecRecall) encodeFeature(userFeatures map[string]*easyrec.PBFeature, name string, value interface{}, context *context.RecommendContext) {
	separator := genRecMultiValueSeparator
	if name == r.genRecConf.HistFeatureName {
		separator = ";"
	}
	str, ok := genRecEncodeValue(value, separator)
	if !ok {
		log.Warning(fmt.Sprintf("requestId=%s\tmodule=GenRecRecall\tname=%s\tmsg=skip unsupported feature type\tfield=%s\ttype=%T",
			context.RecommendId, r.modelName, name, value))
		return
	}
	if name == r.genRecConf.HistFeatureName {
		// FeatureStore returns multi-value strings joined by CHR(29), while
		// the GenRec hist field expects ';'.
		str = strings.ReplaceAll(str, genRecMultiValueSeparator, ";")
	}
	userFeatures[name] = genRecStringFeature(str)
}

func genRecStringFeature(value string) *easyrec.PBFeature {
	return &easyrec.PBFeature{Value: &easyrec.PBFeature_StringFeature{StringFeature: value}}
}

// genRecEncodeValue renders a scalar as a string and joins array elements with
// separator, preserving order and duplicates. Empty arrays become "".
func genRecEncodeValue(value interface{}, separator string) (string, bool) {
	if value == nil {
		return "", true
	}
	if data, ok := value.([]byte); ok {
		return string(data), true
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		items := make([]string, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			item, ok := genRecScalarString(rv.Index(i).Interface())
			if !ok {
				return "", false
			}
			items = append(items, item)
		}
		return strings.Join(items, separator), true
	case reflect.String:
		return rv.String(), true
	default:
		return genRecScalarString(value)
	}
}

func genRecScalarString(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case json.Number:
		return v.String(), true
	case int, int32, int64, uint32, float32, float64:
		return utils.ToString(v, ""), true
	default:
		return "", false
	}
}
