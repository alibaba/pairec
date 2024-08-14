package holo

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
)

func init() {
	//sql.Register("hologres", &HologresDriver{})
}

type HologresDriver struct {
	driver pq.Driver
}

func (d HologresDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.driver.Open(name)
	if err != nil {
		return nil, err
	}

	if stmt, err := conn.Prepare("set statement_timeout = 500"); err == nil {
		stmt.Exec(nil)
		stmt.Close()
	}
	return conn, err
}

type Postgres struct {
	DSN  string
	DB   *sql.DB
	Name string
}

var postgresqlInstances = make(map[string]*Postgres)

func GetPostgres(name string) (*Postgres, error) {
	if _, ok := postgresqlInstances[name]; !ok {
		return nil, fmt.Errorf("Postgres not found, name:%s", name)
	}

	return postgresqlInstances[name], nil
}
func (m *Postgres) Init() error {
	db, err := sql.Open("hologres", m.DSN)
	if err != nil {
		return err
	}

	db.SetConnMaxLifetime(60 * time.Minute)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(100)

	m.DB = db
	err = m.DB.Ping()
	go m.loopDBStats()
	return err
}
func (m *Postgres) loopDBStats() {
	for {
		stat := m.DB.Stats()
		j, _ := json.Marshal(stat)
		log.Info(fmt.Sprintf("event=dbstat\tname=%s\tmsg=%s", m.Name, string(j)))

		time.Sleep(10 * time.Second)
	}
}
func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.HologresConfs {
		if _, ok := postgresqlInstances[name]; ok {
			continue
		}
		m := &Postgres{
			DSN:  conf.DSN,
			Name: name,
		}
		err := m.Init()
		if err != nil {
			panic(err)
		}
		postgresqlInstances[name] = m
	}
}
