package module

import (
	"database/sql"
	"fmt"
	"github.com/huandu/go-sqlbuilder"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"time"
)

type BoostScoreByWeightByWeightHologresDao struct {
	db              *sql.DB
	itemWeightMap   map[string]float64
	tableName       string
	ItemFieldName   string
	WeightFieldName string
	timeInterval    int
	stmt            *sql.Stmt
}

func NewBoostScoreByWeightHologresDao(config recconf.SortConfig) *BoostScoreByWeightByWeightHologresDao {

	hologres, err := holo.GetPostgres(config.BoostScoreByWeightDao.HologresName)
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	dao := BoostScoreByWeightByWeightHologresDao{
		db:              hologres.DB,
		tableName:       config.BoostScoreByWeightDao.HologresTableName,
		ItemFieldName:   config.BoostScoreByWeightDao.ItemFieldName,
		WeightFieldName: config.BoostScoreByWeightDao.WeightFieldName,
		itemWeightMap:   make(map[string]float64),
		timeInterval:    config.TimeInterval,
	}
	go dao.loopLoad()
	return &dao
}

func (s *BoostScoreByWeightByWeightHologresDao) Sort(items []*Item) []*Item {

	if len(s.itemWeightMap) == 0 {
		s.itemWeightMap = s.GetItems()
	}

	for _, item := range items {
		weight, ok := s.itemWeightMap[string(item.Id)]
		if ok {
			item.Score = weight * item.Score
		}
	}

	return items
}

func (s *BoostScoreByWeightByWeightHologresDao) GetItems() map[string]float64 {
	data := make(map[string]float64)
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select(s.ItemFieldName, s.WeightFieldName)
	builder.From(s.tableName)
	sqlQuery, args := builder.Build()
	if s.stmt == nil {
		stmt, err := s.db.Prepare(sqlQuery)
		if err != nil {
			log.Error(fmt.Sprintf("module=BoostScoreByWeightHologresDao\terror=hologres error(%v)", err))
		}
		s.stmt = stmt
	}

	rows, err := s.stmt.Query(args...)
	if err != nil {
		log.Error(fmt.Sprintf("module=BoostScoreByWeightHologresDao\terror=hologres error(%v)", err))
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var itemId string
		var weight float64
		if err := rows.Scan(&itemId, &weight); err == nil {
			data[itemId] = weight
		}
	}
	return data
}

func (s *BoostScoreByWeightByWeightHologresDao) loopLoad() {
	for {
		data := s.GetItems()
		if len(data) > 0 {
			s.itemWeightMap = data
		}
		time.Sleep(time.Duration(s.timeInterval) * time.Second)
	}
}
