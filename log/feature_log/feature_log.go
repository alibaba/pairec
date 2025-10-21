package feature_log

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/datahub"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func FeatureLog(user *module.User, items []*module.Item, context *context.RecommendContext) {
	scene := context.GetParameter("scene").(string)

	config, ok := context.Config.FeatureLogConfs[scene]

	if !ok {
		return
	}

	var featureLogFlag bool
	if config.Rate == 100 || config.Rate == 0 {
		featureLogFlag = true
	} else {
		if rand.Intn(100) < config.Rate {
			featureLogFlag = true
		}
	}
	if !featureLogFlag {
		return
	}

	messages := getFeatureData(user, config.UserFeatures, items, config.ItemFeatures, context)

	if len(messages) == 0 {
		return
	}

	if config.OutputType == "datahub" {
		datahubApi, err := datahub.GetDatahub(config.DatahubName)
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tevent=FeatureLog\terr=%v", context.RecommendId, err))
			return
		}
		go datahubApi.SendMessage(messages)
	}
}

func getFeatureData(user *module.User, userFields string, items []*module.Item, itemFields string, context *context.RecommendContext) []map[string]interface{} {
	messages := make([]map[string]interface{}, 0, len(items))
	if len(items) == 0 {
		return messages
	}

	var userData string
	userFeatures := getUserFeatures(user, userFields)
	if len(userFeatures) > 0 {
		data, err := json.Marshal(userFeatures)
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tevent=FeatureLog\terr=%v", context.RecommendId, err))
		} else {
			userData = string(data)
		}
	}
	requestTime := time.Now().Unix()
	for i, item := range items {

		logMap := make(map[string]interface{}, 8)
		logMap["request_id"] = context.RecommendId
		logMap["scene_id"] = context.GetParameter("scene")
		if context.ExperimentResult != nil {
			logMap["exp_id"] = context.ExperimentResult.GetExpId()
		}
		logMap["request_time"] = requestTime
		if userData != "" {
			logMap["user_features"] = userData
		}
		/*
			userFeatures := getUserFeatures(user, userFields)
			if len(userFeatures) > 0 {
				data, err := json.Marshal(userFeatures)
				if err != nil {
					log.Error(fmt.Sprintf("requestId=%s\tevent=FeatureLog\terr=%v", context.RecommendId, err))
				} else {
					logMap["user_features"] = string(data)
				}
			}
		*/

		logMap["user_id"] = string(user.Id)
		logMap["item_id"] = string(item.Id)
		logMap["position"] = strconv.Itoa(i + 1)
		itemFeatures := getItemFeatures(item, itemFields)
		if len(itemFeatures) > 0 {
			data, err := json.Marshal(itemFeatures)
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tevent=FeatureLog\terr=%v", context.RecommendId, err))
			} else {
				logMap["item_features"] = string(data)
			}
		}
		messages = append(messages, logMap)
	}
	return messages
}

func getUserFeatures(user *module.User, userFields string) (result map[string]interface{}) {
	result = make(map[string]interface{}, 8)

	if userFields == "" {
		return
	} else if userFields == "*" {
		result = user.MakeUserFeatures2()
		return
	}

	userFieldsArray := strings.Split(userFields, ",")

	for _, field := range userFieldsArray {
		result[field] = user.GetProperty(field)
	}
	return
}

func getItemFeatures(item *module.Item, itemFields string) (result map[string]interface{}) {
	result = make(map[string]interface{}, 8)
	result["retrieve_id"] = item.RetrieveId
	result["score"] = item.Score
	result["algo_score"] = item.CloneAlgoScores()

	if itemFields == "" {
		return
	} else if itemFields == "*" {
		features := item.GetFeatures()

		for key, value := range features {
			result[key] = value
		}
		return
	}

	itemFieldsArray := strings.Split(itemFields, ",")
	for _, field := range itemFieldsArray {
		result[field] = item.GetProperty(field)
	}
	return
}
