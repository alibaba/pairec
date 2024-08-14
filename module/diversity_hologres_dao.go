package module

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/goburrow/cache"
	"github.com/huandu/go-sqlbuilder"
	pctx "github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/alibaba/pairec/v2/utils/sqlutil"
	"strconv"
	"strings"
	"time"
)

type DiversityHologresDao struct {
	db             *sql.DB
	table          string
	cache          cache.Cache
	itemKeyField   string
	distinctFields []string
}

func NewDiversityHologresDao(config recconf.FilterConfig) *DiversityHologresDao {
	pg, err := holo.GetPostgres(config.DiversityDaoConf.HologresName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}
	cacheTime := 70
	if config.DiversityDaoConf.CacheTimeInMinutes > 0 {
		cacheTime = config.DiversityDaoConf.CacheTimeInMinutes
	}
	d := &DiversityHologresDao{
		db:             pg.DB,
		table:          config.DiversityDaoConf.HologresTableName,
		itemKeyField:   config.DiversityDaoConf.ItemKeyField,
		distinctFields: config.DiversityDaoConf.DistinctFields,
		cache:          cache.New(cache.WithMaximumSize(10000000), cache.WithExpireAfterWrite(time.Duration(cacheTime)*time.Minute)),
	}
	return d
}

func (d *DiversityHologresDao) GetDistinctFields() []string {
	return d.distinctFields
}

func (d *DiversityHologresDao) GetDistinctValue(items []*Item, ctx *pctx.RecommendContext) error {
	itemMap := make(map[ItemId]*Item)

	itemIds := make([]interface{}, 0, len(items))
	for _, item := range items {
		if distinct, ok := d.cache.GetIfPresent(item.Id); ok {
			values := distinct.(map[string]interface{})
			item.AddProperties(values)
		} else {
			itemIds = append(itemIds, item.Id)
			itemMap[item.Id] = item
		}
	}

	if len(itemIds) == 0 {
		return nil
	}

	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select(d.itemKeyField + "," + strings.Join(d.distinctFields, ","))
	sb.From(d.table)
	sb.Where(sb.In(d.itemKeyField, itemIds...))

	querySql, args := sb.Build()
	ctx.LogDebug("DiversityHologresDao sql:" + querySql)
	ctx.LogDebug(fmt.Sprintf("DiversityHologresDao sql args: %v", args))

	sqlCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	rows, err := d.db.QueryContext(sqlCtx, querySql, args...)
	if err != nil {
		ctx.LogError(fmt.Sprintf("module=DiversityHologresDao\terror=hologres error(%v)", err))
		return err
	}
	defer rows.Close()

	columns, err := rows.ColumnTypes()
	if err != nil {
		ctx.LogError(fmt.Sprintf("module=DiversityHologresDao\terror=hologres error(%v)", err))
		return err
	}
	count := 0
	values := sqlutil.ColumnValues(columns)
	for rows.Next() {
		if err = rows.Scan(values...); err != nil {
			ctx.LogError(fmt.Sprintf("module=DiversityHologresDao\tscan err=%v", err))
			continue
		} else {
			distinct := make(map[string]interface{})
			var itemId ItemId = ""
			for i, column := range columns {
				name := column.Name()
				val := values[i]
				switch v := val.(type) {
				case *sql.NullString:
					if v.Valid {
						distinct[name] = v.String
					}
				case *sql.NullInt32:
					if v.Valid {
						distinct[name] = strconv.Itoa(int(v.Int32))
					}
				case *sql.NullInt64:
					if v.Valid {
						distinct[name] = strconv.Itoa(int(v.Int64))
					}
				}
			}

			if key, ok := distinct[d.itemKeyField]; ok {
				itemId = ItemId(utils.ToString(key, ""))
				delete(distinct, d.itemKeyField)
				d.cache.Put(itemId, distinct)
				if item, okey := itemMap[itemId]; okey {
					item.AddProperties(distinct)
					count++
				}
			}
		}
	}

	ctx.LogInfo(fmt.Sprintf("module=DiversityHologresDao\tload %d diversity property", count))
	return nil
}
