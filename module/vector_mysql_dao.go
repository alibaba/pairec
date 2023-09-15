package module

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/persist/mysqldb"
	"github.com/alibaba/pairec/recconf"
)

type VectorMysqlDao struct {
	db             *sql.DB
	table          string
	embeddingField string
	keyField       string
	mu             sync.RWMutex
	dbStmt         *sql.Stmt
}

func NewVectorMysqlDao(config recconf.RecallConfig) *VectorMysqlDao {
	mysql, err := mysqldb.GetMysql(config.VectorDaoConf.MysqlName)
	if err != nil {
		panic(err)
	}

	dao := &VectorMysqlDao{
		db:             mysql.DB,
		table:          config.VectorDaoConf.MysqlTable,
		embeddingField: config.VectorDaoConf.EmbeddingField,
		keyField:       config.VectorDaoConf.KeyField,
	}

	return dao
}

func (d *VectorMysqlDao) VectorString(id string) (string, error) {
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
		log.Error(fmt.Sprintf("module=VectorMysqlDao\tevent=VectorString\terror=%v", err))
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

func conversionEmbeddingFormat(emb string) string {
	vectors := strings.Split(emb, ",")
	for i, v := range vectors {
		vectors[i] = fmt.Sprintf("%d:%s", i+1, v)
	}
	return strings.Join(vectors, " ")
}
