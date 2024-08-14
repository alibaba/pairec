package hologres

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/lib/pq"
)

func init() {
	sql.Register("hologres", &HologresDriver{})
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

type Hologres struct {
	DSN  string
	DB   *sql.DB
	Name string
}

var hologresInstances = make(map[string]*Hologres)

func GetHologres(name string) (*Hologres, error) {
	if _, ok := hologresInstances[name]; !ok {
		return nil, fmt.Errorf("Hologres not found, name:%s", name)
	}

	return hologresInstances[name], nil
}
func (m *Hologres) Init() error {
	db, err := sql.Open("hologres", m.DSN)
	if err != nil {
		return err
	}

	db.SetConnMaxLifetime(60 * time.Minute)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(100)

	m.DB = db
	err = m.DB.Ping()
	//go m.loopDBStats()
	return err
}

func RegisterHologres(name, dsn string) {
	if _, ok := hologresInstances[name]; ok {
		return
	}
	m := &Hologres{
		DSN:  dsn,
		Name: name,
	}
	err := m.Init()
	if err != nil {
		fmt.Printf("event=RegisterHologres\tdsn=%s\tname=%s", dsn, name)
		panic(err)
	}
	hologresInstances[name] = m

}
