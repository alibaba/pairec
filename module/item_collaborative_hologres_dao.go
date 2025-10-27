package module

import (
	gocontext "context"
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/huandu/go-sqlbuilder"
)

type ItemCollaborativeHologresDao struct {
	db          *sql.DB
	itemType    string
	recallName  string
	table       string
	recallCount int
	mu          sync.RWMutex
	userStmt    *sql.Stmt
}

func NewItemCollaborativeHologresDao(config recconf.RecallConfig) *ItemCollaborativeHologresDao {
	hologres, err := holo.GetPostgres(config.ItemCollaborativeDaoConf.HologresName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}

	dao := &ItemCollaborativeHologresDao{
		db:          hologres.DB,
		table:       config.ItemCollaborativeDaoConf.Item2ItemTable,
		itemType:    config.ItemType,
		recallName:  config.Name,
		recallCount: config.RecallCount,
	}
	return dao
}

func (d *ItemCollaborativeHologresDao) ListItemsByItem(user *User, context *context.RecommendContext) (ret []*Item) {
	// context get recommend item id
	item_id := utils.ToString(context.GetParameter("item_id"), "")
	if item_id == "" {
		return
	}
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select("item_ids")
	builder.From(d.table)
	builder.Where(builder.Equal("item_id", item_id))

	sqlquery, args := builder.Build()
	if d.userStmt == nil {
		d.mu.Lock()
		if d.userStmt == nil {
			stmt, err := d.db.Prepare(sqlquery)
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=ItemCollaborativeHologresDao\trecallName=%v\titemType=%v\terror=hologres error(%v)", context.RecommendId, d.recallName, d.itemType, err))
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
		log.Error(fmt.Sprintf("requestId=%s\tmodule=ItemCollaborativeHologresDao\titem_id=%s\trecallName=%v\titemType=%v\terror=%v", context.RecommendId, item_id, d.recallName, d.itemType, err))
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
func (d *ItemCollaborativeHologresDao) ListItemsByMultiItemIds(item *User, context *context.RecommendContext, itemIds []any) (ret map[string][]*Item) {
	return
}
