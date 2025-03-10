package recall

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goburrow/cache"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

var (
	hologres_vector_sql_v2 = "SELECT %s, pm_approx_squared_euclidean_distance(%s,$1) as distance FROM %s %s ORDER BY distance limit %d"
)

type HologresVectorRecallV2 struct {
	*BaseRecall
	db                   *sql.DB
	dao                  module.VectorDao
	table                string
	vectorEmbeddingField string
	vectorKeyField       string
	where                string
	sql                  string
	mu                   sync.RWMutex
	dbStmt               *sql.Stmt
	userVectorCache      cache.Cache
	timeInterval         int
}

func NewHologresVectorRecallV2(config recconf.RecallConfig) *HologresVectorRecall {
	hologres, err := holo.GetPostgres(config.VectorDaoConf.HologresName)
	if err != nil {
		panic(err)
	}
	recall := &HologresVectorRecall{
		BaseRecall:           NewBaseRecall(config),
		db:                   hologres.DB,
		dao:                  module.NewVectorDao(config),
		table:                config.HologresVectorConf.VectorTable,
		vectorEmbeddingField: config.HologresVectorConf.VectorEmbeddingField,
		vectorKeyField:       config.HologresVectorConf.VectorKeyField,
		where:                config.HologresVectorConf.WhereClause,
		timeInterval:         config.HologresVectorConf.TimeInterval,
	}
	createTime := time.Now().Unix() - int64(recall.timeInterval)
	recall.where = strings.ReplaceAll(recall.where, "${time}", strconv.FormatInt(createTime, 10))
	if recall.where != "" {
		recall.where = "WHERE " + recall.where
	}

	if recall.cacheTime <= 0 {
		recall.cacheTime = 1800
	}
	recall.userVectorCache = cache.New(
		cache.WithMaximumSize(10000),
		cache.WithExpireAfterAccess(time.Duration(recall.cacheTime+100)*time.Second),
	)
	go func(recall *HologresVectorRecall) {
		partition := "{partition}"
		for {
			hologresName := config.VectorDaoConf.HologresName
			table := config.VectorDaoConf.PartitionInfoTable
			field := config.VectorDaoConf.PartitionInfoField
			if config.RecallType == "HologresVectorRecall" && table != "" && field != "" {
				newPartition := module.FetchPartition(hologresName, table, field)
				if newPartition != "" && newPartition != partition {
					recall.table = strings.Replace(recall.table, partition, newPartition, -1)
					partition = newPartition

					recall.mu.Lock()
					recall.dbStmt = nil
					recall.mu.Unlock()
				}
				time.Sleep(time.Minute)
			} else {
				break
			}
		}
	}(recall)

	recall.sql = fmt.Sprintf(hologres_vector_sql_v2, recall.vectorKeyField, recall.vectorEmbeddingField, recall.table, recall.where, recall.recallCount)
	return recall
}

func (r *HologresVectorRecallV2) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()

	var userEmbedding string
	userEmbKey := r.cachePrefix + string(user.Id)
	if value, ok := r.userVectorCache.GetIfPresent(userEmbKey); ok {
		userEmbedding = value.(string)
		user.AddProperty(r.modelName+"_embedding", userEmbedding)
	} else {
		emb, err := r.dao.VectorString(string(user.Id))
		if err != nil {
			if !errors.Is(err, module.VectoryEmptyError) {
				context.LogError(fmt.Sprintf("get user vector failed. %s, err=%v", r.modelName, err))
			}
		} else if emb != "" {
			user.AddProperty(r.modelName+"_embedding", emb)
			r.userVectorCache.Put(userEmbKey, emb)
			userEmbedding = emb
		}
	}

	ret = make([]*module.Item, 0, r.recallCount)
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
			context.LogInfo(fmt.Sprintf("requestId=%s\tmodule=HologresVectorRecallV2\tname=%s\thit cache\tcount=%d\tcost=%d",
				context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
			return
		}
	}

	if userEmbedding == "" {
		return
	}

	if r.dbStmt == nil {
		r.mu.Lock()
		if r.dbStmt == nil {
			r.sql = fmt.Sprintf(hologres_vector_sql_v2, r.vectorKeyField, r.vectorEmbeddingField, r.table, r.where, r.recallCount)
			if context.Debug {
				context.LogInfo("module=HologresVectorRecallV2\tsql=" + r.sql)
			}
			stmt, err := r.db.Prepare(r.sql)
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=HologresVectorRecallV2\tname=%s\terr=%v", context.RecommendId, r.modelName, err))
				r.mu.Unlock()
				return
			}
			r.dbStmt = stmt
			r.mu.Unlock()
		} else {
			r.mu.Unlock()
		}
	}

	rows, err := r.dbStmt.Query(userEmbedding)
	if err != nil {
		emb := userEmbedding
		if len(userEmbedding) > 500 {
			emb = userEmbedding[:500]
		}
		log.Error(fmt.Sprintf("requestId=%s\tmodule=HologresVectorRecallV2\tname=%s\tsql=%s\tuser_embedding=%s\terr=%v", context.RecommendId, r.modelName, r.sql, emb, err))
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
			if err2 := r.cache.Put(key, itemIds, time.Duration(r.cacheTime)*time.Second); err2 != nil {
				context.LogError(fmt.Sprintf("module=HologresVectorRecall\terror=%v", err2))
			}
		}()
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=HologresVectorRecallV2\tname=%s\tcount=%d\tcost=%d",
		context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return
}
