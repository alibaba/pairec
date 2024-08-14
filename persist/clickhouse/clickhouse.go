package clickhouse

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
)

type ClickHouse struct {
	DSN string
	DB  *sql.DB
}

var clickhouseInstances = make(map[string]*ClickHouse)

func GetClickHouse(name string) (*ClickHouse, error) {
	if _, ok := clickhouseInstances[name]; !ok {
		return nil, fmt.Errorf("ClickHouse not found, name:%s", name)
	}

	return clickhouseInstances[name], nil
}
func (m *ClickHouse) Init() error {
	db, err := sql.Open("clickhouse", m.DSN)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}

	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(100)
	m.DB = db
	return nil
}
func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.ClickHouseConfs {
		if _, ok := clickhouseInstances[name]; !ok {
			m := &ClickHouse{
				DSN: conf.DSN,
			}
			err := m.Init()
			if err != nil {
				log.Error(fmt.Sprintf("ClickHouse load error, name:%s, error:%v", name, err))
				continue
			}
			clickhouseInstances[name] = m
		}
	}
}
