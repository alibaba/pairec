package module

import (
	"database/sql"
	"fmt"
	"github.com/huandu/go-sqlbuilder"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"time"
)

type ItemCustomFilterHoloDao struct {
	db           *sql.DB
	table        string
	filterIds    map[ItemId]bool
	stmt         *sql.Stmt
	timeInterval int
	selectCol    string
	whereClause  string
}

func NewItemCustomFilterHoloDao(config recconf.FilterConfig) *ItemCustomFilterHoloDao {
	postgres, err := holo.GetPostgres(config.DaoConf.HologresName)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return nil
	}
	dao := &ItemCustomFilterHoloDao{
		db:           postgres.DB,
		table:        config.DaoConf.HologresTableName,
		filterIds:    make(map[ItemId]bool),
		timeInterval: config.TimeInterval,
		selectCol:    config.FilterVal.SelectCol,
		whereClause:  config.FilterVal.WhereClause,
	}
	go dao.loopLoadFilterIds()
	return dao
}

func (i *ItemCustomFilterHoloDao) GetFilterItems() map[ItemId]bool {
	if len(i.filterIds) > 0 {
		return i.filterIds
	} else {
		return i.getFilterIds()
	}
}

func (i *ItemCustomFilterHoloDao) getFilterIds() (ret map[ItemId]bool) {
	ret = make(map[ItemId]bool)
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select(i.selectCol)
	sb.From(i.table)
	sb.Where(i.whereClause)
	sqlQuery, args := sb.Build()
	if i.stmt == nil {
		stmt, err := i.db.Prepare(sqlQuery)
		if err != nil {
			log.Error(fmt.Sprintf("module=ItemCustomFilterHoloDao\tevent=getFilterIds\terror=%v", err))
			return
		}
		i.stmt = stmt
	}

	rows, err := i.stmt.Query(args...)
	if err != nil {
		log.Error(fmt.Sprintf("module=ItemCustomFilterHoloDao\tevent=getFilterIds\terror=%v", err))
		return
	}
	defer rows.Close()
	for rows.Next() {
		var contentId string
		if err := rows.Scan(&contentId); err == nil {
			ret[ItemId(contentId)] = true
		}
	}
	return
}

func (i *ItemCustomFilterHoloDao) loopLoadFilterIds() {
	for {
		ret := i.getFilterIds()
		if len(ret) > 0 {
			i.filterIds = ret
		}
		time.Sleep(time.Duration(i.timeInterval) * time.Second)
	}
}
