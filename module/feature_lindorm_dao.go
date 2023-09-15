package module

import (
	gocontext "context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/persist/lindorm"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/utils"
	"github.com/alibaba/pairec/utils/sqlutil"

	"github.com/huandu/go-sqlbuilder"
)

type FeatureLindormDao struct {
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

func NewFeatureLindormDao(config recconf.FeatureDaoConfig) *FeatureLindormDao {
	dao := &FeatureLindormDao{
		FeatureBaseDao:     NewFeatureBaseDao(&config),
		table:              config.LindormTableName,
		userFeatureKeyName: config.UserFeatureKeyName,
		itemFeatureKeyName: config.ItemFeatureKeyName,
		userSelectFields:   config.UserSelectFields,
		itemSelectFields:   config.ItemSelectFields,
		itemStmtMap:        make(map[int]*sql.Stmt),
	}
	lindorm, err := lindorm.GetLindorm(config.LindormName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}
	dao.db = lindorm.DB
	return dao
}

func (d *FeatureLindormDao) FeatureFetch(user *User, items []*Item, context *context.RecommendContext) {
	if d.featureStore == Feature_Store_User {
		d.userFeatureFetch(user, context)
	} else {
		d.itemsFeatureFetch(items, context)
	}
}

func (d *FeatureLindormDao) userFeatureFetch(user *User, context *context.RecommendContext) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureLindormDao\terror=%v", context.RecommendId, err))
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
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureLindormDao\terror=property not found(%s)", context.RecommendId, comms[1]))
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
				log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureLindormDao\terror=mysql error(%v)", context.RecommendId, err))
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
			log.Warning("module=FeatureLindormDao\tevent=userFeatureFetch\ttimeout=100")
			return
		}
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureLindormDao\terror=mysql error(%v)", context.RecommendId, err))
		return
	}

	defer rows.Close()
	columns, err := rows.ColumnTypes()
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureLindormDao\terror=mysql error(%v)", context.RecommendId, err))
		return
	}
	values := sqlutil.ColumnValuesByDatabaseTypeName(columns)

	for rows.Next() {
		if err := rows.Scan(values...); err == nil {
			properties := make(map[string]interface{}, len(values))
			for i, column := range columns {
				name := column.Name()

				if value := sqlutil.ParseColumnValues(values[i]); value != nil {
					properties[name] = value

				}
			}
			user.AddProperties(properties)
		} else {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureLindormDao\terror=mysql error(%v)", context.RecommendId, err))
		}
	}
}

func (d *FeatureLindormDao) itemsFeatureFetch(items []*Item, context *context.RecommendContext) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureLindormDao\terror=%v", context.RecommendId, err))
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

	requestCount := 500
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

				builder := sqlbuilder.MySQL.NewSelectBuilder()
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
							log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureLindormDao\terror=mysql error(%v)", context.RecommendId, err))
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
						log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureLindormDao\terror=mysql error(%v)", context.RecommendId, err))
						return
					}

					// check query is timeout
					select {
					case <-ctx.Done():
						if rows != nil {
							rows.Close()
						}
						log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureLindormDao\terror=%v", context.RecommendId, ctx.Err()))
						return
					default:
					}
					rowsChannel <- rows
				}()

				var rows *sql.Rows
				select {
				case <-ctx.Done():
					log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureLindormDao\terror=mysql error(%v)", context.RecommendId, ctx.Err()))
					return
				case rows = <-rowsChannel:
					if rows == nil {
						return
					}
				}

				defer rows.Close()
				columns, err := rows.ColumnTypes()
				if err != nil {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureLindormDao\terror=mysql error(%v)", context.RecommendId, err))
					return
				}

				values := sqlutil.ColumnValuesByDatabaseTypeName(columns)
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
						log.Error(fmt.Sprintf("requestId=%s\tmodule=FeatureLindormDao\terror=mysql error(%v)", context.RecommendId, err))
					}
				}
			default:
			}
		}()
	}
	wg.Wait()
}

func (d *FeatureLindormDao) getItemStmt(key int) *sql.Stmt {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.itemStmtMap[key]
}
