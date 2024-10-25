package feature_log

import (
	"fmt"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/datahub"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"strconv"
	"strings"
	"time"
)

func FeatureLog(user *module.User, items []*module.Item, context *context.RecommendContext) {
	scene := context.GetParameter("scene").(string)

	config, ok := context.Config.FeatureLogConfs[scene]

	if !ok {
		return
	}

	messages := getFeatureData(user, config.UserFeatures, items, config.ItemFeatures, context)

	if len(messages) == 0 {
		return
	}

	if config.OutputType == "datahub" {
		datahubApi, err := datahub.GetDatahub(config.DatahubName)
		if err != nil {
			log.Error(fmt.Sprintf("event=FeatureLog\terr=%v", err))
			return
		}
		go datahubApi.SendMessage(messages)
	}
}

func getFeatureData(user *module.User, userFeatures string, items []*module.Item, itemFeatures string, context *context.RecommendContext) []map[string]interface{} {
	messages := make([]map[string]interface{}, 0, len(items))
	if len(items) == 0 {
		return messages
	}

	for i, item := range items {

		logMap := make(map[string]interface{})
		logMap["request_id"] = context.RecommendId
		logMap["scene_id"] = context.GetParameter("scene")
		if context.ExperimentResult != nil {
			logMap["exp_id"] = context.ExperimentResult.GetExpId()
		}
		logMap["request_time"] = time.Now().Unix()
		logMap["user_features"] = getUserFeatures(user, userFeatures)
		logMap["user_id"] = string(user.Id)

		logMap["item_id"] = string(item.Id)
		logMap["position"] = strconv.Itoa(i + 1)
		logMap["item_features"] = getItemFeatures(item, itemFeatures)
		messages = append(messages, logMap)
	}
	return messages
}

func getUserFeatures(user *module.User, userFeatures string) (result map[string]interface{}) {
	result = make(map[string]interface{}, 8)

	if userFeatures == "" {
		return
	} else if userFeatures == "*" {
		result = user.MakeUserFeatures()
		return
	}

	userFields := strings.Split(userFeatures, ",")

	for _, field := range userFields {
		result[field] = user.GetProperty(field)
	}
	return
}

func getItemFeatures(item *module.Item, itemFeatures string) (result map[string]interface{}) {
	result = make(map[string]interface{}, 8)
	result["retrieve_id"] = item.RetrieveId
	result["score"] = item.Score
	result["algo_score"] = item.CloneAlgoScores()

	if itemFeatures == "" {
		return
	} else if itemFeatures == "*" {
		features := item.GetFeatures()

		for key, value := range features {
			result[key] = value
		}
		return
	}

	itemFields := strings.Split(itemFeatures, ",")
	for _, field := range itemFields {
		result[field] = item.GetProperty(field)
	}
	return
}
