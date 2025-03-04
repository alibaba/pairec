package module

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/huandu/go-sqlbuilder"
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
	orderBy       string

	mu          sync.RWMutex
	sqlStmt     *sql.Stmt
	whereParams []string
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
		orderBy:       config.ColdStartDaoConf.OrderBy,
	}
	if dao.whereClause != "" {
		re := regexp.MustCompile(`\$\{([a-zA-Z_][a-zA-Z0-9_.]*)\}`)

		matches := re.FindAllStringSubmatch(dao.whereClause, -1)

		for _, match := range matches {
			if len(match) > 1 {
				dao.whereParams = append(dao.whereParams, match[1])
			}
		}

	}
	return dao
}

func (d *ColdStartRecallHologresDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select(d.itemFieldName)
	builder.From(d.table)

	if d.whereClause != "" {
		where := d.parseWhere(d.whereClause, user, context)
		builder.Where(where)
	}
	if d.orderBy == "" {
		builder.OrderBy("random()")
	} else {
		builder.OrderBy(d.orderBy)
	}

	builder.Limit(d.recallCount)
	sqlquery, args := builder.Build()
	if context.Debug {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=ColdStartRecallHologresDao\tsql=%s\targs=%v", context.RecommendId, sqlquery, args))
	}

	rows, err := d.db.Query(sqlquery, args...)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=ColdStartRecallHologresDao\terror=%v", context.RecommendId, err))
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

func (d *ColdStartRecallHologresDao) parseWhere(whereSql string, user *User, context *context.RecommendContext) string {
	where := whereSql
	contextFeatures := context.GetParameter("features").(map[string]interface{})
	for _, param := range d.whereParams {
		if param == "time" { // time
			createTime := time.Now().Add(time.Duration(-1*d.timeInterval) * time.Second)
			where = strings.ReplaceAll(where, fmt.Sprintf("${%s}", param), fmt.Sprintf("'%s'", createTime.Format(time.RFC3339)))
		} else if strings.HasPrefix(param, "context.features.") { // context features from api param features
			key := strings.TrimPrefix(param, "context.features.")
			if value, ok := contextFeatures[key]; ok {
				switch value.(type) {
				case string:
					where = strings.ReplaceAll(where, fmt.Sprintf("${%s}", param), fmt.Sprintf("'%s'", value))
				default:
					where = strings.ReplaceAll(where, fmt.Sprintf("${%s}", param), fmt.Sprintf("%v", value))
				}
			}
		} else if strings.HasPrefix(param, "user.") {
			key := strings.TrimPrefix(param, "user.")

			value := user.GetProperty(key)
			if value != nil {
				switch value.(type) {
				case string:
					where = strings.ReplaceAll(where, fmt.Sprintf("${%s}", param), fmt.Sprintf("'%s'", value))
				default:
					where = strings.ReplaceAll(where, fmt.Sprintf("${%s}", param), fmt.Sprintf("%v", value))
				}
			}

		}
	}

	return where
}
