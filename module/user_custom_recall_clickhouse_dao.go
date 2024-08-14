package module

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"sync"

	"github.com/huandu/go-sqlbuilder"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/clickhouse"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type UserCustomRecallClickHouseDao struct {
	db          *sql.DB
	itemType    string
	recallName  string
	table       string
	recallCount int
	mu          sync.RWMutex
	userStmt    *sql.Stmt
}

func NewUserCusteomRecallClickHouseDao(config recconf.RecallConfig) *UserCustomRecallClickHouseDao {
	clickhouseDB, err := clickhouse.GetClickHouse(config.DaoConf.ClickHouseName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}

	dao := &UserCustomRecallClickHouseDao{
		recallCount: config.RecallCount,
		db:          clickhouseDB.DB,
		table:       config.DaoConf.ClickHouseTableName,
		itemType:    config.ItemType,
		recallName:  config.Name,
	}
	return dao
}

func (d *UserCustomRecallClickHouseDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	uid := string(user.Id)
	builder := sqlbuilder.NewSelectBuilder()
	builder.Select("item_ids")
	builder.From(d.table)
	builder.Where(builder.Equal("user_id", uid))

	sqlquery, args := builder.Build()
	if d.userStmt == nil {
		d.mu.Lock()
		if d.userStmt == nil {
			stmt, err := d.db.Prepare(sqlquery)
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCustomRecallClickHouseDao\terror=clickhouse error(%v)", context.RecommendId, err))
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
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserCustomRecallClickHouseDao\tuid=%s\terror=%v", context.RecommendId, uid, err))
		return
	}
	itemIds := make([]string, 0, d.recallCount)
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
		return
	}

	if len(itemIds) > d.recallCount {
		rand.Shuffle(len(itemIds)/2, func(i, j int) {
			itemIds[i], itemIds[j] = itemIds[j], itemIds[i]
		})

		itemIds = itemIds[:d.recallCount]
	}

	for _, id := range itemIds {
		strs := strings.Split(id, ":")
		if len(strs) == 1 {
			// itemid
			item := NewItem(strs[0])
			item.ItemType = d.itemType
			item.RetrieveId = d.recallName
			ret = append(ret, item)
		} else if len(strs) == 2 {
			// itemid:RetrieveId
			item := NewItem(strs[0])
			item.ItemType = d.itemType
			if strs[1] != "" {
				item.RetrieveId = strs[1]
			} else {
				item.RetrieveId = d.recallName
			}
			ret = append(ret, item)
		} else if len(strs) == 3 {
			item := NewItem(strs[0])
			item.ItemType = d.itemType
			if strs[1] != "" {
				item.RetrieveId = strs[1]
			} else {
				item.RetrieveId = d.recallName
			}
			item.Score = utils.ToFloat(strs[2], float64(0))
			ret = append(ret, item)
		}
	}

	return
}
