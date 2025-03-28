package module

import (
	"database/sql"
	"fmt"
	gosort "sort"
	"strconv"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type RealtimeUser2ItemFeatureStoreDao struct {
	*RealtimeUser2ItemBaseDao
	hasPlayTimeField          bool
	itemCount                 int
	fsClient                  *fs.FSClient
	userTriggerTable          string
	itemTable                 string
	weightEvaluableExpression *govaluate.EvaluableExpression
	weightMode                string
	itemIdFieldName           string
	eventFieldName            string
	timestampFieldName        string
	playtimeFieldName         string
	events                    []any
	similarItemIdField        string
}

func NewRealtimeUser2ItemFeatureStoreDao(config recconf.RecallConfig) *RealtimeUser2ItemFeatureStoreDao {
	dao := &RealtimeUser2ItemFeatureStoreDao{
		itemCount:                config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.ItemCount,
		userTriggerTable:         config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.FeatureStoreViewName,
		hasPlayTimeField:         true,
		itemTable:                config.RealTimeUser2ItemDaoConf.Item2ItemFeatureViewName,
		weightMode:               config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.WeightMode,
		RealtimeUser2ItemBaseDao: NewRealtimeUser2ItemBaseDao(&config),
		itemIdFieldName:          "item_id",
		eventFieldName:           "event",
		playtimeFieldName:        "play_time",
		timestampFieldName:       "timestamp",
		similarItemIdField:       "similar_item_ids",
	}
	if config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.NoUsePlayTimeField {
		dao.hasPlayTimeField = false
	}
	fsclient, err := fs.GetFeatureStoreClient(config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.FeatureStoreName)
	if err != nil {
		panic(fmt.Sprintf("error=%v", err))
	}

	dao.fsClient = fsclient

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

	if config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.ItemIdFieldName != "" {
		dao.itemIdFieldName = config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.ItemIdFieldName
	}
	if config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.EventFieldName != "" {
		dao.eventFieldName = config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.EventFieldName
	}
	if config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.TimestampFieldName != "" {
		dao.timestampFieldName = config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.TimestampFieldName
	}
	if config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.PlayTimeFieldName != "" {
		dao.playtimeFieldName = config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.PlayTimeFieldName
	}
	if config.RealTimeUser2ItemDaoConf.SimilarItemIdField != "" {
		dao.similarItemIdField = config.RealTimeUser2ItemDaoConf.SimilarItemIdField
	}

	for k := range dao.eventWeightMap {
		dao.events = append(dao.events, k)
	}
	return dao
}

func (d *RealtimeUser2ItemFeatureStoreDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {

	itemTriggers := d.GetTriggers(user, context)
	if len(itemTriggers) == 0 {
		return
	}
	if d.itemTable == "" {
		for itemId, weight := range itemTriggers {
			item := NewItem(itemId)
			item.RetrieveId = d.recallName
			item.Score = weight
			ret = append(ret, item)
		}
		return
	}

	var itemIds []interface{}
	for id := range itemTriggers {
		itemIds = append(itemIds, id)
	}

	featureView := d.fsClient.GetProject().GetFeatureView(d.itemTable)
	if featureView == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=RealtimeUser2ItemFeatureStoreDao\trecallName=%s\terror=featureView not found, featureview:%s", context.RecommendId, d.recallName, d.itemTable))
		return
	}

	features, err := featureView.GetOnlineFeatures(itemIds, []string{d.similarItemIdField}, nil)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=RealtimeUser2ItemFeatureStoreDao\terror=%v", context.RecommendId, err))
		return
	}
	featureView.GetFeatureEntityName()
	featureEntity := d.fsClient.GetProject().GetFeatureEntity(featureView.GetFeatureEntityName())
	if featureEntity == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=RealtimeUser2ItemFeatureStoreDao\terror=featureEntity not found, featureEntity:%s", context.RecommendId, featureView.GetFeatureEntityName()))
		return
	}

	for _, featureMap := range features {

		triggerId := utils.ToString(featureMap[featureEntity.FeatureEntityJoinid], "")
		ids := utils.ToString(featureMap[d.similarItemIdField], "")
		preferScore := itemTriggers[triggerId]
		list := strings.Split(ids, ",")
		for _, str := range list {
			strs := strings.Split(str, ":")
			if strs[0] == "" || strs[0] == "null" {
				continue
			}
			if len(strs) == 2 {
				item := NewItem(strs[0])
				item.RetrieveId = d.recallName
				if tmpScore, err := strconv.ParseFloat(strings.TrimSpace(strs[1]), 64); err == nil {
					item.Score = tmpScore * preferScore
				} else {
					item.Score = preferScore
				}

				ret = append(ret, item)
			} else if len(strs) == 3 { // compatible format itemid1:recall1:score1
				item := NewItem(strs[0])
				item.RetrieveId = d.recallName
				if tmpScore, err := strconv.ParseFloat(strings.TrimSpace(strs[2]), 64); err == nil {
					item.Score = tmpScore * preferScore
				} else {
					item.Score = preferScore
				}

				ret = append(ret, item)
			}
		}

	}

	gosort.Sort(gosort.Reverse(ItemScoreSlice(ret)))
	ret = uniqItems(ret)

	if len(ret) > d.recallCount {
		ret = ret[:d.recallCount]
	}

	return
}

func (d *RealtimeUser2ItemFeatureStoreDao) GetTriggerInfos(user *User, context *context.RecommendContext) (triggerInfos []*TriggerInfo) {
	featureView := d.fsClient.GetProject().GetFeatureView(d.userTriggerTable)
	if featureView == nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=RealtimeUser2ItemFeatureStoreDao\trecallName=%s\terror=featureView not found, featureview:%s", context.RecommendId, d.recallName, d.userTriggerTable))
		return
	}
	itemTriggerMap := make(map[string]*TriggerInfo, 50)
	var selectFields []string
	if d.hasPlayTimeField {
		selectFields = []string{d.itemIdFieldName, d.eventFieldName, d.playtimeFieldName, d.timestampFieldName}
	} else {
		selectFields = []string{d.itemIdFieldName, d.eventFieldName, d.timestampFieldName}
	}
	if len(d.propertyFields) > 0 {
		selectFields = append(selectFields, d.propertyFields...)
	}
	features, err := featureView.GetBehaviorFeatures([]any{user.Id}, d.events, selectFields)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=RealtimeUser2ItemFeatureStoreDao\terror=featurestore error(%v)", context.RecommendId, err))
		return
	}

	currentTime := time.Now()
	for _, seqData := range features {
		trigger := new(TriggerInfo)
		trigger.ItemId = utils.ToString(seqData[d.itemIdFieldName], "")
		trigger.event = utils.ToString(seqData[d.eventFieldName], "")
		trigger.timestamp = utils.ToInt64(seqData[d.timestampFieldName], 0)
		if d.hasPlayTimeField {
			trigger.playTime = utils.ToFloat(seqData[d.playtimeFieldName], 0)
		}
		if t, exist := d.eventPlayTimeMap[trigger.event]; exist {
			if trigger.playTime <= t {
				continue
			}
		}
		for _, propertyField := range d.propertyFields {
			trigger.propertyFieldValues = append(trigger.propertyFieldValues, sql.NullString{String: utils.ToString(seqData[propertyField], ""), Valid: true})
		}
		weightScore := float64(1)
		if score, ok := d.eventWeightMap[trigger.event]; ok {
			weightScore = score
		}

		eventScore := float64(0)
		properties := map[string]interface{}{
			"currentTime": float64(currentTime.Unix()),
			"eventTime":   float64(trigger.timestamp),
		}
		if result, err := d.weightEvaluableExpression.Evaluate(properties); err == nil {
			if value, ok := result.(float64); ok {
				eventScore = value
			}
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
func (d *RealtimeUser2ItemFeatureStoreDao) GetTriggers(user *User, context *context.RecommendContext) (itemTriggers map[string]float64) {
	triggerInfos := d.GetTriggerInfos(user, context)
	itemTriggers = make(map[string]float64, len(triggerInfos))

	for _, trigger := range triggerInfos {
		itemTriggers[trigger.ItemId] = trigger.Weight
	}

	return
}
