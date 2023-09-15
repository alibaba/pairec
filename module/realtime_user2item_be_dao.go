package module

import (
	"database/sql"
	"fmt"
	gosort "sort"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
	be "github.com/aliyun/aliyun-be-go-sdk"
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/datasource/beengine"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/utils"
)

type RealtimeUser2ItemBeDao struct {
	*RealtimeUser2ItemBaseDao
	hasPlayTimeField bool
	//itemCount                 int
	beClient                  *be.Client
	bizName                   string
	beRecallName              string
	weightEvaluableExpression *govaluate.EvaluableExpression
	weightMode                string
	beItemFeatureKeyName      string
	beTimestampFeatureKeyName string
	beEventFeatureKeyName     string
	bePlayTimeFeatureKeyName  string
}

func NewRealtimeUser2ItemBeDao(config recconf.RecallConfig) *RealtimeUser2ItemBeDao {
	dao := &RealtimeUser2ItemBeDao{
		bizName: config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.BizName,
		//itemCount:                 config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.ItemCount,
		hasPlayTimeField:          true,
		weightMode:                config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.WeightMode,
		beItemFeatureKeyName:      config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.BeItemFeatureKeyName,
		beTimestampFeatureKeyName: config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.BeTimestampFeatureKeyName,
		beEventFeatureKeyName:     config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.BeEventFeatureKeyName,
		bePlayTimeFeatureKeyName:  config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.BePlayTimeFeatureKeyName,
		RealtimeUser2ItemBaseDao:  NewRealtimeUser2ItemBaseDao(&config),
		beRecallName:              "sequence_feature",
	}
	if config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.BeRecallName != "" {
		dao.beRecallName = config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.BeRecallName
	}
	if config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.NoUsePlayTimeField {
		dao.hasPlayTimeField = false
	}

	client, err := beengine.GetBeClient(config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.BeName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}
	dao.beClient = client.BeClient

	expression, err := govaluate.NewEvaluableExpressionWithFunctions(config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.WeightExpression,
		govaluateFunctions)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}

	dao.weightEvaluableExpression = expression

	if dao.weightMode == "" {
		dao.weightMode = weight_mode_sum
	}

	return dao
}

func (d *RealtimeUser2ItemBeDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {

	itemTriggers := d.GetTriggers(user, context)
	if len(itemTriggers) == 0 {
		return
	}

	return
}

func (d *RealtimeUser2ItemBeDao) GetTriggerInfos(user *User, context *context.RecommendContext) (triggerInfos []*TriggerInfo) {
	itemTriggerMap := make(map[string]*TriggerInfo, d.limit)

	readRequest := be.NewReadRequest(d.bizName, d.limit)
	readRequest.IsRawRequest = true

	var sequence_feature_list []string
	for event := range d.eventWeightMap {
		sequence_feature_list = append(sequence_feature_list, fmt.Sprintf("%s_%s:1", user.Id, event))
	}
	params := make(map[string]string)
	params[fmt.Sprintf("%s_list", d.beRecallName)] = strings.Join(sequence_feature_list, ",")
	params[fmt.Sprintf("%s_return_count", d.beRecallName)] = fmt.Sprintf("%d", d.limit)

	readRequest.SetQueryParams(params)

	if context.Debug {
		uri := readRequest.BuildUri()
		log.Info(fmt.Sprintf("requestId=%s\tmodule=RealtimeUser2ItemBeDao\tbizName=%s\turl=%s", context.RecommendId, d.bizName, uri.RequestURI()))
	}

	readResponse, err := d.beClient.Read(*readRequest)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=RealtimeUser2ItemBeDao\terror=be error(%v)", context.RecommendId, err.Error()))
		return
	}

	matchItems := readResponse.Result.MatchItems
	if matchItems == nil || len(matchItems.FieldValues) == 0 {
		return
	}

	currentTime := time.Now()
	for _, values := range matchItems.FieldValues {
		trigger := new(TriggerInfo)
		var propertyFieldValues []sql.NullString
		properties := make(map[string]interface{}, 8)
		for i, value := range values {
			if matchItems.FieldNames[i] == d.beItemFeatureKeyName {
				trigger.ItemId = utils.ToString(value, "")
			} else if matchItems.FieldNames[i] == d.beEventFeatureKeyName {
				trigger.event = utils.ToString(value, "")
			} else if matchItems.FieldNames[i] == d.beTimestampFeatureKeyName {
				trigger.timestamp = utils.ToInt64(value, 0)
			} else if matchItems.FieldNames[i] == d.bePlayTimeFeatureKeyName {
				trigger.playTime = utils.ToFloat(value, 0)
			} else {
				propertyFieldValues = append(propertyFieldValues, sql.NullString{String: utils.ToString(value, ""), Valid: true})
				properties[matchItems.FieldNames[i]] = utils.ToFloat(value, 0)

			}
		}
		if trigger.ItemId != "" && trigger.event != "" {
			trigger.propertyFieldValues = propertyFieldValues
		}
		if t, exist := d.eventPlayTimeMap[trigger.event]; exist {
			if trigger.playTime <= t {
				continue
			}
		}

		weightScore := float64(1)
		if score, ok := d.eventWeightMap[trigger.event]; ok {
			weightScore = score
		}

		eventScore := float64(0)

		properties["currentTime"] = float64(currentTime.Unix())
		properties["eventTime"] = float64(trigger.timestamp)

		if result, err := d.weightEvaluableExpression.Evaluate(properties); err == nil {
			if value, ok := result.(float64); ok {
				eventScore = value
			}
		} else {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=RealtimeUser2ItemBeDao\terror=%v", context.RecommendId, err))
		}

		weight := weightScore * eventScore

		if info, exist := itemTriggerMap[trigger.ItemId]; exist {
			switch d.weightMode {
			case weight_mode_max:
				if weight > info.Weight {
					info.Weight = weight
				}
			default:
				info.Weight += weight
			}
		} else {
			trigger.Weight = weight
			itemTriggerMap[trigger.ItemId] = trigger
		}
	}

	for _, triggerInfo := range itemTriggerMap {
		triggerInfos = append(triggerInfos, triggerInfo)
	}
	gosort.Sort(gosort.Reverse(TriggerInfoSlice(triggerInfos)))

	triggerInfos = d.DiversityTriggers(triggerInfos)

	if len(triggerInfos) > d.triggerCount {
		triggerInfos = triggerInfos[:d.triggerCount]
	}

	return
}
func (d *RealtimeUser2ItemBeDao) GetTriggers(user *User, context *context.RecommendContext) (itemTriggers map[string]float64) {
	triggerInfos := d.GetTriggerInfos(user, context)
	itemTriggers = make(map[string]float64, len(triggerInfos))

	for _, trigger := range triggerInfos {
		itemTriggers[trigger.ItemId] = trigger.Weight
	}

	return
}
