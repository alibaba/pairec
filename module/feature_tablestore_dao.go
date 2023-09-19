package module

import (
	"database/sql"
	"fmt"
	"math"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/tablestoredb"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type FeatureTablestoreDao struct {
	*FeatureBaseDao
	tablestore              *tablestoredb.TableStore
	table                   string
	userFeatureKeyName      string
	itemFeatureKeyName      string
	userSelectFields        string
	itemSelectFields        string
	timestampFeatureKeyName string
	eventFeatureKeyName     string
	playTimeFeatureKeyName  string
	tsFeatureKeyName        string
	sequenceOfflineTable    string
}

func NewFeatureTablestoreDao(config recconf.FeatureDaoConfig) *FeatureTablestoreDao {
	dao := &FeatureTablestoreDao{
		FeatureBaseDao:          NewFeatureBaseDao(&config),
		table:                   config.TableStoreTableName,
		userFeatureKeyName:      config.UserFeatureKeyName,
		itemFeatureKeyName:      config.ItemFeatureKeyName,
		userSelectFields:        config.UserSelectFields,
		itemSelectFields:        config.ItemSelectFields,
		timestampFeatureKeyName: config.TimestampFeatureKeyName,
		eventFeatureKeyName:     config.EventFeatureKeyName,
		playTimeFeatureKeyName:  config.PlayTimeFeatureKeyName,
		tsFeatureKeyName:        config.TsFeatureKeyName,
		sequenceOfflineTable:    config.SequenceOfflineTableName,
	}
	tablestore, err := tablestoredb.GetTableStore(config.TableStoreName)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return nil
	}
	dao.tablestore = tablestore
	return dao
}

func (d *FeatureTablestoreDao) FeatureFetch(user *User, items []*Item, context *context.RecommendContext) {
	if d.featureStore == Feature_Store_User && d.featureType == Feature_Type_Sequence {
		d.userSequenceFeatureFetch(user, context)
	} else if d.featureStore == Feature_Store_User {
		d.userFeatureFetch(user, context)
	} else {
		d.itemsFeatureFetch(items, context)
	}
}

func (d *FeatureTablestoreDao) userFeatureFetch(user *User, context *context.RecommendContext) {
	comms := strings.Split(d.featureKey, ":")
	if len(comms) < 2 {
		log.Error(fmt.Sprintf("requestId=%s\tuid=%s\terror=featureKey error(%s)", context.RecommendId, user.Id, d.featureKey))
		return
	}

	key := user.StringProperty(comms[1])
	if key == "" {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureTablestoreDao\terror=property not found(%s)", context.RecommendId, comms[1]))
		return
	}

	getRowRequest := new(tablestore.GetRowRequest)
	criteria := new(tablestore.SingleRowQueryCriteria)
	putPk := new(tablestore.PrimaryKey)
	putPk.AddPrimaryKeyColumn(d.userFeatureKeyName, key)

	criteria.PrimaryKey = putPk
	if d.userSelectFields != "" {
		criteria.ColumnsToGet = strings.Split(d.userSelectFields, ",")
	}
	getRowRequest.SingleRowQueryCriteria = criteria
	getRowRequest.SingleRowQueryCriteria.TableName = d.table
	getRowRequest.SingleRowQueryCriteria.MaxVersion = 1
	getResp, err := d.tablestore.Client.GetRow(getRowRequest)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureTablestoreDao\terror=%v", context.RecommendId, err))
		return
	}

	properties := make(map[string]any)
	for _, column := range getResp.Columns {
		name := column.ColumnName
		switch val := column.Value.(type) {
		case string:
			properties[name] = val
		case int:
			properties[name] = val
		case float64:
			properties[name] = val
		default:
			properties[name] = val
		}
	}

	if d.cacheFeaturesName != "" {
		user.AddCacheFeatures(d.cacheFeaturesName, properties)
	} else {
		user.AddProperties(properties)
	}
}

// userSequenceFeatureFetch get sequence feature bind to user
func (d *FeatureTablestoreDao) userSequenceFeatureFetch(user *User, context *context.RecommendContext) {
	defer func() {
		if err := recover(); err != nil {
			stack := string(debug.Stack())
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureTablestoreDao\terror=%v\tstack=%v", context.RecommendId, err, strings.ReplaceAll(stack, "\n", "\t")))
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
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureTablestoreDao\terror=property not found(%s)", context.RecommendId, comms[1]))
		return
	}

	currTime := time.Now().Unix()

	item_feature_key_name := "item_id"
	if d.itemFeatureKeyName != "" {
		item_feature_key_name = d.itemFeatureKeyName
	}
	event_feature_key_name := "event"
	if d.eventFeatureKeyName != "" {
		event_feature_key_name = d.eventFeatureKeyName
	}
	play_time_feature_key_name := "play_time"
	if d.playTimeFeatureKeyName != "" {
		play_time_feature_key_name = d.playTimeFeatureKeyName
	}
	timestamp_feature_key_name := "timestamp"
	if d.timestampFeatureKeyName != "" {
		timestamp_feature_key_name = d.timestampFeatureKeyName
	}
	ts_feature_key_name := "ts"
	if d.tsFeatureKeyName != "" {
		ts_feature_key_name = d.tsFeatureKeyName
	}

	sequence_event_selections := strings.Split(d.sequenceEvent, ",")

	selectFields := []string{item_feature_key_name, play_time_feature_key_name, timestamp_feature_key_name}
	if len(d.sequenceDimFields) > 0 {
		selectFields = append(selectFields, d.sequenceDimFields...)
	}

	fetchDataFunc := func(table string) (sequences []*sequenceInfo) {

		batchGetReq := &tablestore.BatchGetRowRequest{}
		mqCriteria := &tablestore.MultiRowQueryCriteria{}
		for _, event := range sequence_event_selections {
			pkToGet := new(tablestore.PrimaryKey)
			pkToGet.AddPrimaryKeyColumn(d.userFeatureKeyName, key)
			pkToGet.AddPrimaryKeyColumn(event_feature_key_name, event)
			mqCriteria.AddRow(pkToGet)

		}
		mqCriteria.MaxVersion = d.sequenceLength
		mqCriteria.TableName = table
		mqCriteria.ColumnsToGet = selectFields

		timeRange := new(tablestore.TimeRange)
		timeRange.End = currTime * 1000
		timeRange.Start = (currTime - 86400*5) * 1000
		mqCriteria.TimeRange = timeRange

		batchGetReq.MultiRowQueryCriteria = append(batchGetReq.MultiRowQueryCriteria, mqCriteria)

		batchGetResponse, err := d.tablestore.Client.BatchGetRow(batchGetReq)
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureTablestoreDao\terror=%v", context.RecommendId, err))
			return
		}

		for _, row := range batchGetResponse.TableToRowsResult[table] {
			if row.IsSucceed {
				if row.PrimaryKey.PrimaryKeys == nil {
					continue
				}
				// get all versions
				versions := make([]int64, 0, 1)
				versionSeqmap := make(map[int64]*sequenceInfo, 0)
				for _, column := range row.Columns {
					if column.ColumnName == item_feature_key_name {
						versions = append(versions, column.Timestamp)
						seq := new(sequenceInfo)
						seq.event = utils.ToString(row.PrimaryKey.PrimaryKeys[1].Value, "")
						versionSeqmap[column.Timestamp] = seq

					}
				}
				for _, column := range row.Columns {
					seq := versionSeqmap[column.Timestamp]
					switch column.ColumnName {
					case item_feature_key_name:
						seq.itemId = utils.ToString(column.Value, "")
					case play_time_feature_key_name:
						seq.playTime = utils.ToFloat(column.Value, 0)
					case timestamp_feature_key_name:
						seq.timestamp = utils.ToInt64(column.Value, 0)
					default:
						sqlValue := sql.NullString{String: utils.ToString(column.Value, ""), Valid: true}
						seq.dimensionFields = append(seq.dimensionFields, sqlValue)

					}
				}
				for _, version := range versions {
					seq := versionSeqmap[version]
					if seq.event == "" || seq.itemId == "" {
						continue
					}
					if t, exist := d.sequencePlayTimeMap[seq.event]; exist {
						if seq.playTime <= t {
							continue
						}
					}

					sequences = append(sequences, seq)
				}
			}
		}

		return
	}
	var wg sync.WaitGroup
	var onlineSequences []*sequenceInfo
	var offlineSequences []*sequenceInfo

	wg.Add(1)
	go func() {
		defer wg.Done()
		onlineSequences = fetchDataFunc(d.table)
	}()
	if d.sequenceOfflineTable != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			offlineSequences = fetchDataFunc(d.sequenceOfflineTable)
		}()
	}
	wg.Wait()

	if len(offlineSequences) > 0 {
		index := 0
		for index < len(onlineSequences) {
			if onlineSequences[index].timestamp < offlineSequences[0].timestamp {
				break
			}
			index++
		}

		onlineSequences = onlineSequences[:index]
		onlineSequences = append(onlineSequences, offlineSequences...)
		if len(onlineSequences) > d.sequenceLength {
			onlineSequences = onlineSequences[:d.sequenceLength]
		}
	}

	// seqeunce feature correspond to easyrec processor
	sequencesValueMap := make(map[string][]string)
	sequenceMap := make(map[string]bool, 0)
	for _, seq := range onlineSequences {
		key := fmt.Sprintf("%s#%s", seq.itemId, seq.event)
		if _, exist := sequenceMap[key]; !exist {
			sequenceMap[key] = true
			sequencesValueMap[item_feature_key_name] = append(sequencesValueMap[item_feature_key_name], seq.itemId)
			sequencesValueMap[timestamp_feature_key_name] = append(sequencesValueMap[timestamp_feature_key_name], fmt.Sprintf("%d", seq.timestamp))
			sequencesValueMap[event_feature_key_name] = append(sequencesValueMap[event_feature_key_name], seq.event)
			sequencesValueMap[play_time_feature_key_name] = append(sequencesValueMap[play_time_feature_key_name], fmt.Sprintf("%.2f", seq.playTime))
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

func (d *FeatureTablestoreDao) itemsFeatureFetch(items []*Item, context *context.RecommendContext) {
	fk := d.featureKey
	if fk != "item:id" {
		comms := strings.Split(d.featureKey, ":")
		if len(comms) < 2 {
			log.Error(fmt.Sprintf("requestId=%s\tevent=itemsFeatureFetch\terror=featureKey error(%s)", context.RecommendId, d.featureKey))
			return
		}

		fk = comms[1]
	}

	cpuCount := utils.MaxInt(int(math.Ceil(float64(len(items))/float64(100))), 1)
	maps := make(map[int][]*Item)
	for i, item := range items {
		maps[i%cpuCount] = append(maps[i%cpuCount], item)
	}

	requestCh := make(chan []*Item, cpuCount)
	defer close(requestCh)

	for _, itemlist := range maps {
		requestCh <- itemlist
	}

	var wg sync.WaitGroup
	for i := 0; i < cpuCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case itemlist := <-requestCh:
				// var keys []interface{}
				key2Item := make(map[string]*Item, len(itemlist))
				batchGetReq := &tablestore.BatchGetRowRequest{}
				mqCriteria := &tablestore.MultiRowQueryCriteria{}
				for _, item := range itemlist {
					var key string
					if fk == "item:id" {
						key = string(item.Id)
					} else {
						key = item.StringProperty(fk)
					}
					key2Item[key] = item

					// keys = append(keys, key)
					pkToGet := new(tablestore.PrimaryKey)
					pkToGet.AddPrimaryKeyColumn(d.itemFeatureKeyName, key)
					mqCriteria.AddRow(pkToGet)
				}
				mqCriteria.MaxVersion = 1
				if d.itemSelectFields != "" {
					mqCriteria.ColumnsToGet = strings.Split(d.itemSelectFields, ",")
				}
				mqCriteria.TableName = d.table
				batchGetReq.MultiRowQueryCriteria = append(batchGetReq.MultiRowQueryCriteria, mqCriteria)
				batchGetResponse, err := d.tablestore.Client.BatchGetRow(batchGetReq)

				if err != nil {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureTablestoreDao\terror=%v", context.RecommendId, err))
					return
				}

				rowsResult, ok := batchGetResponse.TableToRowsResult[d.table]
				if !ok {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureTablestoreDao\ttable row result empty", context.RecommendId))
					return
				}

				for _, row := range rowsResult {
					if !row.IsSucceed {
						continue
					}
					var key string
					if len(row.PrimaryKey.PrimaryKeys) > 0 {
						pkColumn := row.PrimaryKey.PrimaryKeys[0]
						key = pkColumn.Value.(string)
					}
					if key == "" {
						continue
					}
					item := key2Item[key]
					properties := make(map[string]interface{}, len(row.Columns))

					for _, column := range row.Columns {
						name := column.ColumnName
						switch val := column.Value.(type) {
						case string:
							properties[name] = val
						case int:
							properties[name] = val
						case float64:
							properties[name] = val
						default:
							properties[name] = val
						}

					}

					item.AddProperties(properties)
				}

			default:
			}
		}()
	}
	wg.Wait()
}
