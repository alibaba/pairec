package easyrec

import (
	"bytes"
	"strings"
)

const (
	MULTI_SPLIT      = "|"
	HA3_MULTI_SPLIT  = "\u001D"
	FEA_LIST_MIN_LEN = 2
)

type EasyrecRequestBuilder struct {
	request   *PBRequest
	separator string
}

func NewEasyrecRequestBuilder() *EasyrecRequestBuilder {

	return &EasyrecRequestBuilder{
		request: &PBRequest{
			UserFeatures:    make(map[string]*PBFeature, 0),
			ContextFeatures: make(map[string]*ContextFeatures, 0),
			// DebugLevel:      int32(1),
		},
		separator: "\u0002",
	}
}
func NewEasyrecRequestBuilderDebug() *EasyrecRequestBuilder {

	builder := NewEasyrecRequestBuilder()
	builder.request.DebugLevel = int32(1)
	return builder
}

func NewEasyrecRequestBuilderDebugWithLevel(level int) *EasyrecRequestBuilder {

	builder := NewEasyrecRequestBuilder()
	builder.request.DebugLevel = int32(level)
	return builder
}

func (b *EasyrecRequestBuilder) EasyrecRequest() *PBRequest {
	return b.request
}

func (b *EasyrecRequestBuilder) AddUserFeature(k string, v interface{}) {
	switch val := v.(type) {
	case float32:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_FloatFeature{val}}
	case int32:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_IntFeature{val}}
	case int:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_IntFeature{int32(val)}}
	case int64:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_LongFeature{val}}
	case float64:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_FloatFeature{float32(val)}}
	case string:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_StringFeature{val}}
	default:
	}
}

func (b *EasyrecRequestBuilder) AddUserFeatureStr(featureStr string) {
	userFeas := strings.Split(featureStr, b.separator)

	for _, fea := range userFeas {
		if !strings.Contains(fea, ":") {
			continue
		}

		feaList := strings.Split(fea, ":")

		value := b.buildValue(feaList)

		b.request.UserFeatures[feaList[0]] = &PBFeature{Value: &PBFeature_StringFeature{value}}
	}
}

func (b *EasyrecRequestBuilder) AddItemId(itemId string) {
	b.request.ItemIds = append(b.request.ItemIds, itemId)
}

func (b *EasyrecRequestBuilder) AddItemIds(itemIdsStr string) {
	itemIds := strings.Split(itemIdsStr, ",")
	b.request.ItemIds = append(b.request.ItemIds, itemIds...)
}

func (b *EasyrecRequestBuilder) buildValue(feaList []string) string {

	if len(feaList) < FEA_LIST_MIN_LEN {
		return ""
	}

	joinStr := feaList[1]

	if len(feaList) > FEA_LIST_MIN_LEN {
		var buf bytes.Buffer

		for i := 1; i < len(feaList); i++ {
			buf.WriteString(feaList[i])
			if i < len(feaList)-1 {
				buf.WriteString(":")
			}
		}

		joinStr = buf.String()
	}

	if strings.Index(joinStr, MULTI_SPLIT) > 0 && strings.Index(joinStr, HA3_MULTI_SPLIT) > 0 {
		joinStr = strings.ReplaceAll(joinStr, HA3_MULTI_SPLIT, ",")
	}

	return joinStr
}
func (b *EasyrecRequestBuilder) AddContextFeature(key string, features []interface{}) {
	contextFeatures := &ContextFeatures{
		Features: make([]*PBFeature, 0),
	}

	for _, f := range features {
		switch val := f.(type) {
		case float32:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_FloatFeature{val}})
		case int32:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntFeature{val}})
		case int:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntFeature{int32(val)}})
		case int64:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_LongFeature{val}})
		case float64:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_FloatFeature{float32(val)}})
		case string:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringFeature{val}})
		default:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringFeature{""}})
		}
	}

	b.request.ContextFeatures[key] = contextFeatures
}
