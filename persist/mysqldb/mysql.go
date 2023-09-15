package mysqldb

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/alibaba/pairec/recconf"
)

type Mysql struct {
	DSN string
	DB  *sql.DB
}

var mysqlInstances = make(map[string]*Mysql)

func GetMysql(name string) (*Mysql, error) {
	if _, ok := mysqlInstances[name]; !ok {
		return nil, fmt.Errorf("Mysql not found, name:%s", name)
	}

	return mysqlInstances[name], nil
}
func (m *Mysql) Init() error {
	db, err := sql.Open("mysql", m.DSN)
	if err != nil {
		return err
	}
	db.SetConnMaxLifetime(time.Minute)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(100)
	m.DB = db
	return nil
}
func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.MysqlConfs {
		if _, ok := mysqlInstances[name]; !ok {
			m := &Mysql{
				DSN: conf.DSN,
			}
			err := m.Init()
			if err != nil {
				panic(err)
			}
			mysqlInstances[name] = m
		}
	}
}
