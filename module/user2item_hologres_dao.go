package module

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"

	"github.com/huandu/go-sqlbuilder"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
)

type User2ItemHologresDao struct {
	db          *sql.DB
	userTable   string
	itemTable   string
	itemType    string
	recallName  string
	mu          sync.RWMutex
	userStmt    *sql.Stmt
	itemStmtMap map[int]*sql.Stmt
}

func NewUser2ItemHologresDao(config recconf.RecallConfig) *User2ItemHologresDao {
	hologres, err := holo.GetPostgres(config.User2ItemDaoConf.HologresName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}
	dao := &User2ItemHologresDao{
		db:         hologres.DB,
		userTable:  config.User2ItemDaoConf.User2ItemTable,
		itemTable:  config.User2ItemDaoConf.Item2ItemTable,
		itemType:   config.ItemType,
		recallName: config.Name,
	}
	return dao
}

func (d *User2ItemHologresDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	uid := string(user.Id)
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select("item_ids")
	builder.From(d.userTable)
	builder.Where(builder.Equal("uid", uid))

	sqlquery, args := builder.Build()
	if d.userStmt == nil {
		d.mu.Lock()
		if d.userStmt == nil {
			stmt, err := d.db.Prepare(sqlquery)
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemHologresDao\terror=hologres error(%v)", context.RecommendId, err))
				d.mu.Unlock()
				return
			}
			d.userStmt = stmt
			d.mu.Unlock()
		} else {
			d.mu.Unlock()
		}
	}
	rows, err := d.userStmt.Query(args...)
	if err != nil {
		d.mu.Lock()
		if d.userStmt != nil {
			d.userStmt.Close()
		}
		d.userStmt = nil
		d.mu.Unlock()
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemHologresDao\terror=hologres error(%v)", context.RecommendId, err))
		return
	}

	itemIds := make([]string, 0)
	for rows.Next() {
		var ids string
		if err := rows.Scan(&ids); err == nil {
			idList := strings.Split(ids, ",")
			for _, id := range idList {
				if len(id) > 0 {
					itemIds = append(itemIds, id)
				}
			}
		}
	}
	rows.Close()

	if len(itemIds) == 0 {
		log.Info(fmt.Sprintf("module=User2ItemHologresDao\tuid=%s\terr=item ids empty", uid))
		return
	}

	if len(itemIds) > 100 {
		rand.Shuffle(len(itemIds)/2, func(i, j int) {
			itemIds[i], itemIds[j] = itemIds[j], itemIds[i]
		})

		itemIds = itemIds[:100]
	}

	cpuCount := 4
	maps := make(map[int][]interface{})
	for i, id := range itemIds {
		//maps[i%cpuCount] = append(maps[i%cpuCount], "'"+id+"'")
		maps[i%cpuCount] = append(maps[i%cpuCount], id)
	}

	itemIdCh := make(chan []interface{}, cpuCount)
	for _, ids := range maps {
		itemIdCh <- ids
	}

	itemCh := make(chan []*Item, cpuCount)
	for i := 0; i < cpuCount; i++ {
		go func() {
			result := make([]*Item, 0)
		LOOP:
			for {
				select {
				case ids := <-itemIdCh:
					builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
					builder.Select("similar_item_ids")
					builder.From(d.itemTable)
					builder.Where(builder.In("item_id", ids...))

					sqlquery, args := builder.Build()
					stmtkey := len(ids)
					stmt := d.getItemStmt(stmtkey)
					if stmt == nil {
						d.mu.Lock()
						stmt = d.itemStmtMap[stmtkey]
						if stmt == nil {
							stmt2, err := d.db.Prepare(sqlquery)
							if err != nil {
								log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemHologresDao\terror=hologres error(%v)", context.RecommendId, err))
								d.mu.Unlock()
								goto LOOP
							}
							d.itemStmtMap[stmtkey] = stmt2
							stmt = stmt2
							d.mu.Unlock()
						} else {
							d.mu.Unlock()
						}
					}

					rows, err := stmt.Query(args...)
					if err != nil {
						log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemHologresDao\tsql=%s\terror=%v", context.RecommendId, sqlquery, err))
						goto LOOP
					}
					for rows.Next() {
						var ids string
						if err := rows.Scan(&ids); err != nil {
							log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemHologresDao\terror=%v", context.RecommendId, err))
							goto LOOP
						}

						list := strings.Split(ids, ",")
						for _, str := range list {
							strs := strings.Split(str, ":")
							if len(strs) == 2 && len(strs[0]) > 0 && strs[0] != "null" {
								item := NewItem(strs[0])
								item.RetrieveId = d.recallName
								item.ItemType = d.itemType
								if tmpScore, err := strconv.ParseFloat(strs[1], 64); err == nil {
									item.Score = tmpScore
								}

								result = append(result, item)
							}

						}

					}
					rows.Close()
				default:
					goto DONE

				}
			}
		DONE:
			itemCh <- result
		}()
	}
	for i := 0; i < cpuCount; i++ {
		items := <-itemCh
		ret = append(ret, items...)
	}
	close(itemCh)
	close(itemIdCh)
	return
}
func (d *User2ItemHologresDao) getItemStmt(key int) *sql.Stmt {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.itemStmtMap[key]
}
