package module

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/log"

	"github.com/huandu/go-sqlbuilder"
	"github.com/alibaba/pairec/persist/holo"
	"github.com/alibaba/pairec/recconf"
)

type VectorHologresDao struct {
	db             *sql.DB
	table          string
	embeddingField string
	keyField       string
	mu             sync.RWMutex
	dbStmt         *sql.Stmt
}

// NewVectorHologresDao create new VectorHologresDao
func NewVectorHologresDao(config recconf.RecallConfig) *VectorHologresDao {
	hologres, err := holo.GetPostgres(config.VectorDaoConf.HologresName)
	if err != nil {
		panic(err)
	}
	dao := &VectorHologresDao{
		db:             hologres.DB,
		table:          config.VectorDaoConf.HologresTableName,
		embeddingField: config.VectorDaoConf.EmbeddingField,
		keyField:       config.VectorDaoConf.KeyField,
	}

	go func(dao *VectorHologresDao) {
		partition := "{partition}"
		for {
			hologresName := config.VectorDaoConf.HologresName
			table := config.VectorDaoConf.PartitionInfoTable
			field := config.VectorDaoConf.PartitionInfoField
			if config.RecallType == "HologresVectorRecall" && table != "" && field != "" {
				newPartition := FetchPartition(hologresName, table, field)
				if newPartition != "" && newPartition != partition {
					dao.table = strings.Replace(dao.table, partition, newPartition, -1)
					partition = newPartition

					dao.mu.Lock()
					dao.dbStmt = nil
					dao.mu.Unlock()
				}
				time.Sleep(time.Minute)
			} else {
				break
			}
		}
	}(dao)

	return dao
}

// returns vector of string
func (d *VectorHologresDao) VectorString(id string) (string, error) {
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
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
		log.Error(fmt.Sprintf("module=VectorHologresDao\tevent=VectorString\terror=%v", err))
		//d.mu.Lock()
		//if d.dbStmt != nil {
		//	d.dbStmt.Close()
		//}
		//d.dbStmt = nil
		//d.mu.Unlock()
		return "", err
	}
	var embedding string
	for rows.Next() {
		if err := rows.Scan(&embedding); err != nil {
			return embedding, err
		}
	}

	if embedding == "" {
		return embedding, VectoryEmptyError
	}

	return embedding, nil
}

func FetchPartition(hologresName, table, field string) string {
	hologres, err := holo.GetPostgres(hologresName)
	if err != nil {
		log.Error(fmt.Sprintf("module=VectorHologresDao\tevent=FetchPartition\terror=%v", err))
		return ""
	}
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select(field)
	builder.From(table)
	builder.Limit(1)
	sqlquery, args := builder.Build()
	stmt, err := hologres.DB.Prepare(sqlquery)
	if err != nil {
		log.Error(fmt.Sprintf("module=VectorHologresDao\tevent=FetchPartition\terror=%v", err))
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		log.Error(fmt.Sprintf("module=VectorHologresDao\tevent=FetchPartition\terror=%v", err))
		return ""
	}
	var partition string

	for rows.Next() {
		if err := rows.Scan(&partition); err != nil {
			log.Error(fmt.Sprintf("module=VectorHologresDao\tevent=FetchPartition\terror=%v", err))
			return ""
		}
		return partition
	}

	return ""
}
