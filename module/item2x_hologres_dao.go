package module

import (
	gocontext "context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/goburrow/cache"
	"github.com/huandu/go-sqlbuilder"
)

// Item2XHologresDao gets item property(x) values from a hologres item attribute table.
type Item2XHologresDao struct {
	db            *sql.DB
	item2XTable   string
	xKey          string
	itemKeyField  string
	cache         cache.Cache
	mu            sync.RWMutex
	item2XStmtMap map[int]*sql.Stmt
}

func NewItem2XHologresDao(config recconf.Item2XConfig) *Item2XHologresDao {
	dao := &Item2XHologresDao{
		item2XTable:   config.Item2XTable,
		xKey:          config.XKey,
		itemKeyField:  "item_id",
		cache:         newItem2XCache(config),
		item2XStmtMap: make(map[int]*sql.Stmt, 0),
	}
	if config.ItemKeyField != "" {
		dao.itemKeyField = config.ItemKeyField
	}

	hologres, err := holo.GetPostgres(config.HologresName)
	if err != nil {
		panic(fmt.Sprintf("get hologres error, name:%s, error:%v", config.HologresName, err))
	}
	dao.db = hologres.DB

	return dao
}

func (d *Item2XHologresDao) getItem2XStmt(key int) *sql.Stmt {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.item2XStmtMap[key]
}

func (d *Item2XHologresDao) ListItem2X(itemIds []string, context *context.RecommendContext) map[string]string {
	item2XMap := make(map[string]string, len(itemIds))
	if len(itemIds) == 0 {
		return item2XMap
	}

	missedIds := fetchItem2XFromCache(d.cache, itemIds, item2XMap)
	if len(missedIds) == 0 {
		return item2XMap
	}

	ids := make([]interface{}, 0, len(missedIds))
	for _, id := range missedIds {
		ids = append(ids, id)
	}

	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select(d.itemKeyField, d.xKey).
		From(d.item2XTable).
		Where(
			sb.In(d.itemKeyField, ids...),
		)
	sqlStr, args := sb.Build()

	stmtkey := len(ids)
	stmt := d.getItem2XStmt(stmtkey)
	if stmt == nil {
		d.mu.Lock()
		stmt = d.item2XStmtMap[stmtkey]
		if stmt == nil {
			stmt2, err := d.db.Prepare(sqlStr)
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=Item2XHologresDao\terror=hologres error(%v)", context.RecommendId, err))
				d.mu.Unlock()
				return item2XMap
			}
			d.item2XStmtMap[stmtkey] = stmt2
			stmt = stmt2
			d.mu.Unlock()
		} else {
			d.mu.Unlock()
		}
	}

	ctx, cancel := gocontext.WithTimeout(gocontext.Background(), 100*time.Millisecond)
	defer cancel()

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=Item2XHologresDao\tsql=%s\terror=%v", context.RecommendId, sqlStr, err))
		return item2XMap
	}
	defer rows.Close()

	for rows.Next() {
		var itemId string
		var xVal sql.NullString
		if err := rows.Scan(&itemId, &xVal); err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=Item2XHologresDao\terror=%v", context.RecommendId, err))
			continue
		}
		if xVal.Valid && xVal.String != "" {
			item2XMap[itemId] = xVal.String
		}
	}

	putItem2XToCache(d.cache, missedIds, item2XMap)

	return item2XMap
}
