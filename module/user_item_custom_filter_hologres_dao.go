package module

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/goburrow/cache"
	"github.com/huandu/go-sqlbuilder"
)

type User2ItemCustomFilterHologresDao struct {
	table    string
	db       *sql.DB
	cache    cache.Cache
	mu       sync.RWMutex
	userStmt *sql.Stmt
}

func NewUser2ItemCustomFilterHologresDao(config recconf.FilterConfig) *User2ItemCustomFilterHologresDao {
	hologres, err := holo.GetPostgres(config.DaoConf.HologresName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}
	dao := &User2ItemCustomFilterHologresDao{
		db:    hologres.DB,
		table: config.DaoConf.HologresTableName,
	}
	if config.ItemStateCacheSize > 0 {
		cacheTime := 3600
		if config.ItemStateCacheTime > 0 {
			cacheTime = config.ItemStateCacheTime
		}
		dao.cache = cache.New(cache.WithMaximumSize(config.ItemStateCacheSize),
			cache.WithExpireAfterAccess(time.Second*time.Duration(cacheTime)))
	}
	return dao
}

func (d *User2ItemCustomFilterHologresDao) Filter(uid UID, items []*Item, ctx *context.RecommendContext) (ret []*Item) {
	var itemIds []string
	if d.cache != nil {
		if list, ok := d.cache.GetIfPresent(uid); ok {
			itemIds = list.([]string)
		} else {
			itemIds = d.fetchFromHologres(uid, ctx)
		}
	} else {
		itemIds = d.fetchFromHologres(uid, ctx)
	}

	if len(itemIds) == 0 {
		ret = items
		return
	}

	fiterIds := make(map[string]bool, len(itemIds))

	for _, id := range itemIds {
		fiterIds[id] = true
	}

	for _, item := range items {
		if _, ok := fiterIds[string(item.Id)]; !ok {
			ret = append(ret, item)
		}
	}
	if d.cache != nil {
		d.cache.Put(uid, itemIds)
	}
	return
}

func (d *User2ItemCustomFilterHologresDao) fetchFromHologres(uid UID, ctx *context.RecommendContext) (itemIds []string) {
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select("item_ids")
	builder.From(d.table)

	builder.Where(builder.Equal("user_id", string(uid)))

	sqlquery, args := builder.Build()
	if d.userStmt == nil {
		d.mu.Lock()
		if d.userStmt == nil {
			stmt, err := d.db.Prepare(sqlquery)
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemCustomFilterHologresDao\terror=hologres error(%v)", ctx.RecommendId, err))
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
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemCustomFilterHologresDao\tuid=%s\terror=%v", ctx.RecommendId, uid, err))
		return
	}
	defer rows.Close()
	itemIds = make([]string, 0, 100)
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
	return
}
