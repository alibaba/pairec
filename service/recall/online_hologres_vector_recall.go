package recall

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goburrow/cache"

	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/algorithm/eas"
	"github.com/alibaba/pairec/v2/algorithm/response"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/feature"
	"github.com/alibaba/pairec/v2/service/rank"
	"github.com/alibaba/pairec/v2/utils"
)

var (
	online_hologres_vector_sql = "SELECT %s, pm_approx_inner_product_distance(%s,$1) as distance FROM %s %s ORDER BY distance desc limit $2"
)

type OnlineHologresVectorRecall struct {
	*BaseRecall
	db                   *sql.DB
	table                string
	vectorEmbeddingField string
	vectorKeyField       string
	where                string
	sql                  string
	features             []*feature.Feature
	recallAlgoType       string
	mu                   sync.RWMutex
	dbStmt               *sql.Stmt
	userVectorCache      cache.Cache
	timeInterval         int
}

func NewOnlineHologresVectorRecall(config recconf.RecallConfig) *OnlineHologresVectorRecall {
	hologres, err := holo.GetPostgres(config.HologresVectorConf.HologresName)
	if err != nil {
		panic(err)
	}
	recall := &OnlineHologresVectorRecall{
		BaseRecall:           NewBaseRecall(config),
		db:                   hologres.DB,
		table:                config.HologresVectorConf.VectorTable,
		vectorEmbeddingField: config.HologresVectorConf.VectorEmbeddingField,
		vectorKeyField:       config.HologresVectorConf.VectorKeyField,
		where:                config.HologresVectorConf.WhereClause,
		timeInterval:         config.HologresVectorConf.TimeInterval,
		recallAlgoType:       eas.Eas_Processor_EASYREC,
	}
	createTime := time.Now().Unix() - int64(recall.timeInterval)
	recall.where = strings.ReplaceAll(recall.where, "${time}", strconv.FormatInt(createTime, 10))
	if recall.where != "" {
		recall.where = "WHERE " + recall.where
	}

	if recall.cacheTime <= 0 && recall.cache != nil {
		recall.cacheTime = 1800
	}

	recall.userVectorCache = cache.New(
		cache.WithMaximumSize(10000),
		cache.WithExpireAfterAccess(time.Duration(recall.cacheTime+10)*time.Second),
	)

	recall.sql = fmt.Sprintf(online_hologres_vector_sql, recall.vectorKeyField, recall.vectorEmbeddingField, recall.table, recall.where)
	var features []*feature.Feature
	for _, conf := range config.UserFeatureConfs {
		f := feature.LoadWithConfig(conf)
		features = append(features, f)
	}

	recall.features = features
	return recall
}

func (r *OnlineHologresVectorRecall) loadUserFeatures(user *module.User, context *context.RecommendContext) {
	var wg sync.WaitGroup
	for _, fea := range r.features {
		wg.Add(1)
		go func(fea *feature.Feature) {
			defer wg.Done()
			fea.LoadFeatures(user, nil, context)
		}(fea)
	}

	wg.Wait()

}
func (r *OnlineHologresVectorRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()

	var userEmbedding string
	userEmbKey := r.cachePrefix + string(user.Id)
	if value, ok := r.userVectorCache.GetIfPresent(userEmbKey); ok {
		userEmbedding = value.(string)
		//user.AddProperty(r.modelName+"_embedding", userEmbedding)
	} else {
		// get user emb from eas model
		//first get user features
		r.loadUserFeatures(user, context)
		// second invoke eas model
		algoGenerator := rank.CreateAlgoDataGenerator(r.recallAlgoType, nil)
		algoGenerator.SetItemFeatures(nil)
		algoGenerator.AddFeatures(nil, nil, user.MakeUserFeatures2())
		algoData := algoGenerator.GeneratorAlgoData()
		algoRet, err := algorithm.Run(r.recallAlgo, algoData.GetFeatures())
		if err != nil {
			context.LogError(fmt.Sprintf("requestId=%s\tmodule=OnlineHologresVectorRecall\tname=%s\terr=%v", context.RecommendId, r.modelName, err))
		} else {
			// eas model invoke success
			if result, ok := algoRet.([]response.AlgoResponse); ok && len(result) > 0 {
				if userEmbResponse, ok := result[0].(*eas.EasyrecUserEmbResponse); ok {
					userEmbedding = userEmbResponse.GetUserEmb()
				}
			}
			//user.AddProperty(r.modelName+"_embedding", userEmbedding)
			r.userVectorCache.Put(userEmbKey, userEmbedding)
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
					item.Score = f
				} else {
					item = module.NewItem(id)
				}

				item.RetrieveId = r.modelName
				ret = append(ret, item)
			}
			context.LogInfo(fmt.Sprintf("module=OnlineHologresVectorRecall\tname=%s\thit cache\tcount=%d\tcost=%d",
				r.modelName, len(ret), utils.CostTime(start)))
			return
		}
	}

	if userEmbedding == "" {
		return
	}

	if r.dbStmt == nil {
		r.mu.Lock()
		if r.dbStmt == nil {
			if context.Debug {
				context.LogInfo("module=OnlineHologresVectorRecall\tsql=" + r.sql)
			}
			stmt, err := r.db.Prepare(r.sql)
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=OnlineHologresVectorRecall\tname=%s\terr=%v", context.RecommendId, r.modelName, err))
				r.mu.Unlock()
				return
			}
			r.dbStmt = stmt
			r.mu.Unlock()
		} else {
			r.mu.Unlock()
		}
	}

	userEmbeddingList := strings.Split(userEmbedding, "|")
	var wg sync.WaitGroup
	ch := make(chan []*module.Item, len(userEmbeddingList))
	for _, userEmb := range userEmbeddingList {
		wg.Add(1)
		go func(userEmb string) {
			defer wg.Done()
			items := make([]*module.Item, 0, r.recallCount/len(userEmbeddingList))
			rows, err := r.dbStmt.Query(fmt.Sprintf("{%s}", userEmb), r.recallCount/len(userEmbeddingList))
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=OnlineHologresVectorRecall\tname=%s\terr=%v", context.RecommendId, r.modelName, err))
				ch <- items
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
				item.Score = distance

				items = append(items, item)
			}
			ch <- items

		}(userEmb)
	}
	wg.Wait()
	for i := 0; i < len(userEmbeddingList); i++ {
		items := <-ch
		ret = append(ret, items...)
	}

	if r.cache != nil && len(ret) > 0 {
		go func() {
			key := r.cachePrefix + string(user.Id)
			var itemIds string
			for _, item := range ret {
				itemIds += fmt.Sprintf("%s::%v", string(item.Id), item.Score) + ","
			}
			itemIds = itemIds[:len(itemIds)-1]
			if err2 := r.cache.Put(key, itemIds, time.Duration(r.cacheTime)*time.Second); err2 != nil {
				context.LogError(fmt.Sprintf("module=OnlineHologresVectorRecall\terror=%v", err2))
			}
		}()
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=OnlineHologresVectorRecall\tname=%s\tcount=%d\tcost=%d",
		context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return
}
