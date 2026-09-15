package feature

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
)

type FeatureOp interface {
	UserTransOp(featureName string, source string, remove bool, normalizer Normalizer, user *module.User, context *context.RecommendContext)
	ItemTransOp(featureName string, source string, remove bool, normalizer Normalizer, user *module.User, item *module.Item, context *context.RecommendContext)
}

func NewFeatureOp(t string) FeatureOp {
	if t == "raw_feature" {
		return RawFeatureOp{}
	} else if t == "compose_feature" {
		return ComposeFeatureOp{}
	} else if t == "delete_feature" {
		return DeleteFeatureOp{}
	} else if t == "batch_raw_feature" {
		return BatchRawFeatureOp{}
	} else if t == "new_feature" {
		return CreateNewFeatureOp{}
	} else if t == "context_feature" {
		return ContextFeatureOp{}
	} else if t == "expand_json_feature" {
		return ExpandJsonFeatureOp{}
	}

	panic(fmt.Sprintf("not find feature type:%s", t))
}

type featureOp struct {
}

func (op featureOp) getRequestId(context *context.RecommendContext) string {
	if nil == context {
		return ""
	}

	return context.RecommendId
}

type RawFeatureOp struct {
	featureOp
}

// UserTransOp of RawFeatureOp to Trans user feature
// it create new feature store in user properties
func (op RawFeatureOp) UserTransOp(featureName string, source string, remove bool, normalizer Normalizer, user *module.User, context *context.RecommendContext) {
	comms := strings.Split(source, ":")
	if len(comms) >= 2 {
		value := user.StringProperty(comms[1])
		user.AddProperty(featureName, value)
		if remove {
			user.DeleteProperty(comms[1])
		}
	}
}

// ItemTransOp of RawFeatureOp to Trans item or user feature
// it create new feature store in item properties
func (op RawFeatureOp) ItemTransOp(featureName string, source string, remove bool, normalizer Normalizer, user *module.User, item *module.Item, context *context.RecommendContext) {
	comms := strings.Split(source, ":")
	if len(comms) >= 2 {
		if comms[0] == SOURCE_USER {
			value := user.StringProperty(comms[1])
			item.AddProperty(featureName, value)
			if remove {
				user.DeleteProperty(comms[1])
			}
		} else {
			var newValue interface{}
			value := item.StringProperty(comms[1])
			newValue = value
			if normalizer != nil {
				newValue = normalizer.Apply(value)
			}
			item.AddProperty(featureName, newValue)
			if remove {
				item.DeleteProperty(comms[1])
			}

		}
	}
}

type ComposeFeatureOp struct {
	featureOp
}

func (op ComposeFeatureOp) UserTransOp(featureName string, source string, remove bool, normalizer Normalizer, user *module.User, context *context.RecommendContext) {
	vals := strings.Split(source, ",")
	var featureValue string
	for _, val := range vals {
		comms := strings.Split(val, ":")
		if len(comms) >= 2 {
			value := user.StringProperty(comms[1])
			featureValue += "_" + value
			if remove {
				user.DeleteProperty(comms[1])
			}
		}
	}

	user.AddProperty(featureName, featureValue)
}

func (op ComposeFeatureOp) ItemTransOp(featureName string, source string, remove bool, normalizer Normalizer, user *module.User, item *module.Item, context *context.RecommendContext) {
	vals := strings.Split(source, ",")
	featureValue := featureName
	for _, val := range vals {
		comms := strings.Split(val, ":")
		if len(comms) >= 2 {
			if comms[0] == SOURCE_USER {
				value := user.StringProperty(comms[1])
				// item.AddProperty(featureName, value)
				featureValue += "_" + value
			} else {
				var value string
				if comms[1] == "id" {
					value = string(item.Id)
				} else {
					value = item.StringProperty(comms[1])
				}
				featureValue += "_" + value
				if remove {
					item.DeleteProperty(comms[1])
				}

			}
		}
	}

	item.AddProperty(featureName, featureValue)
}

// ExpandJsonFeatureOp expands a property holding a JSON object string into
// separate properties, one per key of that object. It is used to restore a
// feature snapshot that is stored as a single JSON string column, such as the
// user_features column written by feature log.
//
// FeatureName is not used by this op, the expanded feature names come from the
// keys of the JSON object. Values are restored by their JSON type, so an integer
// stays an integer and a list or object becomes the concrete slice or map type
// the easyrec request builder has a case for, see convertJsonValue.
//
// RemoveFeatureSource is ignored when FeatureStore is item and FeatureSource
// points at a user property, because that single property is the source shared
// by every item, same as ComposeFeatureOp does.
type ExpandJsonFeatureOp struct {
	featureOp
}

func (op ExpandJsonFeatureOp) UserTransOp(featureName string, source string, remove bool, normalizer Normalizer, user *module.User, context *context.RecommendContext) {
	comms := strings.Split(source, ":")
	if len(comms) < 2 {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=ExpandJsonFeatureOp\terror=featureSource error(%s)", op.getRequestId(context), source))
		return
	}

	value := user.StringProperty(comms[1])
	if value == "" {
		return
	}

	properties, err := unmarshalJsonProperties(value)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=ExpandJsonFeatureOp\tsource=%s\terror=%v", op.getRequestId(context), comms[1], err))
		return
	}

	// remove the source before expanding, so a key of the same name inside the
	// JSON object is kept instead of being deleted right after being added
	if remove {
		user.DeleteProperty(comms[1])
	}
	user.AddProperties(properties)
}

func (op ExpandJsonFeatureOp) ItemTransOp(featureName string, source string, remove bool, normalizer Normalizer, user *module.User, item *module.Item, context *context.RecommendContext) {
	comms := strings.Split(source, ":")
	if len(comms) < 2 {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=ExpandJsonFeatureOp\terror=featureSource error(%s)", op.getRequestId(context), source))
		return
	}

	var value string
	if comms[0] == SOURCE_USER {
		value = user.StringProperty(comms[1])
	} else {
		value = item.StringProperty(comms[1])
	}
	if value == "" {
		return
	}

	properties, err := unmarshalJsonProperties(value)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=ExpandJsonFeatureOp\tsource=%s\terror=%v", op.getRequestId(context), comms[1], err))
		return
	}

	// only a source on the item itself is removed. FeatureTran calls ItemTransOp
	// once per item, so removing a user property while handling the first item
	// would leave the remaining items with an empty source to expand.
	if remove && comms[0] != SOURCE_USER {
		item.DeleteProperty(comms[1])
	}
	item.AddProperties(properties)
}

// unmarshalJsonProperties decodes a JSON object string into properties, keeping
// the numbers as json.Number. The default decoding turns every number into a
// float64, which loses precision above 2^53 and also changes the feature type
// that is finally sent to the processor, so a snapshot would not restore the
// features the model was scored with.
func unmarshalJsonProperties(value string) (map[string]interface{}, error) {
	decoder := json.NewDecoder(strings.NewReader(value))
	decoder.UseNumber()

	properties := make(map[string]interface{})
	if err := decoder.Decode(&properties); err != nil {
		return nil, err
	}
	for k, v := range properties {
		properties[k] = convertJsonValue(v)
	}
	return properties, nil
}

// convertJsonValue restores the Go type of a value decoded with json.Number.
//
// The restoration is driven by the json type of the value, because a json round
// trip keeps the string and number distinction: a []string is written back as a
// list of quoted elements while a []int64 is written back as a list of bare
// numbers. Restoring by that distinction keeps the feature type close to the one
// the model was scored with, which a blanket conversion to string would not.
func convertJsonValue(value interface{}) interface{} {
	switch v := value.(type) {
	case json.Number:
		return convertJsonNumber(v)
	case []interface{}:
		return convertJsonList(v)
	case map[string]interface{}:
		return convertJsonMap(v)
	default:
		return value
	}
}

// convertJsonNumber restores the numeric type of a json.Number, the same way
// web.FeaturesMap does for the features carried in a request body.
func convertJsonNumber(number json.Number) interface{} {
	if i64, err := number.Int64(); err == nil {
		return int(i64)
	}
	if f64, err := number.Float64(); err == nil {
		return f64
	}
	return number.String()
}

// convertJsonList restores a list to the concrete slice type matching the json
// type of its elements, so the easyrec request builder has a case for it. A list
// the builder has no type for only gets its elements restored one by one.
func convertJsonList(list []interface{}) interface{} {
	if len(list) == 0 {
		// an empty list gives no hint about its element type
		return list
	}

	allString, allNumber, allList := true, true, true
	integral, numeric := true, true
	for _, elem := range list {
		switch v := elem.(type) {
		case string:
			allNumber, allList = false, false
		case json.Number:
			allString, allList = false, false
			if _, err := v.Int64(); err != nil {
				integral = false
				if _, err := v.Float64(); err != nil {
					numeric = false
				}
			}
		case []interface{}:
			allString, allNumber = false, false
		default:
			allString, allNumber, allList = false, false, false
		}
	}

	switch {
	case allString:
		values := make([]string, len(list))
		for i, elem := range list {
			values[i] = elem.(string)
		}
		return values
	case allNumber && integral:
		values := make([]int, len(list))
		for i, elem := range list {
			i64, _ := elem.(json.Number).Int64()
			values[i] = int(i64)
		}
		return values
	case allNumber && numeric:
		values := make([]float64, len(list))
		for i, elem := range list {
			f64, _ := elem.(json.Number).Float64()
			values[i] = f64
		}
		return values
	case allList:
		return convertJsonNestedList(list)
	}

	for i, elem := range list {
		list[i] = convertJsonValue(elem)
	}
	return list
}

// convertJsonMap restores a json object to the concrete map type matching the
// json type of its values, so the easyrec request builder has a case for it.
//
// A json object always has string keys, so the key type of the original feature
// is not recoverable: a map[int64]string and a map[string]string look exactly the
// same once written. The keys are restored as strings, which turns a LongStringMap
// into a StringStringMap. That is accepted, because the builder has no case for
// map[string]interface{} at all, so the alternative is dropping the feature.
//
// An integral value gives a map[string]int64 and not a map[string]int, because
// the builder casts a map[string]int down to int32 without a range check.
func convertJsonMap(object map[string]interface{}) interface{} {
	if len(object) == 0 {
		// an empty object gives no hint about its value type
		return object
	}

	allString, allNumber := true, true
	integral, numeric := true, true
	for _, elem := range object {
		switch v := elem.(type) {
		case string:
			allNumber = false
		case json.Number:
			allString = false
			if _, err := v.Int64(); err != nil {
				integral = false
				if _, err := v.Float64(); err != nil {
					numeric = false
				}
			}
		default:
			allString, allNumber = false, false
		}
	}

	switch {
	case allString:
		values := make(map[string]string, len(object))
		for key, elem := range object {
			values[key] = elem.(string)
		}
		return values
	case allNumber && integral:
		values := make(map[string]int64, len(object))
		for key, elem := range object {
			values[key], _ = elem.(json.Number).Int64()
		}
		return values
	case allNumber && numeric:
		values := make(map[string]float64, len(object))
		for key, elem := range object {
			values[key], _ = elem.(json.Number).Float64()
		}
		return values
	}

	for key, elem := range object {
		object[key] = convertJsonValue(elem)
	}
	return object
}

// convertJsonNestedList restores a list of lists, the builder carries those as
// [][]string, [][]int64 or [][]float64.
func convertJsonNestedList(list []interface{}) interface{} {
	allString, allNumber := true, true
	integral, numeric := true, true
	for _, elem := range list {
		for _, inner := range elem.([]interface{}) {
			switch v := inner.(type) {
			case string:
				allNumber = false
			case json.Number:
				allString = false
				if _, err := v.Int64(); err != nil {
					integral = false
					if _, err := v.Float64(); err != nil {
						numeric = false
					}
				}
			default:
				allString, allNumber = false, false
			}
		}
	}

	switch {
	case allString:
		values := make([][]string, len(list))
		for i, elem := range list {
			inner := elem.([]interface{})
			values[i] = make([]string, len(inner))
			for j, v := range inner {
				values[i][j] = v.(string)
			}
		}
		return values
	case allNumber && integral:
		values := make([][]int64, len(list))
		for i, elem := range list {
			inner := elem.([]interface{})
			values[i] = make([]int64, len(inner))
			for j, v := range inner {
				values[i][j], _ = v.(json.Number).Int64()
			}
		}
		return values
	case allNumber && numeric:
		values := make([][]float64, len(list))
		for i, elem := range list {
			inner := elem.([]interface{})
			values[i] = make([]float64, len(inner))
			for j, v := range inner {
				values[i][j], _ = v.(json.Number).Float64()
			}
		}
		return values
	}

	for i, elem := range list {
		list[i] = convertJsonValue(elem)
	}
	return list
}

// ContextFeatureOp add context feature to user
type ContextFeatureOp struct {
	//featureOp
}

func (op ContextFeatureOp) UserTransOp(featureName string, source string, remove bool, normalizer Normalizer, user *module.User, context *context.RecommendContext) {

	contextFeatures := context.GetParameter("features")
	if contextFeatures != nil {
		if ctxFeatures, ok := contextFeatures.(map[string]any); ok {
			user.AddProperties(ctxFeatures)
		}
	}
}

func (op ContextFeatureOp) ItemTransOp(featureName string, source string, remove bool, normalizer Normalizer, user *module.User, item *module.Item, context *context.RecommendContext) {
}
