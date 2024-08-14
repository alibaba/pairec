package feature

import (
	"fmt"
	"strings"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
)

type BatchRawFeatureOp struct {
	featureOp
}

// UserTransOp of BatchRawFeatureOp to Trans user feature
// it create new feature store in user properties
func (op BatchRawFeatureOp) UserTransOp(featureName string, source string, remove bool, normalizer Normalizer, user *module.User, context *context.RecommendContext) {
	names := strings.Split(featureName, ",")
	sources := strings.Split(source, ",")
	if len(names) != len(sources) {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=BatchRawFeatureOp\terror=featureName, featureSource length not equal", op.getRequestId(context)))
		return
	}

	for i, s := range sources {
		comms := strings.Split(s, ":")
		if len(comms) >= 2 {
			value := user.StringProperty(comms[1])
			user.AddProperty(names[i], value)
		}
	}
}

// ItemTransOp of BatchRawFeatureOp to Trans item or user feature
// it create new feature store in item properties
func (op BatchRawFeatureOp) ItemTransOp(featureName string, source string, remove bool, normalizer Normalizer, user *module.User, item *module.Item, context *context.RecommendContext) {
	names := strings.Split(featureName, ",")
	sources := strings.Split(source, ",")
	if len(names) != len(sources) {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=BatchRawFeatureOp\terror=featureName, featureSource length not equal", op.getRequestId(context)))
		return
	}

	for i, s := range sources {
		comms := strings.Split(s, ":")
		if len(comms) >= 2 {
			if comms[0] == SOURCE_USER {
				value := user.StringProperty(comms[1])
				item.AddProperty(names[i], value)
			} else {
				value := item.StringProperty(comms[1])
				item.AddProperty(names[i], value)
			}
		}
	}
}
