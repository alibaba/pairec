package slsfeature

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/sls"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/service/hook"
)

type FeatureLogSLSFunc func(*sls.SlsClient, *module.User, []*module.Item, *context.RecommendContext)

func FeatureLogToSLS(slsName string, f FeatureLogSLSFunc) {
	slsclient, err := sls.GetSlsClient(slsName)
	if err != nil {
		panic(fmt.Sprintf("get datahub error, :%v", err))
	}
	hook.AddRecommendCleanHook(func(slsclient *sls.SlsClient, f FeatureLogSLSFunc) hook.RecommendCleanHookFunc {

		return func(context *context.RecommendContext, params ...interface{}) {
			user := params[0].(*module.User)
			items := params[1].([]*module.Item)
			f(slsclient, user, items, context)
		}
	}(slsclient, f))
}

func itemFeature(item *module.Item) (result map[string]interface{}) {
	result = item.GetFeatures()
	result["retrieveId"] = item.RetrieveId
	result["score"] = item.Score
	return
}

func DefaultFeatureLogSLSFunc(slsclient *sls.SlsClient, user *module.User, items []*module.Item, context *context.RecommendContext) {
	if len(items) == 0 {
		return
	}

	log := make(map[string]string)
	log["RequestId"] = context.RecommendId
	log["SceneId"] = context.GetParameter("scene").(string)
	if context.ExperimentResult != nil {
		log["ExpId"] = context.ExperimentResult.GetExpId()
	}
	log["RequestTime"] = strconv.FormatInt(time.Now().Unix(), 10)
	j, _ := json.Marshal(user.MakeUserFeatures())
	log["UserFeatures"] = string(j)
	log["UserId"] = string(user.Id)
	for i, item := range items {
		log["ItemId"] = string(item.Id)
		log["Position"] = strconv.Itoa(i + 1)
		j, _ := json.Marshal(itemFeature(item))
		log["itemFeatures"] = string(j)
		slsclient.SendLog(log)
	}

}
