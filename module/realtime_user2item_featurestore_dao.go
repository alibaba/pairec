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

	itemTriggers, triggerIds := d.GetTriggersBySort(user, context)
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
	triggerIdItemMap := make(map[string][]*Item, len(itemTriggers))
	for id := range itemTriggers {
		if d.cache != nil {
			if cacheValue, ok := d.cache.GetIfPresent(id); ok {
				if items, ok := cacheValue.([]*Item); ok {
					newItems := d.cloneItems(items)
					triggerIdItemMap[id] = newItems
					ret = append(ret, newItems...)
					continue
				}
			}
		}
		itemIds = append(itemIds, id)
	}

	isSnakeMode := d.isSnakeMergeMode()

	if len(itemIds) > 0 {
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
					triggerIdItemMap[triggerId] = append(triggerIdItemMap[triggerId], item)
				} else if len(strs) == 3 { // compatible format itemid1:recall1:score1
					item := NewItem(strs[0])
					item.RetrieveId = d.recallName
					if tmpScore, err := strconv.ParseFloat(strings.TrimSpace(strs[2]), 64); err == nil {
						item.Score = tmpScore * preferScore
					} else {
						item.Score = preferScore
					}

					ret = append(ret, item)
					triggerIdItemMap[triggerId] = append(triggerIdItemMap[triggerId], item)
				}
			}
		}
		if d.cache != nil {
			for triggerId, items := range triggerIdItemMap {
				if _, exist := d.cache.GetIfPresent(triggerId); !exist {
					d.cache.Put(triggerId, d.cloneItems(items))
				}
			}
		}
		if context.Debug {
			log.Info(fmt.Sprintf("requestId=%s\tmodule=RealtimeUser2ItemFeatureStoreDao\ttriggerId size=%d\titemIds size=%d", context.RecommendId, len(triggerIds), len(itemIds)))
		}
	}
	// sort items
	if isSnakeMode || context.Debug {
		for _, triggerId := range triggerIds {
			items := triggerIdItemMap[triggerId]
			if len(items) > 0 {
				gosort.Sort(gosort.Reverse(ItemScoreSlice(items)))
				triggerIdItemMap[triggerId] = items
			}
		}
	}
	if context.Debug {
		for _, triggerId := range triggerIds {
			items := triggerIdItemMap[triggerId]
			if len(items) > 0 {
				str := d.debugItemsString(items)
				log.Info(fmt.Sprintf("requestId=%s\tmodule=RealtimeUser2ItemFeatureStoreDao\ttriggerId=%s\ttriggerScore=%f\titems=%s", context.RecommendId, triggerId, itemTriggers[triggerId], str))
			}
		}
	}

	if isSnakeMode {
		uniq := make(map[ItemId]bool, d.recallCount)
		resultItems := make([]*Item, 0, d.recallCount)
		emptyItemsCount := 0
		for len(resultItems) < d.recallCount {
			emptyItemsCount = 0
			for _, triggerId := range triggerIds {
				items := triggerIdItemMap[triggerId]
				if len(items) == 0 {
					emptyItemsCount++
					continue
				} else {
					item := items[0]
					if _, ok := uniq[item.Id]; !ok {
						uniq[item.Id] = true
						resultItems = append(resultItems, item)
					}
					triggerIdItemMap[triggerId] = items[1:]
				}
			}
			if emptyItemsCount == len(triggerIds) {
				break
			}
		}

		ret = resultItems
	} else {
		gosort.Sort(gosort.Reverse(ItemScoreSlice(ret)))
		ret = uniqItems(ret)
	}

	if len(ret) > d.recallCount {
		ret = ret[:d.recallCount]
	}

	if context.Debug {
		str := d.debugItemsString(ret)
		log.Info(fmt.Sprintf("requestId=%s\tmodule=RealtimeUser2ItemFeatureStoreDao\titems=%s", context.RecommendId, str))
	}

	return
}
func (d *RealtimeUser2ItemFeatureStoreDao) debugItemsString(itmes []*Item) string {
	ret := make([]string, 0, len(itmes))
	for _, item := range itmes {
		ret = append(ret, fmt.Sprintf("%s:%f", item.Id, item.Score))
	}
	return strings.Join(ret, ",")
}
func (d *RealtimeUser2ItemFeatureStoreDao) cloneItems(items []*Item) (ret []*Item) {
	ret = make([]*Item, len(items))
	for i, item := range items {
		newItem := NewItem(string(item.Id))
		newItem.RetrieveId = item.RetrieveId
		newItem.Score = item.Score
		ret[i] = newItem
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

func (d *RealtimeUser2ItemFeatureStoreDao) GetTriggersBySort(user *User, context *context.RecommendContext) (itemTriggers map[string]float64, triggerIds []string) {
	triggerInfos := d.GetTriggerInfos(user, context)
	itemTriggers = make(map[string]float64, len(triggerInfos))

	for _, trigger := range triggerInfos {
		itemTriggers[trigger.ItemId] = trigger.Weight
		triggerIds = append(triggerIds, trigger.ItemId)
	}

	return
}
