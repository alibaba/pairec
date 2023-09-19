package module

import (
	"database/sql"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"

	"github.com/huandu/go-sqlbuilder"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type UserGlobalHotRecallHologresDao struct {
	db          *sql.DB
	itemType    string
	recallName  string
	table       string
	recallCount int
	mu          sync.RWMutex
	userStmt    *sql.Stmt
}

func NewUserGlobalHotRecallHologresDao(config recconf.RecallConfig) *UserGlobalHotRecallHologresDao {
	hologres, err := holo.GetPostgres(config.DaoConf.HologresName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}

	dao := &UserGlobalHotRecallHologresDao{
		recallCount: config.RecallCount,
		db:          hologres.DB,
		table:       config.DaoConf.HologresTableName,
		itemType:    config.ItemType,
		recallName:  config.Name,
	}
	return dao
}

func (d *UserGlobalHotRecallHologresDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	uid := string(user.Id)
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select("item_ids")
	builder.From(d.table)
	builder.Where(builder.Equal("trigger_id", "-1"))

	sqlquery, args := builder.Build()
	if d.userStmt == nil {
		d.mu.Lock()
		if d.userStmt == nil {
			stmt, err := d.db.Prepare(sqlquery)
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=UserGlobalHotRecallHologresDao\terror=hologres error(%v)", context.RecommendId, err))
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
		log.Error(fmt.Sprintf("requestId=%s\tmodule=UserGlobalHotRecallHologresDao\tuid=%s\terror=%v", context.RecommendId, uid, err))
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

	for _, id := range itemIds {
		strs := strings.Split(id, ":")
		if len(strs) == 1 {
			// itemid
			item := NewItem(strs[0])
			item.ItemType = d.itemType
			item.RetrieveId = d.recallName
			item.Score = rand.Float64()
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
			item.Score = rand.Float64()
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
			item.AddAlgoScore("hot_score", item.Score)
			ret = append(ret, item)
		}
	}

	if len(ret) > d.recallCount {
		sort.Sort(sort.Reverse(ItemScoreSlice(ret)))
		return ret[:d.recallCount]
	}
	return
}
