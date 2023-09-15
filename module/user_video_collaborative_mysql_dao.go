package module

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/persist/mysqldb"
	"github.com/alibaba/pairec/recconf"
)

type UserVideoCollaborativeMysqlDao struct {
	db         *sql.DB
	userTable  string
	itemTable  string
	itemType   string
	recallName string

	normalization bool
}

var (
	sql_user_filter = "SELECT item_ids FROM %s WHERE user_id='%s'"
	sql_item_filter = "SELECT item_id, similar_item_ids FROM %s WHERE item_id in (%s)"
)

func NewUserVideoCollaborativeMysqlDao(config recconf.RecallConfig) *UserVideoCollaborativeMysqlDao {
	dao := &UserVideoCollaborativeMysqlDao{}
	mysql, err := mysqldb.GetMysql(config.UserCollaborativeDaoConf.MysqlName)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return nil
	}
	dao.db = mysql.DB
	dao.userTable = config.UserCollaborativeDaoConf.User2ItemTable
	dao.itemTable = config.UserCollaborativeDaoConf.Item2ItemTable
	dao.itemType = config.ItemType
	dao.recallName = config.Name
	if config.UserCollaborativeDaoConf.Normalization == "on" || config.UserCollaborativeDaoConf.Normalization == "" {
		dao.normalization = true
	}
	return dao
}

func (d *UserVideoCollaborativeMysqlDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	uid := string(user.Id)
	sql := fmt.Sprintf(sql_user_filter, d.userTable, uid)
	rows, err := d.db.Query(sql)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserVideoCollaborativeMysqlDao\tuid=%s\terror=%v", context.RecommendId, uid, err))
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
						log.Error(fmt.Sprintf("requestId=%s\tmodule=UserVideoCollaborativeMysqlDao\tevent=ParsePreferScore\tuid=%s\terr=%v", context.RecommendId, uid, err))
					}
				}
			}
		}
	}
	rows.Close()

	if len(itemIds) == 0 {
		log.Info(fmt.Sprintf("module=UserVideoCollaborativeMysqlDao\tuid=%s\terr=item ids empty", uid))
		return
	}

	if len(itemIds) > 100 {
		rand.Shuffle(len(itemIds)/2, func(i, j int) {
			itemIds[i], itemIds[j] = itemIds[j], itemIds[i]
		})

		itemIds = itemIds[:100]
	}

	cpuCount := 4
	maps := make(map[int][]string)
	for i, id := range itemIds {
		maps[i%cpuCount] = append(maps[i%cpuCount], "'"+id+"'")
	}

	itemIdCh := make(chan []string, cpuCount)
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
					sql := fmt.Sprintf(sql_item_filter, d.itemTable, strings.Join(ids, ","))
					rows, err := d.db.Query(sql)
					if err != nil {
						log.Error(fmt.Sprintf("requestId=%s\tmodule=UserVideoCollaborativeMysqlDao\tsql=%s\terror=%v", context.RecommendId, sql, err))
						goto LOOP
					}
					for rows.Next() {
						var triggerId, ids string
						if err := rows.Scan(&triggerId, &ids); err != nil {
							log.Error(fmt.Sprintf("requestId=%s\tmodule=UserVideoCollaborativeMysqlDao\terror=%v", context.RecommendId, err))
							goto LOOP
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
func (d *UserVideoCollaborativeMysqlDao) GetTriggers(user *User, context *context.RecommendContext) (itemTriggers map[string]float64) {
	itemTriggers = make(map[string]float64)
	triggerInfos := d.GetTriggerInfos(user, context)

	for _, trigger := range triggerInfos {
		itemTriggers[trigger.ItemId] = trigger.Weight
	}

	return
}

func (d *UserVideoCollaborativeMysqlDao) GetTriggerInfos(user *User, context *context.RecommendContext) (triggerInfos []*TriggerInfo) {
	uid := string(user.Id)
	sql := fmt.Sprintf(sql_user_filter, d.userTable, uid)
	rows, err := d.db.Query(sql)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserVideoCollaborativeMysqlDao\tuid=%s\terror=%v", context.RecommendId, uid, err))
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
						trigger.Weight = score
					} else {
						log.Error(fmt.Sprintf("requestId=%s\tmodule=UserVideoCollaborativeMysqlDao\tevent=ParsePreferScore\tuid=%s\terr=%v", context.RecommendId, uid, err))
					}
				}
				triggerInfos = append(triggerInfos, trigger)
			}
		}
	}
	return
}
