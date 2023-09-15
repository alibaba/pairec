package lindorm

import (
	"database/sql"
	"fmt"
	"time"

	avatica "github.com/apache/calcite-avatica-go/v5"
	"github.com/alibaba/pairec/recconf"
)

func init() {
}

type Lindorm struct {
	Url      string
	User     string
	Password string
	Database string
	DB       *sql.DB
	Name     string
}

var lindormInstances = make(map[string]*Lindorm)

func GetLindorm(name string) (*Lindorm, error) {
	if _, ok := lindormInstances[name]; !ok {
		return nil, fmt.Errorf("lindorm not found, name:%s", name)
	}

	return lindormInstances[name], nil
}
func (m *Lindorm) Init() error {

	conn := avatica.NewConnector(m.Url).(*avatica.Connector)
	conn.Info = map[string]string{
		"user":     m.User,
		"password": m.Password,
		"database": m.Database,
	}
	db := sql.OpenDB(conn)
	// set connection pool params
	db.SetConnMaxIdleTime(10 * time.Minute)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(100)

	m.DB = db
	err := m.DB.Ping()
	return err
}

func Load(config *recconf.RecommendConfig) {
	for name, conf := range config.LindormConfs {
		if _, ok := lindormInstances[name]; ok {
			continue
		}
		m := &Lindorm{
			Url:      conf.Url,
			User:     conf.User,
			Password: conf.Password,
			Database: conf.Database,
			Name:     name,
		}
		err := m.Init()
		if err != nil {
			panic(err)
		}
		lindormInstances[name] = m
	}
}
