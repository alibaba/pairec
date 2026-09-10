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
// keys of the JSON object. Note that the JSON round trip degrades every number
// to float64, so integer valued features are restored as float64.
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

	properties := make(map[string]interface{})
	if err := json.Unmarshal([]byte(value), &properties); err != nil {
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

	properties := make(map[string]interface{})
	if err := json.Unmarshal([]byte(value), &properties); err != nil {
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
