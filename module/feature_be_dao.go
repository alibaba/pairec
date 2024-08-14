package module

import (
	"fmt"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	be "github.com/aliyun/aliyun-be-go-sdk"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/beengine"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type RTCntField struct {
	fieldNames  []string
	fieldDelims []string
}

type userRtCnt struct {
	key       string
	cnt       int32
	timestamp int64
}

type FeatureBeDao struct {
	*FeatureBaseDao
	hasPlayTimeField          bool
	beClient                  *be.Client
	bizName                   string
	beRecallName              string
	userFeatureKeyName        string
	itemFeatureKeyName        string
	timestampFeatureKeyName   string
	eventFeatureKeyName       string
	playTimeFeatureKeyName    string
	tsFeatureKeyName          string
	beItemFeatureKeyName      string
	beTimestampFeatureKeyName string
	beEventFeatureKeyName     string
	bePlayTimeFeatureKeyName  string
	// such as "is_home", indicates whether it is
	// the current recommendation scene
	beIsHomeField string
	beRTCntFields []RTCntField
	beRTTable     string
	rtCntMaxKey   int
	rtCntWins     []int
	// rtCnt feature window delay
	rtCntWinDelay int64
	// user__style_id_rt_clk_30m
	// user__%s_rt_%s_%s
	outRTCntFeaPattern string
	// user__style_id_rt_home_clk_30m
	// user__%s_rt_home_%s_%s
	outHomeRTCntFeaPattern string
	outRTCntFieldAlias     []string
	outRTCntWinNames       []string
	outEventName           string
}

func NewFeatureBeDao(config recconf.FeatureDaoConfig) *FeatureBeDao {
	var rtCntWins []int
	if config.FeatureType == Feature_Type_RT_Cnt {
		rtCntWinToks := strings.Split(config.RTCntWins, ",")
		for _, winTok := range rtCntWinToks {
			tWin := utils.ToInt(winTok, -1)
			if tWin <= 0 {
				log.Error(fmt.Sprintf("invalid rtCntWins: %s", winTok))
				return nil
			}
			rtCntWins = append(rtCntWins, tWin)
		}
	}
	var beRTCntFields []RTCntField
	var outRTCntFieldAlias []string
	if config.FeatureType == Feature_Type_RT_Cnt {
		if len(config.BeRTCntFieldInfo) > 0 {
			for fieldId, fieldInfo := range config.BeRTCntFieldInfo {
				tmpToks := fieldInfo.FieldNames
				for tid := range tmpToks {
					tmpToks[tid] = strings.TrimSpace(tmpToks[tid])
				}
				var delimToks []string
				if len(fieldInfo.Delims) > 0 {
					delimToks = fieldInfo.Delims
				} else {
					for i := 0; i < len(tmpToks); i++ {
						delimToks = append(delimToks, "")
					}
				}
				if len(delimToks) != len(tmpToks) {
					log.Error(fmt.Sprintf("len(fieldToks) != len(delimToks): %d vs %d, fieldId=%d, fieldTok=%s, delimStr=%s",
						len(tmpToks), len(delimToks), fieldId,
						strings.Join(fieldInfo.FieldNames, ":"),
						strings.Join(fieldInfo.Delims, ":")))
					return nil
				}
				beRTCntFields = append(beRTCntFields,
					RTCntField{tmpToks, delimToks})
				tmpAlias := strings.TrimSpace(fieldInfo.Alias)
				if len(tmpAlias) == 0 {
					tmpAlias = tmpToks[0]
				}
				outRTCntFieldAlias = append(outRTCntFieldAlias, tmpAlias)
			}
		} else { // [deprecicated], to be compatible with existing config
			rtCntFields := strings.Split(config.BeRTCntFields, ",")
			for _, fieldTok := range rtCntFields {
				tmpToks := strings.Split(fieldTok, ":")
				for tid := range tmpToks {
					tmpToks[tid] = strings.TrimSpace(tmpToks[tid])
				}
				delimToks := make([]string, len(tmpToks))
				beRTCntFields = append(beRTCntFields, RTCntField{tmpToks, delimToks})
			}
			tAliasArr := strings.Split(config.OutRTCntFieldAlias, ",")
			for _, tmpAlias := range tAliasArr {
				outRTCntFieldAlias = append(outRTCntFieldAlias,
					strings.TrimSpace(tmpAlias))
			}
			if len(beRTCntFields) != len(outRTCntFieldAlias) {
				log.Error(fmt.Sprintf("len(beRTCntFields) != len(outRTCntFieldAlias): %d vs %d", len(beRTCntFields), len(outRTCntFieldAlias)))
				return nil
			}
		}
	}

	dao := &FeatureBeDao{
		FeatureBaseDao:            NewFeatureBaseDao(&config),
		bizName:                   config.BizName,
		beRecallName:              "sequence_feature",
		userFeatureKeyName:        config.UserFeatureKeyName,
		itemFeatureKeyName:        config.ItemFeatureKeyName,
		timestampFeatureKeyName:   config.TimestampFeatureKeyName,
		eventFeatureKeyName:       config.EventFeatureKeyName,
		playTimeFeatureKeyName:    config.PlayTimeFeatureKeyName,
		tsFeatureKeyName:          config.TsFeatureKeyName,
		hasPlayTimeField:          true,
		beItemFeatureKeyName:      config.BeItemFeatureKeyName,
		beTimestampFeatureKeyName: config.BeTimestampFeatureKeyName,
		beEventFeatureKeyName:     config.BeEventFeatureKeyName,
		bePlayTimeFeatureKeyName:  config.BePlayTimeFeatureKeyName,

		beIsHomeField: strings.TrimSpace(config.BeIsHomeField),
		beRTCntFields: beRTCntFields,
		beRTTable:     strings.TrimSpace(config.BeRTTable),

		rtCntWins:     rtCntWins,
		rtCntMaxKey:   utils.ToInt(config.RTCntMaxKey, 100),
		rtCntWinDelay: utils.ToInt64(config.RTCntWinDelay, 5),

		outRTCntFeaPattern:     strings.TrimSpace(config.OutRTCntFeaPattern),
		outHomeRTCntFeaPattern: strings.TrimSpace(config.OutHomeRTCntFeaPattern),
		outRTCntWinNames:       strings.Split(config.OutRTCntWinNames, ","),
		outRTCntFieldAlias:     outRTCntFieldAlias,
		outEventName:           strings.TrimSpace(config.OutEventName),
	}
	if config.BeRecallName != "" {
		dao.beRecallName = config.BeRecallName
	}

	if dao.featureType == Feature_Type_RT_Cnt {
		for tid := range dao.outRTCntWinNames {
			dao.outRTCntWinNames[tid] = strings.TrimSpace(
				dao.outRTCntWinNames[tid])
		}

		if len(dao.rtCntWins) != len(dao.outRTCntWinNames) {
			log.Error(fmt.Sprintf("len(rtCntWins)[%d] != len(outRTCntWinNames)[%d]",
				len(dao.rtCntWins), len(dao.outRTCntWinNames)))
			return nil
		}
		if len(dao.beRTCntFields) != len(dao.outRTCntFieldAlias) {
			log.Error(fmt.Sprintf("len(beRTCntFields)[%d] != len(outRTCntFieldAlias)[%d]",
				len(dao.beRTCntFields), len(dao.outRTCntFieldAlias)))
			return nil
		}
		log.Info(fmt.Sprintf("RTCntConfig:beIsHomeField=%s", dao.beIsHomeField))
		log.Info(fmt.Sprintf("RTCntConfig:beRTTable=%s", dao.beRTTable))
		log.Info(fmt.Sprintf("RTCntConfig:rtCntWinDelay=%d", dao.rtCntWinDelay))
		log.Info(fmt.Sprintf("RTCntConfig:rtCntMaxKey=%d", dao.rtCntMaxKey))
	}

	client, err := beengine.GetBeClient(config.BeName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}
	dao.beClient = client.BeClient
	if config.NoUsePlayTimeField {
		dao.hasPlayTimeField = false
	}
	return dao
}

func (d *FeatureBeDao) FeatureFetch(user *User, items []*Item, context *context.RecommendContext) {
	if d.featureStore == Feature_Store_User && d.featureType == Feature_Type_Sequence {
		d.userSequenceFeatureFetch(user, context)
	} else if d.featureStore == Feature_Store_User && d.featureType == Feature_Type_RT_Cnt {
		d.userRTCntFeatureFetch(user, context)
	} else if d.featureStore == Feature_Store_User {
		d.userFeatureFetch(user, context) // empty
	} else {
		d.itemsFeatureFetch(items, context) // empty
	}
}

func (d *FeatureBeDao) userFeatureFetch(user *User, context *context.RecommendContext) {
}

func debugMapKeys[T any](tmpMap map[string]T, sep string) string {
	var keys []string
	for key := range tmpMap {
		keys = append(keys, key)
	}
	return strings.Join(keys, sep)
}

func debugMap[T any](tmpMap map[string]T, sep string, kv_sep string) string {
	var kvs []string
	for k, v := range tmpMap {
		kvs = append(kvs, fmt.Sprintf("%s%s%s", k, kv_sep, utils.ToString(v, "")))
	}
	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i] < kvs[j]
	})
	return strings.Join(kvs, sep)
}

func debugArr[T any](tmpArr []T) string {
	var valStr []string
	for _, v := range tmpArr {
		valStr = append(valStr, utils.ToString(v, ""))
	}
	return strings.Join(valStr, ",")
}

func buildRTCntOutput(tmpMap map[string]userRtCnt, rtCntMaxKey int) string {
	var eventCntKVs []string
	if len(tmpMap) > rtCntMaxKey {
		// sort by timestamp
		var eventCnts []userRtCnt
		for _, val := range tmpMap {
			eventCnts = append(eventCnts, val)
		}
		sort.Slice(eventCnts, func(x, y int) bool {
			return eventCnts[x].timestamp > eventCnts[y].timestamp
		})
		for k := 0; k < rtCntMaxKey; k++ {
			eventCntKVs = append(eventCntKVs,
				fmt.Sprintf("%s:%d", eventCnts[k].key, eventCnts[k].cnt))
		}
	} else {
		for _, val := range tmpMap {
			eventCntKVs = append(eventCntKVs,
				fmt.Sprintf("%s:%d", val.key, val.cnt))
		}
	}
	return strings.Join(eventCntKVs, "")
}

func (d *FeatureBeDao) userRTCntFeatureFetch(user *User, context *context.RecommendContext) {
	defer func() {
		if err := recover(); err != nil {
			stack := string(debug.Stack())
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureBeDao\terror=%v\tstack=%v", context.RecommendId, err, strings.ReplaceAll(stack, "\n", "\t")))
			return
		}
	}()

	comms := strings.Split(d.featureKey, ":")
	if len(comms) < 2 {
		log.Error(fmt.Sprintf("requestId=%s\tuid=%s\terror=featureKey error(%s)", context.RecommendId, user.Id, d.featureKey))
		return
	}

	// uid
	key := user.StringProperty(comms[1])
	if key == "" {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureBeDao\terror=property not found(%s)", context.RecommendId, comms[1]))
		return
	}

	currTime := time.Now().Unix()

	readRequest := be.NewReadRequest(d.bizName, d.sequenceLength)
	readRequest.IsRawRequest = true

	params := make(map[string]string)
	params[fmt.Sprintf("%s_list", d.beRecallName)] = fmt.Sprintf("%s_%s:1", key, d.sequenceEvent)
	params[fmt.Sprintf("%s_return_count", d.beRecallName)] = fmt.Sprintf("%d", d.sequenceLength)
	params[fmt.Sprintf("%s_table", d.beRecallName)] = d.beRTTable

	readRequest.SetQueryParams(params)

	if context.Debug {
		uri := readRequest.BuildUri()
		log.Info(fmt.Sprintf("module=FeatureBeDao\tfunc=userRTCntFeatureFetch\trequestId=%s\tbizName=%s\turl=%s", context.RecommendId, d.bizName, uri.RequestURI()))
	}

	readResponse, err := d.beClient.Read(*readRequest)
	if err != nil {
		log.Error(fmt.Sprintf("module=FeatureBeDao\tfunc=userRTCntFeatureFetch\trequestId=%s\terror=be error(%v)", context.RecommendId, err.Error()))
		return
	}

	matchItems := readResponse.Result.MatchItems
	if matchItems == nil || len(matchItems.FieldValues) == 0 {
		if context.Debug {
			log.Warning(fmt.Sprintf("requestId=%s\tmodule=FeatureBeDao\tfunc=userRTCntFeatureFetch\tkey=%s\tevent=%s\twarning=NoMatchedItems",
				context.RecommendId, key, d.sequenceEvent))
		}
		return
	}

	var cntMaps []map[string]userRtCnt
	var homeCntMaps []map[string]userRtCnt
	for i := 0; i < len(d.beRTCntFields); i++ {
		for j := 0; j < len(d.rtCntWins); j++ {
			cntMaps = append(cntMaps, make(map[string]userRtCnt))
			homeCntMaps = append(homeCntMaps, make(map[string]userRtCnt))
		}
	}

	if context.Debug {
		log.Info(fmt.Sprintf("requestId=%sfunc=userRTCntFeatureFetch\tkey=%s\tevent=%s\tcurrTime=%d\tlen(matchItems.FieldValues)=%drtCntWins=%s",
			context.RecommendId, key, d.sequenceEvent, currTime, len(matchItems.FieldValues),
			debugArr[int](d.rtCntWins)))
	}
	for _, values := range matchItems.FieldValues {
		valMap := make(map[string]string)
		var itemId string
		ts := int64(-1)
		if context.Debug {
			log.Info(fmt.Sprintf("requestId=%s\tfunc=userRTCntFeatureFetch\tkey=%s\tevent=%s\tmatchItems.FieldValues: len(values)=%d",
				context.RecommendId, key, d.sequenceEvent, len(values)))
		}
		for i, value := range values {
			valMap[matchItems.FieldNames[i]] = utils.ToString(value, "")
			if matchItems.FieldNames[i] == d.beItemFeatureKeyName {
				itemId = utils.ToString(value, "")
			} else if matchItems.FieldNames[i] == d.beTimestampFeatureKeyName {
				ts = utils.ToInt64(value, -1)
			}
		}

		if context.Debug {
			log.Info(fmt.Sprintf("requestId=%s\tfunc=userRTCntFeatureFetch\tkey=%s\tevent=%s\tvalMap=%s",
				context.RecommendId, key, d.sequenceEvent, debugMap[string](valMap, ";", ":")))
		}

		if len(itemId) == 0 {
			log.Error(fmt.Sprintf("requestId=%s\tfunc=userRTCntFeatureFetch\tkey=%s\tevent=%s\tinvalid(itemId): %s",
				context.RecommendId, key, d.sequenceEvent, itemId))
			continue
		}

		if ts <= 0 {
			log.Error(fmt.Sprintf("requestId=%s\tfunc=userRTCntFeatureFetch\tkey=%s\tevent=%s\tinvalid(ts): %d",
				context.RecommendId, key, d.sequenceEvent, ts))
			continue
		}

		for i := 0; i < len(d.beRTCntFields); i++ {
			tmpCntField := d.beRTCntFields[i]
			var vals []string
			// for fields concatenated by multiple sub fields
			// such as rcc from root_category_id and child_category_id
			if len(tmpCntField.fieldNames) > 1 {
				for tId, fName := range tmpCntField.fieldNames {
					tVal, exist := valMap[fName]
					if !exist {
						log.Error(fmt.Sprintf("field key(%s) does not exist, available keys: %s",
							fName, debugMapKeys[string](valMap, ",")))
						continue
					} else {
						var currToks []string
						tmpDelim := tmpCntField.fieldDelims[tId]
						if len(tmpDelim) > 0 && len(tVal) > 0 {
							currToks = strings.Split(tVal, tmpDelim)
						} else {
							currToks = append(currToks, tVal)
						}
						if tId != 0 { // append more fields
							var tmpVals []string
							for _, prev := range vals {
								for _, curr := range currToks {
									tmpVals = append(tmpVals, prev+"_"+curr)
								}
							}
							vals = tmpVals
						} else { // the first one
							vals = currToks
						}
					}
				}
			} else { // for single field
				tVal, exist := valMap[tmpCntField.fieldNames[0]]
				if !exist {
					log.Error(fmt.Sprintf("field key(%s) does not exist, available keys: %s",
						tmpCntField.fieldNames[0], debugMapKeys[string](valMap, ",")))
					continue
				} else {
					tmpDelim := tmpCntField.fieldDelims[0]
					if len(tmpDelim) > 0 {
						vals = strings.Split(tVal, tmpDelim)
					} else {
						vals = append(vals, tVal)
					}
				}
			}

			isHome := utils.ToInt(valMap[d.beIsHomeField], 0)
			mapId := i * len(d.rtCntWins)
			for j := 0; j < len(d.rtCntWins); j++ {
				if ts >= (currTime-d.rtCntWinDelay) || ts < (currTime-d.rtCntWinDelay-int64(d.rtCntWins[j])) {
					continue
				}

				for _, val := range vals {
					if len(val) == 0 {
						// ignore empty value
						continue
					}
					tmpCnt, exist := cntMaps[mapId+j][val]
					if !exist {
						tmpCnt = userRtCnt{val, 1, ts}
					} else {
						tmpCnt.cnt += 1
						if ts > tmpCnt.timestamp {
							tmpCnt.timestamp = ts
						}
					}
					cntMaps[mapId+j][val] = tmpCnt

					if isHome > 0 {
						tmpCnt, exist := homeCntMaps[mapId+j][val]
						if !exist {
							tmpCnt = userRtCnt{val, 1, ts}
						} else {
							tmpCnt.cnt += 1
							if ts > tmpCnt.timestamp {
								tmpCnt.timestamp = ts
							}
						}
						homeCntMaps[mapId+j][val] = tmpCnt
					}
				}
			}
		}
	}

	properties := make(map[string]interface{})
	for i := 0; i < len(d.beRTCntFields); i++ {
		mapId := i * len(d.rtCntWins)
		for j := 0; j < len(d.rtCntWins); j++ {
			// build kv string
			tmpMap := cntMaps[mapId+j]
			if len(tmpMap) > 0 {
				tmpRTFeaStr := buildRTCntOutput(tmpMap, d.rtCntMaxKey)
				// user__style_id_rt_clk_30m
				feaOutName := fmt.Sprintf(d.outRTCntFeaPattern,
					d.outRTCntFieldAlias[i],
					d.outEventName, d.outRTCntWinNames[j])
				//user.AddProperty(feaOutName, tmpRTFeaStr)
				properties[feaOutName] = tmpRTFeaStr
				if context.Debug {
					log.Info(fmt.Sprintf("requestId=%sfunc=userRTCntFeatureFetch\tkey=%s\tevent=%s\tadd_user_fea:%s=%s",
						context.RecommendId, key, d.sequenceEvent, feaOutName, tmpRTFeaStr))
				}
			}

			// build home kv string
			tmpHomeMap := homeCntMaps[mapId+j]
			if len(tmpHomeMap) > 0 {
				tmpRTHomeFeaStr := buildRTCntOutput(tmpHomeMap, d.rtCntMaxKey)
				// user__style_id_rt_home_clk_30m
				homeFeaOutName := fmt.Sprintf(d.outHomeRTCntFeaPattern,
					d.outRTCntFieldAlias[i], d.outEventName, d.outRTCntWinNames[j])
				//user.AddProperty(homeFeaOutName, tmpRTHomeFeaStr)
				properties[homeFeaOutName] = tmpRTHomeFeaStr
				if context.Debug {
					log.Info(fmt.Sprintf("requestId=%sfunc=userRTCntFeatureFetch\tkey=%s\tevent=%s\tadd_user_home_fea:%s=%s",
						context.RecommendId, key, d.sequenceEvent, homeFeaOutName, tmpRTHomeFeaStr))
				}
			}
		}
	}

	if d.cacheFeaturesName != "" {
		user.AddCacheFeatures(d.cacheFeaturesName, properties)
	} else {
		user.AddProperties(properties)
	}

}

func (d *FeatureBeDao) userSequenceFeatureFetch(user *User, context *context.RecommendContext) {
	defer func() {
		if err := recover(); err != nil {
			stack := string(debug.Stack())
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureBeDao\terror=%v\tstack=%v", context.RecommendId, err, strings.ReplaceAll(stack, "\n", "\t")))
			return
		}
	}()

	comms := strings.Split(d.featureKey, ":")
	if len(comms) < 2 {
		log.Error(fmt.Sprintf("requestId=%s\tuid=%s\terror=featureKey error(%s)", context.RecommendId, user.Id, d.featureKey))
		return
	}

	key := user.StringProperty(comms[1])
	if key == "" {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureBeDao\terror=property not found(%s)", context.RecommendId, comms[1]))
		return
	}

	currTime := time.Now().Unix()
	var (
		item_feature_key_name      = "item_id"
		event_feature_key_name     = "event"
		play_time_feature_key_name = "play_time"
		timestamp_feature_key_name = "timestamp"
		ts_feature_key_name        = "ts"
		onlineSequences            []*sequenceInfo
	)

	if d.itemFeatureKeyName != "" {
		item_feature_key_name = d.itemFeatureKeyName
	}
	if d.eventFeatureKeyName != "" {
		event_feature_key_name = d.eventFeatureKeyName
	}
	if d.playTimeFeatureKeyName != "" {
		play_time_feature_key_name = d.playTimeFeatureKeyName
	}
	if d.timestampFeatureKeyName != "" {
		timestamp_feature_key_name = d.timestampFeatureKeyName
	}
	if d.tsFeatureKeyName != "" {
		ts_feature_key_name = d.tsFeatureKeyName
	}

	readRequest := be.NewReadRequest(d.bizName, d.sequenceLength)
	readRequest.IsRawRequest = true

	params := make(map[string]string)
	params[fmt.Sprintf("%s_list", d.beRecallName)] = fmt.Sprintf("%s_%s:1", key, d.sequenceEvent)
	params[fmt.Sprintf("%s_return_count", d.beRecallName)] = fmt.Sprintf("%d", d.sequenceLength)

	readRequest.SetQueryParams(params)

	if context.Debug {
		uri := readRequest.BuildUri()
		log.Info(fmt.Sprintf("requestId=%s\tevent=userSequenceFeatureFetch\tbizName=%s\turl=%s", context.RecommendId, d.bizName, uri.RequestURI()))
	}

	readResponse, err := d.beClient.Read(*readRequest)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureBeDao\terror=be error(%v)", context.RecommendId, err))
		return
	}

	matchItems := readResponse.Result.MatchItems
	if matchItems == nil || len(matchItems.FieldValues) == 0 {
		return
	}

	for _, values := range matchItems.FieldValues {
		seq := new(sequenceInfo)
		for i, value := range values {
			if matchItems.FieldNames[i] == d.beItemFeatureKeyName {
				seq.itemId = utils.ToString(value, "")
			} else if matchItems.FieldNames[i] == d.beEventFeatureKeyName {
				seq.event = utils.ToString(value, "")
			} else if matchItems.FieldNames[i] == d.beTimestampFeatureKeyName {
				seq.timestamp = utils.ToInt64(value, 0)
			} else if matchItems.FieldNames[i] == d.bePlayTimeFeatureKeyName {
				seq.playTime = utils.ToFloat(value, 0)
			}
		}
		if seq.itemId != "" && seq.event != "" {
			onlineSequences = append(onlineSequences, seq)
		}
	}

	// seqeunce feature correspond to easyrec processor
	sequencesValueMap := make(map[string][]string)
	sequenceMap := make(map[string]bool, len(onlineSequences))
	for _, seq := range onlineSequences {
		key := fmt.Sprintf("%s#%s", seq.itemId, seq.event)
		if _, exist := sequenceMap[key]; !exist {
			sequenceMap[key] = true
			sequencesValueMap[item_feature_key_name] = append(sequencesValueMap[item_feature_key_name], seq.itemId)
			sequencesValueMap[timestamp_feature_key_name] = append(sequencesValueMap[timestamp_feature_key_name], fmt.Sprintf("%d", seq.timestamp))
			sequencesValueMap[event_feature_key_name] = append(sequencesValueMap[event_feature_key_name], seq.event)
			if d.hasPlayTimeField {
				sequencesValueMap[play_time_feature_key_name] = append(sequencesValueMap[play_time_feature_key_name], fmt.Sprintf("%.2f", seq.playTime))
			}
			sequencesValueMap[ts_feature_key_name] = append(sequencesValueMap[ts_feature_key_name], fmt.Sprintf("%d", currTime-seq.timestamp))
			for index, field := range seq.dimensionFields {
				if field.Valid {
					sequencesValueMap[d.sequenceDimFields[index]] = append(sequencesValueMap[d.sequenceDimFields[index]], field.String)
				}
			}
		}
	}

	delim := d.sequenceDelim
	if delim == "" {
		delim = ";"
	}
	properties := make(map[string]interface{})
	for key, value := range sequencesValueMap {
		curSequenceSubName := (d.sequenceName + "__" + key)
		properties[curSequenceSubName] = strings.Join(value, delim)
	}
	properties[d.sequenceName] = strings.Join(sequencesValueMap[item_feature_key_name], delim)

	if d.cacheFeaturesName != "" {
		user.AddCacheFeatures(d.cacheFeaturesName, properties)
	} else {
		user.AddProperties(properties)
	}

}

func (d *FeatureBeDao) itemsFeatureFetch(items []*Item, context *context.RecommendContext) {
}
