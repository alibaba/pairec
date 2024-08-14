package feature

import (
	"strings"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
)

// CreateNewFeatureOp create new feature by Normalizer
type CreateNewFeatureOp struct {
}

// it create new feature store in user properties
func (op CreateNewFeatureOp) UserTransOp(featureName string, source string, remove bool, normalizer Normalizer, user *module.User, context *context.RecommendContext) {
	if normalizer != nil {
		if _, ok := normalizer.(*CreateConstValueNormalizer); ok {
			user.AddProperty(featureName, source)
		} else {
			user.AddProperty(featureName, normalizer.Apply(nil))
		}
	}
}

func (op CreateNewFeatureOp) ItemTransOp(featureName string, source string, remove bool, normalizer Normalizer, user *module.User, item *module.Item, context *context.RecommendContext) {
	params := make(map[string]interface{})
	if source == "item:recall_name" {
		params["recall_name"] = item.RetrieveId
	} else if source == "" {
		params = item.GetFeatures()
	} else {
		comms := strings.Split(source, ":")
		if len(comms) >= 2 {
			if comms[0] == SOURCE_USER {
				value := user.GetProperty(comms[1])
				params[comms[1]] = value
			} else {
				value := item.GetProperty(comms[1])
				params[comms[1]] = value
			}
		}

	}

	result := normalizer.Apply(params)
	if boolValue, ok := result.(bool); ok {
		if boolValue {
			item.AddProperty(featureName, 1)
		} else {
			item.AddProperty(featureName, 0)
		}
	} else {
		item.AddProperty(featureName, result)
	}
}
