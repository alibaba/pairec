package feature

import (
	"strings"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/utils/fasttime"
)

// CreateNewFeatureOp create new feature by Normalizer
type CreateNewFeatureOp struct {
}

// it create new feature store in user properties
func (op CreateNewFeatureOp) UserTransOp(featureName string, source string, remove bool, normalizer Normalizer, user *module.User, context *context.RecommendContext) {
	if normalizer != nil {
		if _, ok := normalizer.(*CreateConstValueNormalizer); ok {
			user.AddProperty(featureName, source)
		} else if _, ok := normalizer.(*ExpressionNormalizer); ok {
			params := user.MakeUserFeatures2()
			result := normalizer.Apply(params)
			if boolValue, ok := result.(bool); ok {
				if boolValue {
					user.AddProperty(featureName, 1)
				} else {
					user.AddProperty(featureName, 0)
				}
			} else {
				user.AddProperty(featureName, result)
			}
		} else if _, ok := normalizer.(*ExprNormalizer); ok {
			params := user.MakeUserFeatures2()
			result := normalizer.Apply(map[string]any{"user": params, "currentTime": fasttime.UnixTimestamp()})
			if boolValue, ok := result.(bool); ok {
				if boolValue {
					user.AddProperty(featureName, 1)
				} else {
					user.AddProperty(featureName, 0)
				}
			} else {
				user.AddProperty(featureName, result)
			}
		} else {
			user.AddProperty(featureName, normalizer.Apply(nil))
		}
	}
}

const (
	Context_User_Features_Key = "_user_features_"
)

func (op CreateNewFeatureOp) ItemTransOp(featureName string, source string, remove bool, normalizer Normalizer, user *module.User, item *module.Item, context *context.RecommendContext) {
	params := make(map[string]interface{})
	params["currentTime"] = fasttime.UnixTimestamp() // current time in seconds
	if source == "item:recall_name" {
		params["recall_name"] = item.RetrieveId
	} else if source == "" {
		if _, ok := normalizer.(*ExprNormalizer); ok {
			itemParams := item.GetCloneFeatures()
			if _, ok := itemParams["recall_name"]; !ok {
				itemParams["recall_name"] = item.RetrieveId
			}
			var userParams map[string]interface{}
			if user != nil {
				if userFeatures := context.GetContextParam(Context_User_Features_Key); userFeatures != nil {
					if userFeaturesMap, ok := userFeatures.(map[string]interface{}); ok {
						userParams = userFeaturesMap
					}
				}
				if len(userParams) == 0 {
					userParams = user.MakeUserFeatures2()
					context.AddContextParam(Context_User_Features_Key, userParams)
				}
			}

			params = map[string]interface{}{"item": itemParams, "user": userParams, "currentTime": fasttime.UnixTimestamp()}
		} else {
			params = item.GetCloneFeatures()
			params["currentTime"] = fasttime.UnixTimestamp() // current time in seconds
			if _, ok := params["recall_name"]; !ok {
				params["recall_name"] = item.RetrieveId
			}
		}
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
