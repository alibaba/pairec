package dao

import (
	"database/sql"
	"fmt"
	"hash/crc32"
	"log"
	"sync"
	"time"

	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/api"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/datasource/hologres"
	"github.com/huandu/go-sqlbuilder"
)

type FeatureViewHologresDao struct {
	db              *sql.DB
	table           string
	primaryKeyField string
	eventTimeField  string
	ttl             int
	mu              sync.RWMutex
	stmtMap         map[uint32]*sql.Stmt

	offlineTable string
	onlineTable  string
}

func NewFeatureViewHologresDao(config DaoConfig) *FeatureViewHologresDao {
	dao := FeatureViewHologresDao{
		table:           config.HologresTableName,
		primaryKeyField: config.PrimaryKeyField,
		eventTimeField:  config.EventTimeField,
		ttl:             config.TTL,
		stmtMap:         make(map[uint32]*sql.Stmt, 4),
		offlineTable:    config.HologresOfflineTableName,
		onlineTable:     config.HologresOnlineTableName,
	}
	hologres, err := hologres.GetHologres(config.HologresName)
	if err != nil {
		return nil
	}

	dao.db = hologres.DB
	return &dao
}
func (d *FeatureViewHologresDao) getStmt(key uint32) *sql.Stmt {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.stmtMap[key]
}
func (d *FeatureViewHologresDao) GetFeatures(keys []interface{}, selectFields []string) ([]map[string]interface{}, error) {

	selector := make([]string, 0, len(selectFields))
	for _, field := range selectFields {
		selector = append(selector, fmt.Sprintf("\"%s\"", field))
	}
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select(selector...)
	builder.From(d.table)
	builder.Where(builder.In(fmt.Sprintf("\"%s\"", d.primaryKeyField), keys...))
	if d.ttl > 0 {
		t := time.Now().Add(time.Duration(-1 * d.ttl * int(time.Second)))
		builder.Where(builder.GreaterEqualThan(fmt.Sprintf("\"%s\"", d.eventTimeField), t))
	}

	sql, args := builder.Build()

	stmtKey := crc32.ChecksumIEEE([]byte(sql))
	//stmtKey := Md5(sql)
	stmt := d.getStmt(stmtKey)
	if stmt == nil {
		d.mu.Lock()
		stmt = d.stmtMap[stmtKey]
		if stmt == nil {
			stmt2, err := d.db.Prepare(sql)
			if err != nil {
				d.mu.Unlock()
				return nil, err
			}
			d.stmtMap[stmtKey] = stmt2
			stmt = stmt2
			d.mu.Unlock()
		} else {
			d.mu.Unlock()
		}
	}

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]map[string]interface{}, 0, len(keys))

	columns, _ := rows.ColumnTypes()
	values := ColumnValues(columns)

	for rows.Next() {
		if err := rows.Scan(values...); err == nil {
			properties := make(map[string]interface{}, len(values))
			for i, column := range columns {
				name := column.Name()

				if value := ParseColumnValues(values[i]); value != nil {
					properties[name] = value
				}
			}

			result = append(result, properties)
		}
	}

	return result, nil
}

type sequenceInfo struct {
	itemId    string
	event     string
	playTime  float64
	timestamp int64
}

func (d *FeatureViewHologresDao) GetUserSequenceFeature(keys []interface{}, userIdField string, sequenceConfig api.FeatureViewSeqConfig, onlineConfig []*api.SeqConfig) ([]map[string]interface{}, error) {
	var selectFields []string
	if sequenceConfig.PlayTimeField == "" {
		selectFields = []string{fmt.Sprintf("\"%s\"", sequenceConfig.ItemIdField), fmt.Sprintf("\"%s\"", sequenceConfig.EventField),
			fmt.Sprintf("\"%s\"", sequenceConfig.TimestampField)}
	} else {
		selectFields = []string{fmt.Sprintf("\"%s\"", sequenceConfig.ItemIdField), fmt.Sprintf("\"%s\"", sequenceConfig.EventField),
			fmt.Sprintf("\"%s\"", sequenceConfig.PlayTimeField), fmt.Sprintf("\"%s\"", sequenceConfig.TimestampField)}
	}
	currTime := time.Now().Unix()
	sequencePlayTimeMap := makePlayTimeMap(sequenceConfig)

	onlineFunc := func(seqEvent string, seqLen int, key interface{}) []*sequenceInfo {
		onlineSequences := []*sequenceInfo{}
		builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
		builder.Select(selectFields...)
		builder.From(d.onlineTable)
		where := []string{builder.Equal(fmt.Sprintf("\"%s\"", userIdField), key),
			builder.GreaterThan(fmt.Sprintf("\"%s\"", sequenceConfig.TimestampField), currTime-86400*5),
			builder.Equal(fmt.Sprintf("\"%s\"", sequenceConfig.EventField), seqEvent)}
		builder.Where(where...)
		builder.Limit(seqLen)
		builder.OrderBy(fmt.Sprintf("\"%s\"", sequenceConfig.TimestampField)).Desc()

		sql, args := builder.Build()
		stmtKey := crc32.ChecksumIEEE([]byte(sql))
		stmt := d.getStmt(stmtKey)
		if stmt == nil {
			d.mu.Lock()
			stmt = d.stmtMap[stmtKey]
			if stmt == nil {
				stmt2, err := d.db.Prepare(sql)
				if err != nil {
					d.mu.Unlock()
					log.Println(err)
					return nil
				}
				d.stmtMap[stmtKey] = stmt2
				stmt = stmt2
				d.mu.Unlock()
			} else {
				d.mu.Unlock()
			}
		}
		rows, err := stmt.Query(args...)
		if err != nil {
			log.Println(err)
			return nil
		}
		defer rows.Close()
		for rows.Next() {
			seq := new(sequenceInfo)
			var dst []interface{}
			if sequenceConfig.PlayTimeField == "" {
				dst = []interface{}{&seq.itemId, &seq.event, &seq.timestamp}
			} else {
				dst = []interface{}{&seq.itemId, &seq.event, &seq.timestamp, &seq.playTime}
			}
			if err := rows.Scan(dst...); err == nil {
				if t, exist := sequencePlayTimeMap[seqEvent]; exist {
					if seq.playTime <= t {
						continue
					}
				}
				onlineSequences = append(onlineSequences, seq)
			} else {
				log.Println(err)
				return nil
			}
		}

		return onlineSequences
	}

	offlineFunc := func(seqEvent string, seqLen int, key interface{}) []*sequenceInfo {
		offlineSequences := []*sequenceInfo{}
		builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
		builder.Select(selectFields...)
		builder.From(d.offlineTable)
		where := []string{builder.Equal(fmt.Sprintf("\"%s\"", userIdField), key),
			builder.Equal(fmt.Sprintf("\"%s\"", sequenceConfig.EventField), seqEvent)}
		builder.Where(where...)
		builder.Limit(seqLen)
		builder.OrderBy(fmt.Sprintf("\"%s\"", sequenceConfig.TimestampField)).Desc()

		sql, args := builder.Build()
		stmtKey := crc32.ChecksumIEEE([]byte(sql))
		stmt := d.getStmt(stmtKey)
		if stmt == nil {
			d.mu.Lock()
			stmt = d.stmtMap[stmtKey]
			if stmt == nil {
				stmt2, err := d.db.Prepare(sql)
				if err != nil {
					d.mu.Unlock()
					log.Println(err)
					return nil
				}
				d.stmtMap[stmtKey] = stmt2
				stmt = stmt2
				d.mu.Unlock()
			} else {
				d.mu.Unlock()
			}
		}

		rows, err := stmt.Query(args...)
		if err != nil {
			log.Println(err)
			return nil
		}
		defer rows.Close()
		for rows.Next() {
			seq := new(sequenceInfo)
			var dst []interface{}
			if sequenceConfig.PlayTimeField == "" {
				dst = []interface{}{&seq.itemId, &seq.event, &seq.timestamp}
			} else {
				dst = []interface{}{&seq.itemId, &seq.event, &seq.playTime, &seq.timestamp}
			}
			if err := rows.Scan(dst...); err == nil {
				if t, exist := sequencePlayTimeMap[seqEvent]; exist {
					if seq.playTime <= t {
						continue
					}
				}
				offlineSequences = append(offlineSequences, seq)
			} else {
				log.Println(err)
				return nil
			}
		}

		return offlineSequences

	}

	results := make([]map[string]interface{}, 0, len(keys))

	var wg sync.WaitGroup
	for _, key := range keys {
		wg.Add(1)
		go func(key interface{}) {
			defer wg.Done()
			properties := make(map[string]interface{})
			var mu sync.Mutex

			var eventWg sync.WaitGroup
			for _, seqConfig := range onlineConfig {
				eventWg.Add(1)
				go func(seqConfig *api.SeqConfig) {
					defer eventWg.Done()
					var onlineSequences []*sequenceInfo
					var offlineSequences []*sequenceInfo

					var innerWg sync.WaitGroup
					//get data from online table
					innerWg.Add(1)
					go func(seqEvent string, seqLen int, key interface{}) {
						defer innerWg.Done()
						if onlineresult := onlineFunc(seqEvent, seqLen, key); onlineresult != nil {
							onlineSequences = onlineresult
						}
					}(seqConfig.SeqEvent, seqConfig.SeqLen, key)
					//get data from offline table
					innerWg.Add(1)
					go func(seqEvent string, seqLen int, key interface{}) {
						defer innerWg.Done()
						if offlineresult := offlineFunc(seqEvent, seqLen, key); offlineresult != nil {
							offlineSequences = offlineresult
						}
					}(seqConfig.SeqEvent, seqConfig.SeqLen, key)
					innerWg.Wait()

					subproperties := makeSequenceFeatures(offlineSequences, onlineSequences, seqConfig, sequenceConfig, currTime)
					mu.Lock()
					defer mu.Unlock()
					for k, value := range subproperties {
						properties[k] = value
					}
				}(seqConfig)
			}
			eventWg.Wait()
			properties[userIdField] = key
			results = append(results, properties)
		}(key)
	}

	wg.Wait()

	return results, nil

}
