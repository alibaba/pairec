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

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/huandu/go-sqlbuilder"
)

type UserU2I2X2IHologresDao struct {
	db            *sql.DB
	userTable     string
	item2XTable   string
	x2ItemTable   string
	itemType      string
	recallName    string
	userStmt      *sql.Stmt
	mu            sync.RWMutex
	item2XStmtMap map[int]*sql.Stmt
	x2ItemStmtMap map[int]*sql.Stmt
	xKey          string
	xDelimiter    string

	normalization bool
}

func NewUserU2I2X2IHologresDao(config recconf.RecallConfig) *UserU2I2X2IHologresDao {
	dao := &UserU2I2X2IHologresDao{
		item2XStmtMap: make(map[int]*sql.Stmt, 0),
		x2ItemStmtMap: make(map[int]*sql.Stmt, 0),
	}
	hologres, err := holo.GetPostgres(config.UserCollaborativeDaoConf.HologresName)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return nil
	}
	dao.db = hologres.DB
	dao.userTable = config.UserCollaborativeDaoConf.User2ItemTable
	dao.item2XTable = config.UserCollaborativeDaoConf.Item2XTable
	dao.x2ItemTable = config.UserCollaborativeDaoConf.X2ItemTable
	dao.xKey = config.UserCollaborativeDaoConf.XKey
	dao.xDelimiter = config.UserCollaborativeDaoConf.XDelimiter
	dao.itemType = config.ItemType
	dao.recallName = config.Name
	if config.UserCollaborativeDaoConf.Normalization == "on" || config.UserCollaborativeDaoConf.Normalization == "" {
		dao.normalization = true
	}
	return dao
}

func (d *UserU2I2X2IHologresDao) getItem2XStmt(key int) *sql.Stmt {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.item2XStmtMap[key]
}

func (d *UserU2I2X2IHologresDao) getX2ItemStmt(key int) *sql.Stmt {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.x2ItemStmtMap[key]
}

func (d *UserU2I2X2IHologresDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
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
			log.Error(fmt.Sprintf("requestId=%s\tmodule=UserU2I2X2IHologresDao\tuid=%s\terror=%v", context.RecommendId, uid, err))
			return
		}
		d.userStmt = stmt
	}
	rows, err := d.userStmt.Query(args...)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserU2I2X2IHologresDao\tuid=%s\terror=%v", context.RecommendId, uid, err))
		return
	}
	itemIds := make([]string, 0)
	preferScoreMap := make(map[string]float64)
	xPreferScoreMap := make(map[string]float64)
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
						log.Error(fmt.Sprintf("requestId=%s\tmodule=UserU2I2X2IHologresDao\tevent=ParsePreferScore\tuid=%s\terr=%v", context.RecommendId, uid, err))
					}
				}
			}
		}
	}
	rows.Close()

	if len(itemIds) == 0 {
		log.Info(fmt.Sprintf("module=UserU2I2X2IHologresDao\tuid=%s\terr=item ids empty", uid))
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

	// get X (category etc.) of items
	xValueCh := make(chan []string, cpuCount)
	for i := 0; i < cpuCount; i++ {
		go func() {
			xValues := make([]string, 0)
		LOOP:
			for {
				select {
				case ids := <-itemIdCh:
					sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
					sb.Select("item_id", d.xKey).
						From(d.item2XTable).
						Where(
							sb.In("item_id", ids...),
						)
					sql, args := sb.Build()

					stmtkey := len(ids)
					stmt := d.getItem2XStmt(stmtkey)
					if stmt == nil {
						d.mu.Lock()
						stmt = d.item2XStmtMap[stmtkey]
						if stmt == nil {
							stmt2, err := d.db.Prepare(sql)
							if err != nil {
								log.Error(fmt.Sprintf("requestId=%s\tmodule=UserU2I2X2IHologresDao\terror=hologres error(%v)", context.RecommendId, err))
								goto LOOP
							}
							d.item2XStmtMap[stmtkey] = stmt2
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
						log.Error(fmt.Sprintf("requestId=%s\tmodule=UserU2I2X2IHologresDao\tsql=%s\terror=%v", context.RecommendId, sql, err))
						goto LOOP
					}

					for rows.Next() {
						var triggerId, xVal string
						if err := rows.Scan(&triggerId, &xVal); err != nil {
							log.Error(fmt.Sprintf("requestId=%s\tmodule=UserU2I2X2IHologresDao\terror=%v", context.RecommendId, err))
							continue
						}

						preferScore := preferScoreMap[triggerId]

						if xVal != "" {
							// an item may have many categories, split with delimiter first.
							// if a category appears multiple times, add up the scores.
							if d.xDelimiter != "" {
								for _, v := range strings.Split(xVal, d.xDelimiter) {
									xPreferScoreMap[v] += preferScore
									xValues = append(xValues, v)
								}
							} else {
								xPreferScoreMap[xVal] += preferScore
								xValues = append(xValues, xVal)
							}
						}
					}
					rows.Close()
				default:
					goto DONE

				}
			}
		DONE:
			xValueCh <- xValues
		}()
	}

	xValueMap := make(map[string]bool)

	for i := 0; i < cpuCount; i++ {
		xValues := <-xValueCh
		for _, xVal := range xValues {
			xValueMap[xVal] = true
		}
	}

	var mergedXValues []any
	for val := range xValueMap {
		mergedXValues = append(mergedXValues, val)
	}

	close(xValueCh)
	close(itemIdCh)

	maps = make(map[int][]interface{})
	for i, xVal := range mergedXValues {
		maps[i%cpuCount] = append(maps[i%cpuCount], xVal)
	}

	xValueCh2 := make(chan []interface{}, cpuCount)
	for _, xValues := range maps {
		xValueCh2 <- xValues
	}

	// get items of X (category etc.)
	itemCh := make(chan []*Item, cpuCount)
	for i := 0; i < cpuCount; i++ {
		go func() {
			result := make([]*Item, 0)
		LOOP:
			for {
				select {
				case xValues := <-xValueCh2:
					sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
					sb.Select(d.xKey, "item_id").
						From(d.x2ItemTable).
						Where(
							sb.In(d.xKey, xValues...),
						)
					sql, args := sb.Build()

					stmtkey := len(xValues)
					stmt := d.getX2ItemStmt(stmtkey)
					if stmt == nil {
						d.mu.Lock()
						stmt = d.x2ItemStmtMap[stmtkey]
						if stmt == nil {
							stmt2, err := d.db.Prepare(sql)
							if err != nil {
								log.Error(fmt.Sprintf("requestId=%s\tmodule=UserU2I2X2IHologresDao\terror=hologres error(%v)", context.RecommendId, err))
								goto LOOP
							}
							d.x2ItemStmtMap[stmtkey] = stmt2
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
						var xValue, ids string
						if err := rows.Scan(&xValue, &ids); err != nil {
							log.Error(fmt.Sprintf("requestId=%s\tmodule=UserU2I2X2IHologresDao\terror=%v", context.RecommendId, err))
							continue
						}

						preferScore := xPreferScoreMap[xValue]

						list := strings.Split(ids, ",")
						for _, str := range list {
							strs := strings.Split(str, ":")

							if len(strs[0]) > 0 && strs[0] != "null" {
								if _, ok := preferScoreMap[strs[0]]; ok { // if item id has been in trigger, ignore it
									continue
								}
							}

							item := NewItem(strs[0])
							item.RetrieveId = d.recallName
							item.ItemType = d.itemType
							item.Score = preferScore

							if len(strs) == 2 {
								if tmpScore, err := strconv.ParseFloat(strings.TrimSpace(strs[1]), 64); err == nil {
									item.Score = item.Score * tmpScore
								}
							}

							result = append(result, item)
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
	close(xValueCh2)

	return
}

func (d *UserU2I2X2IHologresDao) GetTriggers(user *User, context *context.RecommendContext) (itemTriggers map[string]float64) {
	itemTriggers = make(map[string]float64)
	triggerInfos := d.GetTriggerInfos(user, context)

	for _, trigger := range triggerInfos {
		itemTriggers[trigger.ItemId] = trigger.Weight
	}

	return
}

func (d *UserU2I2X2IHologresDao) GetTriggerInfos(user *User, context *context.RecommendContext) (triggerInfos []*TriggerInfo) {
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
			log.Error(fmt.Sprintf("requestId=%s\tmodule=UserU2I2X2IHologresDao\tuid=%s\terror=%v", context.RecommendId, uid, err))
			return
		}
		d.userStmt = stmt
	}
	rows, err := d.userStmt.Query(args...)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserU2I2X2IHologresDao\tuid=%s\terror=%v", context.RecommendId, uid, err))
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
						log.Error(fmt.Sprintf("requestId=%s\tmodule=UserU2I2X2IHologresDao\tevent=ParsePreferScore\tuid=%s\terr=%v", context.RecommendId, uid, err))
					}
				}
				triggerInfos = append(triggerInfos, trigger)
			}
		}
	}

	return

}
