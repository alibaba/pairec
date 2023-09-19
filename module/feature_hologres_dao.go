package module

import (
	gocontext "context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/alibaba/pairec/v2/utils/sqlutil"
)

type FeatureHologresDao struct {
	*FeatureBaseDao
	hasPlayTimeField        bool
	db                      *sql.DB
	table                   string
	userFeatureKeyName      string
	itemFeatureKeyName      string
	timestampFeatureKeyName string
	eventFeatureKeyName     string
	playTimeFeatureKeyName  string
	tsFeatureKeyName        string
	userSelectFields        string
	itemSelectFields        string
	sequenceOfflineTable    string
	mu                      sync.RWMutex
	userStmt                *sql.Stmt
	itemStmtMap             map[int]*sql.Stmt
	onlineSequenceStmt      *sql.Stmt
	offlineSequenceStmt     *sql.Stmt
}

func NewFeatureHologresDao(config recconf.FeatureDaoConfig) *FeatureHologresDao {
	dao := &FeatureHologresDao{
		FeatureBaseDao:          NewFeatureBaseDao(&config),
		table:                   config.HologresTableName,
		userFeatureKeyName:      config.UserFeatureKeyName,
		itemFeatureKeyName:      config.ItemFeatureKeyName,
		timestampFeatureKeyName: config.TimestampFeatureKeyName,
		eventFeatureKeyName:     config.EventFeatureKeyName,
		playTimeFeatureKeyName:  config.PlayTimeFeatureKeyName,
		tsFeatureKeyName:        config.TsFeatureKeyName,
		userSelectFields:        config.UserSelectFields,
		itemSelectFields:        config.ItemSelectFields,
		sequenceOfflineTable:    config.SequenceOfflineTableName,
		itemStmtMap:             make(map[int]*sql.Stmt),
		hasPlayTimeField:        true,
	}
	hologres, err := holo.GetPostgres(config.HologresName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}
	dao.db = hologres.DB
	if config.NoUsePlayTimeField {
		dao.hasPlayTimeField = false
	}
	return dao
}
func (d *FeatureHologresDao) getItemStmt(key int) *sql.Stmt {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.itemStmtMap[key]
}

func (d *FeatureHologresDao) FeatureFetch(user *User, items []*Item, context *context.RecommendContext) {
	if d.featureStore == Feature_Store_User && d.featureType == Feature_Type_Sequence {
		d.userSequenceFeatureFetch(user, context)
	} else if d.featureStore == Feature_Store_User {
		d.userFeatureFetch(user, context)
	} else {
		d.itemsFeatureFetch(items, context)
	}
}

func (d *FeatureHologresDao) userFeatureFetch(user *User, context *context.RecommendContext) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=%v", context.RecommendId, err))
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
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=property not found(%s)", context.RecommendId, comms[1]))
		return
	}

	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select(d.userSelectFields)
	builder.From(d.table)
	builder.Where(builder.Equal(d.userFeatureKeyName, key))

	sqlquery, args := builder.Build()
	if d.userStmt == nil {
		d.mu.Lock()
		if d.userStmt == nil {
			stmt, err := d.db.Prepare(sqlquery)
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=hologres error(%v)", context.RecommendId, err))
				d.mu.Unlock()
				return
			}
			d.userStmt = stmt
			d.mu.Unlock()
		} else {
			d.mu.Unlock()
		}
	}

	ctx, cancel := gocontext.WithTimeout(gocontext.Background(), 100*time.Millisecond)
	defer cancel()
	rows, err := d.userStmt.QueryContext(ctx, args...)
	if err != nil {
		if errors.Is(err, gocontext.DeadlineExceeded) {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\tevent=userFeatureFetch\ttable=%s\ttimeout=100", context.RecommendId, d.table))
			return
		}
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=hologres error(%v)", context.RecommendId, err))
		return
	}

	defer rows.Close()
	columns, err := rows.ColumnTypes()
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=hologres error(%v)", context.RecommendId, err))
		return
	}
	values := sqlutil.ColumnValues(columns)

	for rows.Next() {
		if err := rows.Scan(values...); err == nil {
			properties := make(map[string]interface{}, len(values))
			for i, column := range columns {
				name := column.Name()

				if value := sqlutil.ParseColumnValues(values[i]); value != nil {
					properties[name] = value
				}
			}
			if d.cacheFeaturesName != "" {
				user.AddCacheFeatures(d.cacheFeaturesName, properties)
			} else {
				user.AddProperties(properties)
			}
		} else {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=hologres error(%v)", context.RecommendId, err))
		}
	}
}

type sequenceInfo struct {
	itemId          string
	event           string
	playTime        float64
	timestamp       int64
	dimensionFields []sql.NullString
}

func (d *FeatureHologresDao) userSequenceFeatureFetch(user *User, context *context.RecommendContext) {
	defer func() {
		if err := recover(); err != nil {
			stack := string(debug.Stack())
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=%v\tstack=%v", context.RecommendId, err, strings.ReplaceAll(stack, "\n", "\t")))
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
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=property not found(%s)", context.RecommendId, comms[1]))
		return
	}

	currTime := time.Now().Unix()
	var onlineSequences []*sequenceInfo
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

	origin_sequence_event_selections := strings.Split(d.sequenceEvent, ",")
	sequence_event_selections := make([]interface{}, len(origin_sequence_event_selections))
	for i, v := range origin_sequence_event_selections {
		sequence_event_selections[i] = v
	}

	var selectFields []string
	if d.hasPlayTimeField {
		selectFields = []string{item_feature_key_name, event_feature_key_name, play_time_feature_key_name, timestamp_feature_key_name}
	} else {
		selectFields = []string{item_feature_key_name, event_feature_key_name, timestamp_feature_key_name}
	}
	if len(d.sequenceDimFields) > 0 {
		selectFields = append(selectFields, d.sequenceDimFields...)
	}
	onlineFunc := func() {
		builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
		builder.Select(selectFields...)
		builder.From(d.table)
		where := []string{builder.Equal(d.userFeatureKeyName, key), builder.GreaterThan(timestamp_feature_key_name, currTime-86400*5)}
		if d.sequenceEvent != "" {
			if len(sequence_event_selections) > 1 {
				where = append(where, builder.In(event_feature_key_name, sequence_event_selections...))
			} else {
				where = append(where, builder.Equal(event_feature_key_name, d.sequenceEvent))
			}
		}
		builder.Where(where...).Limit(d.sequenceLength)
		builder.OrderBy(timestamp_feature_key_name).Desc()

		sqlquery, args := builder.Build()
		if d.onlineSequenceStmt == nil {
			d.mu.Lock()
			if d.onlineSequenceStmt == nil {
				stmt, err := d.db.Prepare(sqlquery)
				if err != nil {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=hologres error(%v)", context.RecommendId, err))
					d.mu.Unlock()
					return
				}
				d.onlineSequenceStmt = stmt
				d.mu.Unlock()
			} else {
				d.mu.Unlock()
			}
		}
		ctx, cancel := gocontext.WithTimeout(gocontext.Background(), 100*time.Millisecond)
		defer cancel()
		rows, err := d.onlineSequenceStmt.QueryContext(ctx, args...)
		if err != nil {
			if errors.Is(err, gocontext.DeadlineExceeded) {
				log.Warning("module=FeatureHologresDao\tevent=userSequenceFeatureFetch\ttimeout=100")
				return
			}
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=hologres error(%v)", context.RecommendId, err))
			return
		}

		defer rows.Close()
		for rows.Next() {
			seq := new(sequenceInfo)
			var dst []interface{}
			if d.hasPlayTimeField {
				dst = []interface{}{&seq.itemId, &seq.event, &seq.playTime, &seq.timestamp}
			} else {
				dst = []interface{}{&seq.itemId, &seq.event, &seq.timestamp}
			}
			if len(d.sequenceDimFields) > 0 {
				seq.dimensionFields = make([]sql.NullString, len(d.sequenceDimFields))
				for i := range seq.dimensionFields {
					dst = append(dst, &seq.dimensionFields[i])
				}
			}
			if err := rows.Scan(dst...); err == nil {
				if t, exist := d.sequencePlayTimeMap[seq.event]; exist {
					if seq.playTime <= t {
						continue
					}
				}
				onlineSequences = append(onlineSequences, seq)
			} else {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=hologres error(%v)", context.RecommendId, err))
			}
		}

	}

	var offlineSequences []*sequenceInfo
	offlineFunc := func() {
		builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
		builder.Select(selectFields...)
		builder.From(d.sequenceOfflineTable)
		where := []string{builder.Equal(d.userFeatureKeyName, key)}
		if d.sequenceEvent != "" {
			if len(sequence_event_selections) > 1 {
				where = append(where, builder.In(event_feature_key_name, sequence_event_selections...))
			} else {
				where = append(where, builder.Equal(event_feature_key_name, d.sequenceEvent))
			}
		}
		builder.Where(where...).Limit(d.sequenceLength)
		builder.OrderBy(timestamp_feature_key_name).Desc()

		sqlquery, args := builder.Build()
		if d.offlineSequenceStmt == nil {
			d.mu.Lock()
			if d.offlineSequenceStmt == nil {
				stmt, err := d.db.Prepare(sqlquery)
				if err != nil {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=hologres error(%v)", context.RecommendId, err))
					d.mu.Unlock()
					return
				}
				d.offlineSequenceStmt = stmt
				d.mu.Unlock()
			} else {
				d.mu.Unlock()
			}
		}
		ctx, cancel := gocontext.WithTimeout(gocontext.Background(), 100*time.Millisecond)
		defer cancel()
		rows, err := d.offlineSequenceStmt.QueryContext(ctx, args...)
		if err != nil {
			if errors.Is(err, gocontext.DeadlineExceeded) {
				log.Warning("module=FeatureHologresDao\tevent=userSequenceFeatureFetch\ttimeout=100")
				return
			}
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=hologres error(%v)", context.RecommendId, err))
			return
		}

		defer rows.Close()
		for rows.Next() {
			seq := new(sequenceInfo)
			var dst []interface{}
			if d.hasPlayTimeField {
				dst = []interface{}{&seq.itemId, &seq.event, &seq.playTime, &seq.timestamp}
			} else {
				dst = []interface{}{&seq.itemId, &seq.event, &seq.timestamp}
			}
			if len(d.sequenceDimFields) > 0 {
				seq.dimensionFields = make([]sql.NullString, len(d.sequenceDimFields))
				for i := range seq.dimensionFields {
					dst = append(dst, &seq.dimensionFields[i])
				}
			}
			if err := rows.Scan(dst...); err == nil {
				if t, exist := d.sequencePlayTimeMap[seq.event]; exist {
					if seq.playTime <= t {
						continue
					}
				}
				offlineSequences = append(offlineSequences, seq)
			} else {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=hologres error(%v)", context.RecommendId, err))
			}
		}

	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		onlineFunc()
	}()
	if d.sequenceOfflineTable != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			offlineFunc()
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
		//user.AddProperty(curSequenceSubName, strings.Join(value, delim))
	}
	properties[d.sequenceName] = strings.Join(sequencesValueMap[item_feature_key_name], delim)

	if d.cacheFeaturesName != "" {
		user.AddCacheFeatures(d.cacheFeaturesName, properties)
	} else {
		user.AddProperties(properties)
	}
	//user.AddProperty(d.sequenceName, strings.Join(sequencesValueMap[item_feature_key_name], delim))
}

func (d *FeatureHologresDao) itemsFeatureFetch(items []*Item, context *context.RecommendContext) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=%v", context.RecommendId, err))
			return
		}
	}()

	if len(items) == 0 {
		return
	}

	fk := d.featureKey
	if fk != "item:id" {
		comms := strings.Split(d.featureKey, ":")
		if len(comms) < 2 {
			log.Error(fmt.Sprintf("requestId=%s\tevent=itemsFeatureFetch\terror=featureKey error(%s)", context.RecommendId, d.featureKey))
			return
		}

		fk = comms[1]
	}

	requestCount := 600
	cpuCount := utils.MaxInt(int(math.Ceil(float64(len(items))/float64(requestCount))), 1)
	requestCh := make(chan []*Item, cpuCount)
	defer close(requestCh)

	if cpuCount == 1 {
		requestCh <- items
	} else {
		maps := make(map[int][]*Item)
		index := 0
		for i, item := range items {
			maps[index%cpuCount] = append(maps[index%cpuCount], item)
			if (i+1)%requestCount == 0 {
				index++
			}
		}

		for _, itemlist := range maps {
			requestCh <- itemlist
		}

	}

	var wg sync.WaitGroup
	for i := 0; i < cpuCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case itemlist := <-requestCh:
				var keys []interface{}
				key2Item := make(map[string]*Item, len(itemlist))
				for _, item := range itemlist {
					var key string
					if fk == "item:id" {
						key = string(item.Id)
					} else {
						key = item.StringProperty(fk)
					}
					keys = append(keys, key)
					key2Item[key] = item
				}

				builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
				builder.Select(d.itemSelectFields)
				builder.From(d.table)
				if len(keys) < requestCount {
					c := requestCount - len(keys)
					for i := 0; i < c; i++ {
						keys = append(keys, "-1")
					}
				}
				builder.Where(builder.In(d.itemFeatureKeyName, keys...))

				sqlquery, args := builder.Build()
				stmtkey := len(keys)
				stmt := d.getItemStmt(stmtkey)
				if stmt == nil {
					d.mu.Lock()
					stmt = d.itemStmtMap[stmtkey]
					if stmt == nil {
						stmt2, err := d.db.Prepare(sqlquery)
						if err != nil {
							log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=hologres error(%v)", context.RecommendId, err))
							d.mu.Unlock()
							return
						}
						d.itemStmtMap[stmtkey] = stmt2
						stmt = stmt2
						d.mu.Unlock()
					} else {
						d.mu.Unlock()
					}
				}

				rowsChannel := make(chan *sql.Rows, 1)
				ctx, cancel := gocontext.WithTimeout(gocontext.Background(), 100*time.Millisecond)
				defer cancel()
				// async invoke sql query
				go func() {
					rows, err := stmt.Query(args...)
					if err != nil {
						rowsChannel <- nil
						log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=hologres error(%v)", context.RecommendId, err))
						return
					}

					// check query is timeout
					select {
					case <-ctx.Done():
						if rows != nil {
							rows.Close()
						}
						log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=%v", context.RecommendId, ctx.Err()))
						return
					default:
					}
					rowsChannel <- rows
				}()

				var rows *sql.Rows
				select {
				case <-ctx.Done():
					log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=hologres error(%v)", context.RecommendId, ctx.Err()))
					return
				case rows = <-rowsChannel:
					if rows == nil {
						return
					}
				}

				defer rows.Close()
				columns, err := rows.ColumnTypes()
				if err != nil {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=hologres error(%v)", context.RecommendId, err))
					return
				}

				values := sqlutil.ColumnValues(columns)
				for rows.Next() {
					if err := rows.Scan(values...); err == nil {
						var item *Item
						properties := make(map[string]interface{}, len(values))
						for i, column := range columns {
							name := column.Name()
							val := values[i]
							if i == 0 {
								var key string
								if value := sqlutil.ParseColumnValues(val); value != nil {
									key = utils.ToString(value, "")
								}

								if key == "" {
									break
								}
								item = key2Item[key]
								continue
							}

							if value := sqlutil.ParseColumnValues(val); value != nil {
								properties[name] = value
							}
						}
						if nil != item && len(properties) > 0 {
							item.AddProperties(properties)
						}

					} else {
						log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureHologresDao\terror=hologres error(%v)", context.RecommendId, err))
					}
				}
			default:
			}
		}()
	}
	wg.Wait()
}
