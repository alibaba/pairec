package datasource

import (
	"fmt"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/service/hook"
)

type FeatureLogKafkaFunc func(*KafkaProducer, *module.User, []*module.Item, *context.RecommendContext)

func FeatureLogToKafka(kafkaName string, f FeatureLogKafkaFunc) {
	producer, err := GetKafkaProducer(kafkaName)
	if err != nil {
		panic(fmt.Sprintf("get kafka producer error, :%v", err))
	}
	hook.AddRecommendCleanHook(func(producer *KafkaProducer, f FeatureLogKafkaFunc) hook.RecommendCleanHookFunc {

		return func(context *context.RecommendContext, params ...interface{}) {
			user := params[0].(*module.User)
			items := params[1].([]*module.Item)
			f(producer, user, items, context)
		}
	}(producer, f))
}
