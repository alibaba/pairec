package module

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/clickhouse"
	"github.com/alibaba/pairec/v2/recconf"
)

type VectorClickHouseDao struct {
	db             *sql.DB
	table          string
	embeddingField string
	keyField       string
	mu             sync.RWMutex
	dbStmt         *sql.Stmt
}

func NewVectorClickHouseDao(config recconf.RecallConfig) *VectorClickHouseDao {
	clickhouseDB, err := clickhouse.GetClickHouse(config.VectorDaoConf.ClickHouseName)
	if err != nil {
		log.Error(fmt.Sprintf("get clickhouse error:%v", err))
		return nil
	}

	dao := &VectorClickHouseDao{
		db:             clickhouseDB.DB,
		table:          config.VectorDaoConf.ClickHouseTableName,
		embeddingField: config.VectorDaoConf.EmbeddingField,
		keyField:       config.VectorDaoConf.KeyField,
	}

	return dao
}

func (d *VectorClickHouseDao) VectorString(id string) (string, error) {
	builder := sqlbuilder.MySQL.NewSelectBuilder()
	builder.Select(d.embeddingField)
	builder.From(d.table)
	builder.Where(builder.Equal(d.keyField, id))

	sqlquery, args := builder.Build()
	if d.dbStmt == nil {
		d.mu.Lock()
		if d.dbStmt == nil {
			stmt, err := d.db.Prepare(sqlquery)
			if err != nil {
				d.mu.Unlock()
				return "", err
			}
			d.dbStmt = stmt
			d.mu.Unlock()
		} else {
			d.mu.Unlock()
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	rows, err := d.dbStmt.QueryContext(ctx, args...)
	if err != nil {
		log.Error(fmt.Sprintf("module=VectorClickHouseDao\tevent=VectorString\terror=%v", err))
		return "", err
	}
	defer rows.Close()
	var embedding string
	for rows.Next() {
		if err := rows.Scan(&embedding); err != nil {
			return conversionEmbeddingFormat(embedding), err
		}
	}

	if embedding == "" {
		return embedding, VectoryEmptyError
	}

	return conversionEmbeddingFormat(embedding), nil
}
