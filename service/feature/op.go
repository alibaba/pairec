package feature

import (
	"fmt"
	"strings"

	"github.com/alibaba/pairec/v2/context"
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
				value := item.StringProperty(comms[1])
				featureValue += "_" + value
				if remove {
					item.DeleteProperty(comms[1])
				}

			}
		}
	}

	item.AddProperty(featureName, featureValue)
}
