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
			ItemFeatures:    make(map[string]*ContextFeatures, 0),
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
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_DoubleFeature{val}}
	case string:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_StringFeature{val}}
	case map[int64]string:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_LongStringMap{LongStringMap: &LongStringMap{MapField: val}}}
	case map[int64]int32:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_LongIntMap{LongIntMap: &LongIntMap{MapField: val}}}
	case map[int64]int:
		values := make(map[int64]int32, len(val))
		for k, v := range val {
			values[k] = int32(v)
		}
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_LongIntMap{LongIntMap: &LongIntMap{MapField: values}}}
	case map[int64]int64:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_LongLongMap{LongLongMap: &LongLongMap{MapField: val}}}
	case map[int64]float32:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_LongFloatMap{LongFloatMap: &LongFloatMap{MapField: val}}}
	case map[int64]float64:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_LongDoubleMap{LongDoubleMap: &LongDoubleMap{MapField: val}}}
	case map[string]string:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_StringStringMap{StringStringMap: &StringStringMap{MapField: val}}}
	case map[string]int32:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_StringIntMap{StringIntMap: &StringIntMap{MapField: val}}}
	case map[string]int:
		values := make(map[string]int32, len(val))
		for k, v := range val {
			values[k] = int32(v)
		}
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_StringIntMap{StringIntMap: &StringIntMap{MapField: values}}}
	case map[string]int64:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_StringLongMap{StringLongMap: &StringLongMap{MapField: val}}}
	case map[string]float32:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_StringFloatMap{StringFloatMap: &StringFloatMap{MapField: val}}}
	case map[string]float64:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_StringDoubleMap{StringDoubleMap: &StringDoubleMap{MapField: val}}}
	case map[int32]string:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_IntStringMap{IntStringMap: &IntStringMap{MapField: val}}}
	case map[int32]int32:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_IntIntMap{IntIntMap: &IntIntMap{MapField: val}}}
	case map[int32]int64:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_IntLongMap{IntLongMap: &IntLongMap{MapField: val}}}
	case map[int32]float32:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_IntFloatMap{IntFloatMap: &IntFloatMap{MapField: val}}}
	case map[int32]float64:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_IntDoubleMap{IntDoubleMap: &IntDoubleMap{MapField: val}}}
	case map[int]string:
		values := make(map[int32]string, len(val))
		for k, v := range val {
			values[int32(k)] = v
		}
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_IntStringMap{IntStringMap: &IntStringMap{MapField: values}}}
	case map[int]int32:
		values := make(map[int32]int32, len(val))
		for k, v := range val {
			values[int32(k)] = v
		}
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_IntIntMap{IntIntMap: &IntIntMap{MapField: values}}}
	case map[int]int:
		values := make(map[int32]int32, len(val))
		for k, v := range val {
			values[int32(k)] = int32(v)
		}
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_IntIntMap{IntIntMap: &IntIntMap{MapField: values}}}
	case map[int]int64:
		values := make(map[int32]int64, len(val))
		for k, v := range val {
			values[int32(k)] = v
		}
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_IntLongMap{IntLongMap: &IntLongMap{MapField: values}}}
	case map[int]float32:
		values := make(map[int32]float32, len(val))
		for k, v := range val {
			values[int32(k)] = v
		}
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_IntFloatMap{IntFloatMap: &IntFloatMap{MapField: values}}}
	case map[int]float64:
		values := make(map[int32]float64, len(val))
		for k, v := range val {
			values[int32(k)] = v
		}
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_IntDoubleMap{IntDoubleMap: &IntDoubleMap{MapField: values}}}
	case []int32:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_IntList{IntList: &IntList{Features: val}}}
	case []int:
		values := make([]int32, len(val))
		for i, v := range val {
			values[i] = int32(v)
		}
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_IntList{IntList: &IntList{Features: values}}}
	case []int64:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_LongList{LongList: &LongList{Features: val}}}
	case []string:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_StringList{StringList: &StringList{Features: val}}}
	case []float32:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_FloatList{FloatList: &FloatList{Features: val}}}
	case []float64:
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_DoubleList{DoubleList: &DoubleList{Features: val}}}
	case [][]int32:
		values := make([]*IntList, len(val))
		for i, v := range val {
			values[i] = &IntList{Features: v}
		}
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_IntLists{IntLists: &IntLists{Lists: values}}}
	case [][]int64:
		values := make([]*LongList, len(val))
		for i, v := range val {
			values[i] = &LongList{Features: v}
		}
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_LongLists{LongLists: &LongLists{Lists: values}}}
	case [][]string:
		values := make([]*StringList, len(val))
		for i, v := range val {
			values[i] = &StringList{Features: v}
		}
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_StringLists{StringLists: &StringLists{Lists: values}}}
	case [][]float32:
		values := make([]*FloatList, len(val))
		for i, v := range val {
			values[i] = &FloatList{Features: v}
		}
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_FloatLists{FloatLists: &FloatLists{Lists: values}}}
	case [][]float64:
		values := make([]*DoubleList, len(val))
		for i, v := range val {
			values[i] = &DoubleList{Features: v}
		}
		b.request.UserFeatures[k] = &PBFeature{Value: &PBFeature_DoubleLists{DoubleLists: &DoubleLists{Lists: values}}}

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
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_DoubleFeature{val}})
		case string:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringFeature{val}})
		case map[int64]string:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_LongStringMap{LongStringMap: &LongStringMap{MapField: val}}})
		case map[int64]int32:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_LongIntMap{LongIntMap: &LongIntMap{MapField: val}}})
		case map[int64]int:
			values := make(map[int64]int32, len(val))
			for k, v := range val {
				values[k] = int32(v)
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_LongIntMap{LongIntMap: &LongIntMap{MapField: values}}})
		case map[int64]int64:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_LongLongMap{LongLongMap: &LongLongMap{MapField: val}}})
		case map[int64]float32:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_LongFloatMap{LongFloatMap: &LongFloatMap{MapField: val}}})
		case map[int64]float64:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_LongDoubleMap{LongDoubleMap: &LongDoubleMap{MapField: val}}})
		case map[string]string:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringStringMap{StringStringMap: &StringStringMap{MapField: val}}})
		case map[string]int32:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringIntMap{StringIntMap: &StringIntMap{MapField: val}}})
		case map[string]int:
			values := make(map[string]int32, len(val))
			for k, v := range val {
				values[k] = int32(v)
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringIntMap{StringIntMap: &StringIntMap{MapField: values}}})
		case map[string]int64:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringLongMap{StringLongMap: &StringLongMap{MapField: val}}})
		case map[string]float32:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringFloatMap{StringFloatMap: &StringFloatMap{MapField: val}}})
		case map[string]float64:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringDoubleMap{StringDoubleMap: &StringDoubleMap{MapField: val}}})
		case map[int32]string:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntStringMap{IntStringMap: &IntStringMap{MapField: val}}})
		case map[int32]int32:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntIntMap{IntIntMap: &IntIntMap{MapField: val}}})
		case map[int32]int64:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntLongMap{IntLongMap: &IntLongMap{MapField: val}}})
		case map[int32]float32:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntFloatMap{IntFloatMap: &IntFloatMap{MapField: val}}})
		case map[int32]float64:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntDoubleMap{IntDoubleMap: &IntDoubleMap{MapField: val}}})
		case map[int]string:
			values := make(map[int32]string, len(val))
			for k, v := range val {
				values[int32(k)] = v
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntStringMap{IntStringMap: &IntStringMap{MapField: values}}})
		case map[int]int32:
			values := make(map[int32]int32, len(val))
			for k, v := range val {
				values[int32(k)] = v
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntIntMap{IntIntMap: &IntIntMap{MapField: values}}})
		case map[int]int:
			values := make(map[int32]int32, len(val))
			for k, v := range val {
				values[int32(k)] = int32(v)
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntIntMap{IntIntMap: &IntIntMap{MapField: values}}})
		case map[int]int64:
			values := make(map[int32]int64, len(val))
			for k, v := range val {
				values[int32(k)] = v
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntLongMap{IntLongMap: &IntLongMap{MapField: values}}})
		case map[int]float32:
			values := make(map[int32]float32, len(val))
			for k, v := range val {
				values[int32(k)] = v
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntFloatMap{IntFloatMap: &IntFloatMap{MapField: values}}})
		case map[int]float64:
			values := make(map[int32]float64, len(val))
			for k, v := range val {
				values[int32(k)] = v
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntDoubleMap{IntDoubleMap: &IntDoubleMap{MapField: values}}})
		case []int32:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntList{IntList: &IntList{Features: val}}})
		case []int:
			values := make([]int32, len(val))
			for i, v := range val {
				values[i] = int32(v)
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntList{IntList: &IntList{Features: values}}})
		case []int64:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_LongList{LongList: &LongList{Features: val}}})
		case []string:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringList{StringList: &StringList{Features: val}}})
		case []float32:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_FloatList{FloatList: &FloatList{Features: val}}})
		case []float64:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_DoubleList{DoubleList: &DoubleList{Features: val}}})
		case [][]int32:
			values := make([]*IntList, len(val))
			for i, v := range val {
				values[i] = &IntList{Features: v}
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntLists{IntLists: &IntLists{Lists: values}}})
		case [][]int64:
			values := make([]*LongList, len(val))
			for i, v := range val {
				values[i] = &LongList{Features: v}
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_LongLists{LongLists: &LongLists{Lists: values}}})
		case [][]string:
			values := make([]*StringList, len(val))
			for i, v := range val {
				values[i] = &StringList{Features: v}
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringLists{StringLists: &StringLists{Lists: values}}})
		case [][]float32:
			values := make([]*FloatList, len(val))
			for i, v := range val {
				values[i] = &FloatList{Features: v}
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_FloatLists{FloatLists: &FloatLists{Lists: values}}})
		case [][]float64:
			values := make([]*DoubleList, len(val))
			for i, v := range val {
				values[i] = &DoubleList{Features: v}
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_DoubleLists{DoubleLists: &DoubleLists{Lists: values}}})
		default:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringFeature{""}})
		}
	}

	b.request.ContextFeatures[key] = contextFeatures
}

func (b *EasyrecRequestBuilder) AddItemFeature(key string, features []interface{}) {
	contextFeatures := &ContextFeatures{
		Features: make([]*PBFeature, 0, len(features)),
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
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_DoubleFeature{val}})
		case string:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringFeature{val}})
		case map[int64]string:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_LongStringMap{LongStringMap: &LongStringMap{MapField: val}}})
		case map[int64]int32:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_LongIntMap{LongIntMap: &LongIntMap{MapField: val}}})
		case map[int64]int:
			values := make(map[int64]int32, len(val))
			for k, v := range val {
				values[k] = int32(v)
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_LongIntMap{LongIntMap: &LongIntMap{MapField: values}}})
		case map[int64]int64:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_LongLongMap{LongLongMap: &LongLongMap{MapField: val}}})
		case map[int64]float32:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_LongFloatMap{LongFloatMap: &LongFloatMap{MapField: val}}})
		case map[int64]float64:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_LongDoubleMap{LongDoubleMap: &LongDoubleMap{MapField: val}}})
		case map[string]string:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringStringMap{StringStringMap: &StringStringMap{MapField: val}}})
		case map[string]int32:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringIntMap{StringIntMap: &StringIntMap{MapField: val}}})
		case map[string]int:
			values := make(map[string]int32, len(val))
			for k, v := range val {
				values[k] = int32(v)
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringIntMap{StringIntMap: &StringIntMap{MapField: values}}})
		case map[string]int64:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringLongMap{StringLongMap: &StringLongMap{MapField: val}}})
		case map[string]float32:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringFloatMap{StringFloatMap: &StringFloatMap{MapField: val}}})
		case map[string]float64:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringDoubleMap{StringDoubleMap: &StringDoubleMap{MapField: val}}})
		case map[int32]string:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntStringMap{IntStringMap: &IntStringMap{MapField: val}}})
		case map[int32]int32:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntIntMap{IntIntMap: &IntIntMap{MapField: val}}})
		case map[int32]int64:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntLongMap{IntLongMap: &IntLongMap{MapField: val}}})
		case map[int32]float32:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntFloatMap{IntFloatMap: &IntFloatMap{MapField: val}}})
		case map[int32]float64:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntDoubleMap{IntDoubleMap: &IntDoubleMap{MapField: val}}})
		case map[int]string:
			values := make(map[int32]string, len(val))
			for k, v := range val {
				values[int32(k)] = v
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntStringMap{IntStringMap: &IntStringMap{MapField: values}}})
		case map[int]int32:
			values := make(map[int32]int32, len(val))
			for k, v := range val {
				values[int32(k)] = v
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntIntMap{IntIntMap: &IntIntMap{MapField: values}}})
		case map[int]int:
			values := make(map[int32]int32, len(val))
			for k, v := range val {
				values[int32(k)] = int32(v)
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntIntMap{IntIntMap: &IntIntMap{MapField: values}}})
		case map[int]int64:
			values := make(map[int32]int64, len(val))
			for k, v := range val {
				values[int32(k)] = v
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntLongMap{IntLongMap: &IntLongMap{MapField: values}}})
		case map[int]float32:
			values := make(map[int32]float32, len(val))
			for k, v := range val {
				values[int32(k)] = v
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntFloatMap{IntFloatMap: &IntFloatMap{MapField: values}}})
		case map[int]float64:
			values := make(map[int32]float64, len(val))
			for k, v := range val {
				values[int32(k)] = v
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntDoubleMap{IntDoubleMap: &IntDoubleMap{MapField: values}}})
		case []int32:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntList{IntList: &IntList{Features: val}}})
		case []int:
			values := make([]int32, len(val))
			for i, v := range val {
				values[i] = int32(v)
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntList{IntList: &IntList{Features: values}}})
		case []int64:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_LongList{LongList: &LongList{Features: val}}})
		case []string:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringList{StringList: &StringList{Features: val}}})
		case []float32:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_FloatList{FloatList: &FloatList{Features: val}}})
		case []float64:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_DoubleList{DoubleList: &DoubleList{Features: val}}})
		case [][]int32:
			values := make([]*IntList, len(val))
			for i, v := range val {
				values[i] = &IntList{Features: v}
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_IntLists{IntLists: &IntLists{Lists: values}}})
		case [][]int64:
			values := make([]*LongList, len(val))
			for i, v := range val {
				values[i] = &LongList{Features: v}
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_LongLists{LongLists: &LongLists{Lists: values}}})
		case [][]string:
			values := make([]*StringList, len(val))
			for i, v := range val {
				values[i] = &StringList{Features: v}
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringLists{StringLists: &StringLists{Lists: values}}})
		case [][]float32:
			values := make([]*FloatList, len(val))
			for i, v := range val {
				values[i] = &FloatList{Features: v}
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_FloatLists{FloatLists: &FloatLists{Lists: values}}})
		case [][]float64:
			values := make([]*DoubleList, len(val))
			for i, v := range val {
				values[i] = &DoubleList{Features: v}
			}
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_DoubleLists{DoubleLists: &DoubleLists{Lists: values}}})
		default:
			contextFeatures.Features = append(contextFeatures.Features, &PBFeature{Value: &PBFeature_StringFeature{""}})
		}
	}

	b.request.ItemFeatures[key] = contextFeatures
}
