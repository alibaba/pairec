package module

import (
	gocontext "context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/mysqldb"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/alibaba/pairec/v2/utils/sqlutil"

	"github.com/huandu/go-sqlbuilder"
)

type FeatureMysqlDao struct {
	*FeatureBaseDao

	db                 *sql.DB
	table              string
	userFeatureKeyName string
	itemFeatureKeyName string
	userSelectFields   string
	itemSelectFields   string

	mu          sync.RWMutex
	userStmt    *sql.Stmt
	itemStmtMap map[int]*sql.Stmt
}

func NewFeatureMysqlDao(config recconf.FeatureDaoConfig) *FeatureMysqlDao {
	dao := &FeatureMysqlDao{
		FeatureBaseDao:     NewFeatureBaseDao(&config),
		table:              config.MysqlTable,
		userFeatureKeyName: config.UserFeatureKeyName,
		itemFeatureKeyName: config.ItemFeatureKeyName,
		userSelectFields:   config.UserSelectFields,
		itemSelectFields:   config.ItemSelectFields,
		itemStmtMap:        make(map[int]*sql.Stmt),
	}
	mysql, err := mysqldb.GetMysql(config.MysqlName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}
	dao.db = mysql.DB
	return dao
}

func (d *FeatureMysqlDao) FeatureFetch(user *User, items []*Item, context *context.RecommendContext) {
	if d.featureStore == Feature_Store_User {
		d.userFeatureFetch(user, context)
	} else {
		d.itemsFeatureFetch(items, context)
	}
}

func (d *FeatureMysqlDao) userFeatureFetch(user *User, context *context.RecommendContext) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureMysqlDao\terror=%v", context.RecommendId, err))
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
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureMysqlDao\terror=property not found(%s)", context.RecommendId, comms[1]))
		return
	}

	builder := sqlbuilder.MySQL.NewSelectBuilder()
	builder.Select(d.userSelectFields)
	builder.From(d.table)
	builder.Where(builder.Equal(d.userFeatureKeyName, key))

	sqlquery, args := builder.Build()
	if d.userStmt == nil {
		d.mu.Lock()
		if d.userStmt == nil {
			stmt, err := d.db.Prepare(sqlquery)
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureMysqlDao\terror=mysql error(%v)", context.RecommendId, err))
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
			log.Warning("module=FeatureMysqlDao\tevent=userFeatureFetch\ttimeout=100")
			return
		}
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureMysqlDao\terror=mysql error(%v)", context.RecommendId, err))
		return
	}

	defer rows.Close()
	columns, err := rows.ColumnTypes()
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureMysqlDao\terror=mysql error(%v)", context.RecommendId, err))
		return
	}
	values := sqlutil.ColumnValues(columns)

	for rows.Next() {
		if err := rows.Scan(values...); err == nil {
			properties := make(map[string]interface{}, len(values))
			for i, column := range columns {
				name := column.Name()
				if name == d.userFeatureKeyName {
					continue
				}

				if value := sqlutil.ParseColumnValues(values[i]); value != nil {
					properties[name] = value

				}
			}
			user.AddProperties(properties)
		} else {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureMysqlDao\terror=mysql error(%v)", context.RecommendId, err))
		}
	}
}

func (d *FeatureMysqlDao) itemsFeatureFetch(items []*Item, context *context.RecommendContext) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureMysqlDao\terror=%v", context.RecommendId, err))
			return
		}
	}()

	fk := d.featureKey
	if fk != "item:id" {
		comms := strings.Split(d.featureKey, ":")
		if len(comms) < 2 {
			log.Error(fmt.Sprintf("requestId=%s\tevent=itemsFeatureFetch\terror=featureKey error(%s)", context.RecommendId, d.featureKey))
			return
		}

		fk = comms[1]
	}

	cpuCount := utils.MaxInt(int(math.Ceil(float64(len(items))/float64(1000))), 1)
	requestCh := make(chan []*Item, cpuCount)
	defer close(requestCh)

	if cpuCount == 1 {
		requestCh <- items
	} else {
		maps := make(map[int][]*Item)
		for i, item := range items {
			maps[i%cpuCount] = append(maps[i%cpuCount], item)
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

				builder := sqlbuilder.MySQL.NewSelectBuilder()
				builder.Select(d.itemSelectFields)
				builder.From(d.table)
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
							log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureMysqlDao\terror=mysql error(%v)", context.RecommendId, err))
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
						log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureMysqlDao\terror=mysql error(%v)", context.RecommendId, err))
						return
					}

					// check query is timeout
					select {
					case <-ctx.Done():
						if rows != nil {
							rows.Close()
						}
						log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureMysqlDao\terror=%v", context.RecommendId, ctx.Err()))
						return
					default:
					}
					rowsChannel <- rows
				}()

				var rows *sql.Rows
				select {
				case <-ctx.Done():
					log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureMysqlDao\terror=mysql error(%v)", context.RecommendId, ctx.Err()))
					return
				case rows = <-rowsChannel:
					if rows == nil {
						return
					}
				}

				defer rows.Close()
				columns, err := rows.ColumnTypes()
				if err != nil {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureMysqlDao\terror=mysql error(%v)", context.RecommendId, err))
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
								switch v := val.(type) {
								case *sql.NullString:
									if v.Valid {
										key = v.String
									}
								case *sql.NullInt32:
									if v.Valid {
										key = strconv.Itoa(int(v.Int32))
									}
								case *sql.NullInt64:
									if v.Valid {
										key = strconv.Itoa(int(v.Int64))
									}
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
						log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureMysqlDao\terror=mysql error(%v)", context.RecommendId, err))
					}
				}
			default:
			}
		}()
	}
	wg.Wait()
}

func (d *FeatureMysqlDao) getItemStmt(key int) *sql.Stmt {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.itemStmtMap[key]
}
