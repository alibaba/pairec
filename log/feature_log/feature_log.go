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
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
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

	var messages []map[string]interface{}
	if config.OutputType == "featurestore" {
		// for featurestore output, only one log is written per request,
		// item infos are merged into one record when LogItems is enabled
		messages = getFeatureStoreLogData(user, config.UserFeatures, items, config.ItemFeatures, config.LogItems, context)
	} else {
		messages = getFeatureData(user, config.UserFeatures, items, config.ItemFeatures, config.SplitUserItemLogs, context)
	}

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
	} else if config.OutputType == "featurestore" {
		fsClient, err := fs.GetFeatureStoreClient(config.FeatureStoreName)
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tevent=FeatureLog\terr=%v", context.RecommendId, err))
			return
		}
		featureView := fsClient.GetProject().GetFeatureView(config.FeatureStoreViewName)
		if featureView == nil {
			log.Error(fmt.Sprintf("requestId=%s\tevent=FeatureLog\terr=feature view not found, name:%s", context.RecommendId, config.FeatureStoreViewName))
			return
		}
		// write feature logs to FeatureDB directly via fs sdk direct write interface
		go func() {
			if err := featureView.WriteFeatures(messages, domain.WithDirect()); err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tevent=FeatureLog\terr=%v", context.RecommendId, err))
			}
		}()
	}
}

func getFeatureData(user *module.User, userFields string, items []*module.Item, itemFields string, splitLog bool, context *context.RecommendContext) []map[string]interface{} {
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

	if splitLog {
		userLogMap := map[string]interface{}{
			"request_id":   context.RecommendId,
			"scene_id":     context.GetParameter("scene"),
			"user_id":      string(user.Id),
			"request_time": requestTime,
		}

		if context.ExperimentResult != nil {
			userLogMap["exp_id"] = context.ExperimentResult.GetExpId()
		}

		if userData != "" {
			userLogMap["user_features"] = userData
		}

		messages = append(messages, userLogMap)
	}

	for i, item := range items {

		logMap := make(map[string]interface{}, 8)
		logMap["request_id"] = context.RecommendId
		logMap["scene_id"] = context.GetParameter("scene")
		if context.ExperimentResult != nil {
			logMap["exp_id"] = context.ExperimentResult.GetExpId()
		}
		logMap["request_time"] = requestTime
		if userData != "" && !splitLog {
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

// getFeatureStoreLogData builds a single feature log record for featurestore output.
// Unlike getFeatureData which logs one record per item, this returns only one log
// per request. When logItems is enabled, all item infos are merged into one
// record in the "items" field, otherwise item infos are not logged.
func getFeatureStoreLogData(user *module.User, userFields string, items []*module.Item, itemFields string, logItems bool, context *context.RecommendContext) []map[string]interface{} {
	requestTime := time.Now().Unix()

	logMap := map[string]interface{}{
		"request_id":   context.RecommendId,
		"scene_id":     context.GetParameter("scene"),
		"user_id":      string(user.Id),
		"request_time": requestTime,
	}

	if context.ExperimentResult != nil {
		logMap["exp_id"] = context.ExperimentResult.GetExpId()
	}

	userFeatures := getUserFeatures(user, userFields)
	if len(userFeatures) > 0 {
		data, err := json.Marshal(userFeatures)
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tevent=FeatureLog\terr=%v", context.RecommendId, err))
		} else {
			logMap["user_features"] = string(data)
		}
	}

	if logItems {
		itemRecords := make([]map[string]interface{}, 0, len(items))
		for i, item := range items {
			itemMap := getItemFeatures(item, itemFields)
			itemMap["item_id"] = string(item.Id)
			itemMap["position"] = strconv.Itoa(i + 1)
			itemRecords = append(itemRecords, itemMap)
		}
		data, err := json.Marshal(itemRecords)
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tevent=FeatureLog\terr=%v", context.RecommendId, err))
		} else {
			logMap["items"] = string(data)
		}
	}

	return []map[string]interface{}{logMap}
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
