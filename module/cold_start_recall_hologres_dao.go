package module

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/persist/holo"
	"github.com/alibaba/pairec/recconf"
)

type ColdStartRecallHologresDao struct {
	recallCount   int
	timeInterval  int
	db            *sql.DB
	itemType      string
	recallName    string
	table         string
	whereClause   string
	itemFieldName string

	mu      sync.RWMutex
	sqlStmt *sql.Stmt
}

func NewColdStartRecallHologresDao(config recconf.RecallConfig) *ColdStartRecallHologresDao {
	hologres, err := holo.GetPostgres(config.ColdStartDaoConf.HologresName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}

	dao := &ColdStartRecallHologresDao{
		recallCount:   config.RecallCount,
		db:            hologres.DB,
		table:         config.ColdStartDaoConf.HologresTableName,
		itemType:      config.ItemType,
		recallName:    config.Name,
		timeInterval:  config.ColdStartDaoConf.TimeInterval,
		whereClause:   config.ColdStartDaoConf.WhereClause,
		itemFieldName: config.ColdStartDaoConf.PrimaryKey,
	}
	return dao
}

func (d *ColdStartRecallHologresDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select(d.itemFieldName)
	builder.From(d.table)

	where := d.whereClause
	createTime := time.Now().Add(time.Duration(-1*d.timeInterval) * time.Second)
	where = strings.ReplaceAll(where, "${time}", "'"+createTime.Format(time.RFC3339)+"'")
	if where != "" {
		builder.Where(where)
	}
	builder.OrderBy("random()")
	builder.Limit(d.recallCount)
	sqlquery, args := builder.Build()
	rows, err := d.db.Query(sqlquery, args...)
	if err != nil {
		log.Error(fmt.Sprintf("module=ColdStartRecallHologresDao\terror=%v", err))
		return
	}
	itemIds := make([]string, 0, d.recallCount)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			itemIds = append(itemIds, id)
		}
	}
	rows.Close()

	if len(itemIds) == 0 {
		return
	}

	for _, id := range itemIds {
		item := NewItem(id)
		item.ItemType = d.itemType
		item.RetrieveId = d.recallName
		ret = append(ret, item)
	}

	return
}
