package module

import (
	gocontext "context"
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
)

type UserCollaborativeHologresDao struct {
	db          *sql.DB
	userTable   string
	itemTable   string
	itemType    string
	recallName  string
	userStmt    *sql.Stmt
	mu          sync.RWMutex
	itemStmtMap map[int]*sql.Stmt

	normalization bool
}

func NewUserCollaborativeHologresDao(config recconf.RecallConfig) *UserCollaborativeHologresDao {
	dao := &UserCollaborativeHologresDao{
		itemStmtMap: make(map[int]*sql.Stmt, 0),
	}
	hologres, err := holo.GetPostgres(config.UserCollaborativeDaoConf.HologresName)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return nil
	}
	dao.db = hologres.DB
	dao.userTable = config.UserCollaborativeDaoConf.User2ItemTable
	dao.itemTable = config.UserCollaborativeDaoConf.Item2ItemTable
	dao.itemType = config.ItemType
	dao.recallName = config.Name
	if config.UserCollaborativeDaoConf.Normalization == "on" || config.UserCollaborativeDaoConf.Normalization == "" {
		dao.normalization = true
	}
	return dao
}

func (d *UserCollaborativeHologresDao) getItemStmt(key int) *sql.Stmt {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.itemStmtMap[key]
}
func (d *UserCollaborativeHologresDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	uid := string(user.Id)
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("item_ids").
		From(d.userTable).
		Where(
			sb.Equal("user_id", uid),
		)
	sqlquery, args := sb.Build()
	if d.userStmt == nil {
		stmt, err := d.db.Prepare(sqlquery)
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeHologresDao\tuid=%s\terror=%v", context.RecommendId, uid, err))
			return
		}
		d.userStmt = stmt
	}
	rows, err := d.userStmt.Query(args...)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeHologresDao\tuid=%s\terror=%v", context.RecommendId, uid, err))
		return
	}
	itemIds := make([]string, 0)
	preferScoreMap := make(map[string]float64)
	for rows.Next() {
		var ids string
		if err := rows.Scan(&ids); err == nil {
			idList := strings.Split(ids, ",")
			for _, id := range idList {
				strs := strings.Split(id, ":")
				if strs[0] == "" {
					continue
				}
				itemIds = append(itemIds, strs[0])
				preferScoreMap[strs[0]] = 1
				if len(strs) > 1 {
					if score, err := strconv.ParseFloat(strs[1], 64); err == nil {
						preferScoreMap[strs[0]] = score
					} else {
						log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeHologresDao\tevent=ParsePreferScore\tuid=%s\terr=%v", context.RecommendId, uid, err))
					}
				}
			}
		}
	}
	rows.Close()

	if len(itemIds) == 0 {
		log.Info(fmt.Sprintf("module=UserCollaborativeHologresDao\tuid=%s\terr=item ids empty", uid))
		return
	}

	if len(itemIds) > 200 {
		rand.Shuffle(len(itemIds)/2, func(i, j int) {
			itemIds[i], itemIds[j] = itemIds[j], itemIds[i]
		})

		itemIds = itemIds[:200]
	}

	cpuCount := 4
	maps := make(map[int][]interface{})
	for i, id := range itemIds {
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
					sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
					sb.Select("item_id", "similar_item_ids").
						From(d.itemTable).
						Where(
							sb.In("item_id", ids...),
						)
					sql, args := sb.Build()

					stmtkey := len(ids)
					stmt := d.getItemStmt(stmtkey)
					if stmt == nil {
						d.mu.Lock()
						stmt = d.itemStmtMap[stmtkey]
						if stmt == nil {
							stmt2, err := d.db.Prepare(sql)
							if err != nil {
								log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeHologresDao\terror=hologres error(%v)", context.RecommendId, err))
								goto LOOP
							}
							d.itemStmtMap[stmtkey] = stmt2
							stmt = stmt2
							d.mu.Unlock()
						} else {
							d.mu.Unlock()
						}
					}
					ctx, cancel := gocontext.WithTimeout(gocontext.Background(), 200*time.Millisecond)
					defer cancel()
					rows, err := stmt.QueryContext(ctx, args...)
					if err != nil {
						log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeHologresDao\tsql=%s\terror=%v", context.RecommendId, sql, err))
						goto LOOP
					}

					for rows.Next() {
						var triggerId, ids string
						if err := rows.Scan(&triggerId, &ids); err != nil {
							log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeHologresDao\terror=%v", context.RecommendId, err))
							continue
						}

						preferScore := preferScoreMap[triggerId]

						list := strings.Split(ids, ",")
						for _, str := range list {
							strs := strings.Split(str, ":")
							if len(strs) == 2 && len(strs[0]) > 0 && strs[0] != "null" {
								item := NewItem(strs[0])
								item.RetrieveId = d.recallName
								item.ItemType = d.itemType
								if tmpScore, err := strconv.ParseFloat(strings.TrimSpace(strs[1]), 64); err == nil {
									item.Score = tmpScore * preferScore
								} else {
									item.Score = preferScore
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

	ret = mergeUserCollaborativeItemsResult(itemCh, cpuCount, d.normalization)

	close(itemCh)
	close(itemIdCh)
	return
}

func (d *UserCollaborativeHologresDao) GetTriggers(user *User, context *context.RecommendContext) (itemTriggers map[string]float64) {
	itemTriggers = make(map[string]float64)
	triggerInfos := d.GetTriggerInfos(user, context)

	for _, trigger := range triggerInfos {
		itemTriggers[trigger.ItemId] = trigger.Weight
	}

	return
}

func (d *UserCollaborativeHologresDao) GetTriggerInfos(user *User, context *context.RecommendContext) (triggerInfos []*TriggerInfo) {
	uid := string(user.Id)
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("item_ids").
		From(d.userTable).
		Where(
			sb.Equal("user_id", uid),
		)
	sqlquery, args := sb.Build()
	if d.userStmt == nil {
		stmt, err := d.db.Prepare(sqlquery)
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeHologresDao\tuid=%s\terror=%v", context.RecommendId, uid, err))
			return
		}
		d.userStmt = stmt
	}
	rows, err := d.userStmt.Query(args...)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeHologresDao\tuid=%s\terror=%v", context.RecommendId, uid, err))
		return
	}
	defer rows.Close()
	for rows.Next() {
		var ids string
		if err := rows.Scan(&ids); err == nil {
			idList := strings.Split(ids, ",")
			for _, id := range idList {
				strs := strings.Split(id, ":")
				if strs[0] == "" {
					continue
				}
				trigger := &TriggerInfo{
					ItemId: strs[0],
					Weight: 1,
				}
				if len(strs) > 1 {
					if score, err := strconv.ParseFloat(strs[1], 64); err == nil {
						//itemTriggers[strs[0]] = score
						trigger.Weight = score
					} else {
						log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCollaborativeHologresDao\tevent=ParsePreferScore\tuid=%s\terr=%v", context.RecommendId, uid, err))
					}
				}
				triggerInfos = append(triggerInfos, trigger)
			}
		}
	}

	return

}
