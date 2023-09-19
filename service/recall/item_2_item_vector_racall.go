package recall

import (
	gocontext "context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type I2IVectorRecall struct {
	*BaseRecall
	db                   *sql.DB
	dao                  module.VectorDao
	table                string
	vectorEmbeddingField string
	vectorKeyField       string
	sql                  string
	mu                   sync.RWMutex
	dbStmt               *sql.Stmt
}

func NewI2IVectorRecall(config recconf.RecallConfig) *I2IVectorRecall {
	hologres, err := holo.GetPostgres(config.VectorDaoConf.HologresName)
	if err != nil {
		panic(err)
	}
	recall := &I2IVectorRecall{
		BaseRecall:           NewBaseRecall(config),
		db:                   hologres.DB,
		dao:                  module.NewVectorDao(config),
		table:                config.HologresVectorConf.VectorTable,
		vectorEmbeddingField: config.HologresVectorConf.VectorEmbeddingField,
		vectorKeyField:       config.HologresVectorConf.VectorKeyField,
	}

	recall.sql = fmt.Sprintf(hologres_vector_sql, recall.vectorKeyField, recall.vectorEmbeddingField, recall.table, "", recall.recallCount)
	return recall
}

func (r *I2IVectorRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {

	start := time.Now()

	itemId := context.GetParameter("item_id").(string)

	if r.cache != nil {
		key := r.cachePrefix + string(user.Id)
		cacheRet := r.cache.Get(key)
		if itemStr, ok := cacheRet.([]uint8); ok {
			itemIds := strings.Split(string(itemStr), ",")
			for _, id := range itemIds {
				var item *module.Item
				if strings.Contains(id, ":") {
					vars := strings.Split(id, ":")
					item = module.NewItem(vars[0])
					f, _ := strconv.ParseFloat(vars[2], 64)
					// item.AddProperty(vars[1], f)
					item.Score = f
				} else {
					item = module.NewItem(id)
				}
				item.RetrieveId = r.modelName
				item.ItemType = r.itemType
				ret = append(ret, item)
			}
			log.Info(fmt.Sprintf("requestId=%s\tmodule=HologresVectorRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
			return
		}
	}
	value, err := r.dao.VectorString(string(itemId))
	if err != nil {
		if errors.Is(err, module.VectoryEmptyError) {
			log.Info(fmt.Sprintf("requestId=%s\tmodule=HologresVectorRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
		} else {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=HologresVectorRecall\tname=%s\tcount=%d\terr=%v\tcost=%d", context.RecommendId, r.modelName, len(ret), err, utils.CostTime(start)))

		}
		return
	}

	if r.dbStmt == nil {
		r.mu.Lock()
		if r.dbStmt == nil {
			stmt, err := r.db.Prepare(r.sql)
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=HologresVectorRecall\tname=%s\terr=%v", context.RecommendId, r.modelName, err))
				r.mu.Unlock()
				return
			}
			r.dbStmt = stmt
			r.mu.Unlock()
		} else {
			r.mu.Unlock()
		}
	}

	ctx, cancel := gocontext.WithTimeout(gocontext.Background(), 100*time.Millisecond)
	defer cancel()
	rows, err := r.dbStmt.QueryContext(ctx, value)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=HologresVectorRecall\tname=%s\terr=%v", context.RecommendId, r.modelName, err))
		return
	}

	defer rows.Close()
	for rows.Next() {
		var itemId string
		var distance float64
		if err := rows.Scan(&itemId, &distance); err != nil {
			continue
		}

		item := module.NewItem(itemId)
		item.RetrieveId = r.modelName
		item.ItemType = r.itemType
		item.Score = distance

		ret = append(ret, item)
	}

	if r.cache != nil && len(ret) > 0 {
		go func() {
			key := r.cachePrefix + string(user.Id)
			var itemIds string
			for _, item := range ret {
				itemIds += fmt.Sprintf("%s:%s:%v", string(item.Id), r.modelName, item.Score) + ","
			}
			itemIds = itemIds[:len(itemIds)-1]
			cacheTime := r.cacheTime
			if cacheTime == 0 {
				cacheTime = 1800
			}
			if err := r.cache.Put(key, itemIds, time.Duration(cacheTime)*time.Second); err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=HologresVectorRecall\terror=%v",
					context.RecommendId, err))
			}
		}()
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=HologresVectorRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return
}
