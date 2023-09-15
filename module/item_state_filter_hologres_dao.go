package module

import (
	gocontext "context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/persist/holo"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/utils"
	"github.com/alibaba/pairec/utils/sqlutil"
)

var (
	requestCount = 500
)

type ItemStateFilterHologresDao struct {
	db            *sql.DB
	table         string
	whereClause   string
	itemFieldName string
	selectFields  string
	filterParam   *FilterParam
	mu            sync.RWMutex
	stmtMap       map[int]*sql.Stmt
}

func NewItemStateFilterHologresDao(config recconf.FilterConfig) *ItemStateFilterHologresDao {
	hologres, err := holo.GetPostgres(config.ItemStateDaoConf.HologresName)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	dao := &ItemStateFilterHologresDao{
		db:            hologres.DB,
		table:         config.ItemStateDaoConf.HologresTableName,
		itemFieldName: config.ItemStateDaoConf.ItemFieldName,
		whereClause:   config.ItemStateDaoConf.WhereClause,
		selectFields:  config.ItemStateDaoConf.SelectFields,
		stmtMap:       make(map[int]*sql.Stmt),
	}
	if len(config.FilterParams) > 0 {
		dao.filterParam = NewFilterParamWithConfig(config.FilterParams)
	}

	return dao
}

func (d *ItemStateFilterHologresDao) getStmt(key int) *sql.Stmt {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.stmtMap[key]
}

func (d *ItemStateFilterHologresDao) Filter(user *User, items []*Item) (ret []*Item) {
	fields := make(map[string]bool, len(items))
	cpuCount := utils.MaxInt(int(math.Ceil(float64(len(items))/float64(requestCount))), 1)
	requestCh := make(chan []interface{}, cpuCount)
	maps := make(map[int][]interface{}, cpuCount)

	index := 0
	for i, item := range items {
		maps[index%cpuCount] = append(maps[index%cpuCount], string(item.Id))
		if (i+1)%requestCount == 0 {
			index++
		}
	}

	defer close(requestCh)
	for _, idlist := range maps {
		requestCh <- idlist
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	mergeFunc := func(maps map[string]bool) {
		mu.Lock()
		for k, v := range maps {
			fields[k] = v
		}
		mu.Unlock()
	}
	for i := 0; i < cpuCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case idlist := <-requestCh:
				fieldMap := make(map[string]bool, len(idlist))
				builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
				builder.Select(d.itemFieldName)
				if d.selectFields != "" {
					builder.Select(d.itemFieldName + "," + d.selectFields)
				}
				builder.From(d.table)
				if d.whereClause != "" {
					builder.Where(d.whereClause)
				}
				// for use stmt, adjust idlist length
				// if len(idlist) < 1000 && (len(idlist)%100 != 0) {
				// c := (len(idlist)/100+1)*100 - len(idlist)
				// for i := 0; i < c; i++ {
				// idlist = append(idlist, "-1")
				// }
				// }
				if len(idlist) < requestCount {
					c := requestCount - len(idlist)
					for i := 0; i < c; i++ {
						idlist = append(idlist, "-1")
					}
				}
				builder.Where(builder.In(d.itemFieldName, idlist...))

				sqlquery, args := builder.Build()
				stmtkey := len(idlist)
				stmt := d.getStmt(stmtkey)
				if stmt == nil {
					d.mu.Lock()
					stmt = d.stmtMap[stmtkey]
					if stmt == nil {
						stmt2, err := d.db.Prepare(sqlquery)
						if err != nil {
							log.Error(fmt.Sprintf("module=ItemStateFilterHologresDao\terror=hologres error(%v)", err))
							// if error , not filter item
							for _, id := range idlist {
								fieldMap[id.(string)] = true
							}
							mergeFunc(fieldMap)
							d.mu.Unlock()
							return
						}
						d.stmtMap[stmtkey] = stmt2
						stmt = stmt2
						d.mu.Unlock()
					} else {
						d.mu.Unlock()
					}
				}

				rowsChannel := make(chan *sql.Rows, 1)
				ctx, cancel := gocontext.WithTimeout(gocontext.Background(), 200*time.Millisecond)
				defer cancel()
				// async invoke sql query
				go func() {
					rows, err := stmt.Query(args...)
					if err != nil {
						rowsChannel <- nil
						log.Error(fmt.Sprintf("module=ItemStateFilterHologresDao\terror=hologres error(%v)", err))
						return
					}

					// check query is timeout
					select {
					case <-ctx.Done():
						if rows != nil {
							rows.Close()
						}
						return
					default:
					}

					rowsChannel <- rows
					return
				}()

				var rows *sql.Rows
				select {
				case <-ctx.Done():
					log.Error(fmt.Sprintf("module=ItemStateFilterHologresDao\terror=hologres error(%v)", ctx.Err()))
					for _, id := range idlist {
						fieldMap[id.(string)] = true
					}
					mergeFunc(fieldMap)
					return
				case rows = <-rowsChannel:
					if rows == nil {
						for _, id := range idlist {
							fieldMap[id.(string)] = true
						}
						mergeFunc(fieldMap)
						return
					}
				}

				defer rows.Close()
				columns, err := rows.ColumnTypes()
				if err != nil {
					log.Error(fmt.Sprintf("module=ItemStateFilterHologresDao\terror=hologres error(%v)", err))
					// if error , not filter item
					for _, id := range idlist {
						fieldMap[id.(string)] = true
					}
					mergeFunc(fieldMap)
					return
				}
				values := sqlutil.ColumnValues(columns)
				for rows.Next() {
					if err := rows.Scan(values...); err == nil {
						properties := make(map[string]interface{}, len(values))
						var id string
						for i, column := range columns {
							name := column.Name()
							val := values[i]
							if i == 0 {
								switch v := val.(type) {
								case *sql.NullString:
									if v.Valid {
										id = v.String
									}
								case *sql.NullInt32:
									if v.Valid {
										id = strconv.Itoa(int(v.Int32))
									}
								case *sql.NullInt64:
									id = utils.ToString(v.Int64, "")
								}
								if id == "" {
									break
								}
								continue
							}

							switch v := val.(type) {
							case *sql.NullString:
								if v.Valid {
									properties[name] = v.String
								}
							case *sql.NullInt32:
								if v.Valid {
									properties[name] = v.Int32
								}
							case *sql.NullInt64:
								if v.Valid {
									properties[name] = v.Int64
								}
							case *sql.NullFloat64:
								if v.Valid {
									properties[name] = v.Float64
								}
							}
						}

						if d.filterParam != nil {
							result, err := d.filterParam.Evaluate(properties)
							if err == nil && result == true {
								fieldMap[id] = true
							}
						} else {
							fieldMap[id] = true
						}
					}
				}
				mergeFunc(fieldMap)
			default:
			}
		}()
	}

	wg.Wait()

	for _, item := range items {
		if _, ok := fields[string(item.Id)]; ok {
			ret = append(ret, item)
		}
	}
	return
}
