package mysqldb

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
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
func RegisterMysql(name, dsn string) {
	if _, ok := mysqlInstances[name]; !ok {
		m := &Mysql{
			DSN: dsn,
		}
		err := m.Init()
		if err != nil {
			panic(err)
		}
		mysqlInstances[name] = m
	}
}
