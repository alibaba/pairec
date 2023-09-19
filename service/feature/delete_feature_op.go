package feature

import (
	"strings"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
)

type DeleteFeatureOp struct {
	featureOp
}

// UserTransOp of DeleteFeatureOp to Trans user feature
// it create new feature store in user properties
func (op DeleteFeatureOp) UserTransOp(featureName string, source string, remove bool, normalizer Normalizer, user *module.User, context *context.RecommendContext) {
	vals := strings.Split(source, ",")
	var features []string
	for _, val := range vals {
		comms := strings.Split(val, ":")
		if len(comms) >= 2 {
			features = append(features, comms[1])
		} else if len(comms) == 1 {
			features = append(features, comms[0])
		}
	}
	user.DeleteProperties(features)
}

// ItemTransOp of DeleteFeatureOp to Trans item or user feature
// it create new feature store in item properties
func (op DeleteFeatureOp) ItemTransOp(featureName string, source string, remove bool, normalizer Normalizer, user *module.User, item *module.Item, context *context.RecommendContext) {
	vals := strings.Split(source, ",")
	var features []string
	for _, val := range vals {
		comms := strings.Split(val, ":")
		if len(comms) >= 2 {
			features = append(features, comms[1])
		} else if len(comms) == 1 {
			features = append(features, comms[0])
		}
	}
	item.DeleteProperties(features)
}
