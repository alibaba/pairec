package sort

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/alibaba/pairec/v2/abtest"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/goburrow/cache"
	"github.com/huandu/go-sqlbuilder"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
	"math"
	"math/rand"
	gosort "sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SSDSort struct {
	db                   *sql.DB
	tableName            string
	suffixParam          string
	keyField             string
	embeddingField       string
	embSeparator         string
	gamma                float64
	useSSDStar           bool
	dbStmt               *sql.Stmt
	mu                   sync.RWMutex
	embCache             cache.Cache
	lastTableSuffixParam string
	normalizeEmb         bool
	windowSize           int
	abortRunCnt          int
	candidateCnt         int
	minScorePercent      float64
	embMissThreshold     float64
	filterRetrieveIds    []string
	ensurePosSimilarity  bool
	condition            *BoostScoreCondition
}

func NewSSDSort(config recconf.SSDSortConfig) *SSDSort {
	hologres, err := holo.GetPostgres(config.DaoConf.HologresName)
	if err != nil {
		panic(err)
	}
	cacheTime := time.Duration(360)
	if config.CacheTimeInMinutes > 0 {
		cacheTime = time.Duration(config.CacheTimeInMinutes)
	}
	ssd := SSDSort{
		db:                   hologres.DB,
		tableName:            config.TableName,
		suffixParam:          config.TableSuffixParam,
		keyField:             config.TablePKey,
		embeddingField:       config.EmbeddingColumn,
		embSeparator:         config.EmbeddingSeparator,
		gamma:                0.25,
		useSSDStar:           config.UseSSDStar,
		embCache:             cache.New(cache.WithMaximumSize(500000), cache.WithExpireAfterAccess(cacheTime*time.Minute)),
		lastTableSuffixParam: "",
		normalizeEmb:         true,
		windowSize:           config.WindowSize,
		abortRunCnt:          config.AbortRunCount,
		candidateCnt:         config.CandidateCount,
		minScorePercent:      config.MinScorePercent,
		embMissThreshold:     0.5,
		filterRetrieveIds:    config.FilterRetrieveIds,
		ensurePosSimilarity:  true,
	}
	if config.Gamma > 0 {
		ssd.gamma = config.Gamma
	}
	if ssd.windowSize <= 0 {
		ssd.windowSize = 5
	}
	if ssd.embSeparator == "" {
		ssd.embSeparator = ","
	}
	if strings.ToLower(config.NormalizeEmb) == "false" {
		ssd.normalizeEmb = false
	}
	if strings.ToLower(config.EnsurePositiveSim) == "false" {
		ssd.ensurePosSimilarity = false
	}
	if config.EmbMissedThreshold > 0 {
		ssd.embMissThreshold = config.EmbMissedThreshold
	}
	if config.Condition != nil {
		condition, err := NewBoostScoreCondition(config.Condition)
		if err != nil {
			log.Error(fmt.Sprintf("SSD Sort BoostScoreCondition error:%v", err))
		} else {
			ssd.condition = condition
		}
	}
	return &ssd
}

func (s *SSDSort) Sort(sortData *SortData) error {
	candidates, ok := sortData.Data.([]*module.Item)
	if !ok {
		return errors.New("sort data type error")
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx := sortData.Context
	if s.condition != nil {
		userProperties := sortData.User.MakeUserFeatures2()
		itemProperties := make(map[string]interface{})
		if flag, err := s.condition.filterParam.EvaluateByDomain(userProperties, itemProperties); err == nil && !flag {
			gosort.Sort(gosort.Reverse(ItemScoreSlice(candidates)))
			sortData.Data = candidates
			ctx.LogInfo("module=SSDSort\tcondition eval failed, skip")
			return nil
		}
	}
	if s.abortRunCnt > 0 && len(candidates) <= s.abortRunCnt {
		gosort.Sort(gosort.Reverse(ItemScoreSlice(candidates)))
		sortData.Data = candidates
		ctx.LogInfo(fmt.Sprintf("module=SSDSort\tcandidate cnt=%d, abort run cnt=%d", len(candidates), s.abortRunCnt))
		return nil
	}

	params := ctx.ExperimentResult.GetExperimentParams()
	names := params.Get("ssd_filter_retrieve_ids", nil)
	filterRetrieveIds := make([]string, 0)
	if names != nil {
		if values, ok := names.([]interface{}); ok {
			for _, v := range values {
				if name, okay := v.(string); okay {
					filterRetrieveIds = append(filterRetrieveIds, name)
				}
			}
		}
	}
	if len(filterRetrieveIds) == 0 {
		filterRetrieveIds = s.filterRetrieveIds
	} else {
		ctx.LogInfo(fmt.Sprintf("[ssd] filter retrieve ids = %v", filterRetrieveIds))
	}

	start := time.Now()
	var result []*module.Item
	if filterRetrieveIds != nil && len(filterRetrieveIds) > 0 {
		backup := make([]*module.Item, 0)
		selected := make([]*module.Item, 0, len(candidates))
		for _, item := range candidates {
			if utils.IndexOf(filterRetrieveIds, item.RetrieveId) >= 0 {
				backup = append(backup, item)
			} else {
				selected = append(selected, item)
			}
		}
		result = s.doSort(selected, ctx)
		if len(backup) > 0 {
			result = append(result, backup...)
		}
	} else {
		result = s.doSort(candidates, ctx)
	}

	sortData.Data = result
	ctx.LogInfo(fmt.Sprintf("module=SSDSort\tcount=%d\tcost_time=%d",
		len(result), utils.CostTime(start)))
	return nil
}

func (s *SSDSort) loadEmbeddingCache(ctx *context.RecommendContext, items []*module.Item) error {
	client := abtest.GetExperimentClient()
	tableSuffix := ""
	if s.suffixParam != "" && client != nil {
		tableSuffix = client.GetSceneParams("pairec").GetString(s.suffixParam, "")
	}
	if tableSuffix != s.lastTableSuffixParam {
		s.mu.Lock()
		if tableSuffix != s.lastTableSuffixParam {
			s.embCache.InvalidateAll()
			s.lastTableSuffixParam = tableSuffix
		}
		s.mu.Unlock()
	}

	absentItemIds := make([]string, 0)
	embedSize := 0
	lenAbsentItems := 0
	itemMap := make(map[string]*module.Item)
	for _, item := range items {
		if embI, ok := s.embCache.GetIfPresent(string(item.Id)); !ok {
			absentItemIds = append(absentItemIds, string(item.Id))
			itemMap[string(item.Id)] = item
		} else {
			item.Embedding = embI.([]float64)
			if embedSize == 0 {
				embedSize = len(item.Embedding)
			}
		}
	}
	if len(absentItemIds) > 0 {
		triggerItemIds := make([]interface{}, 0, len(absentItemIds))
		for _, itemId := range absentItemIds {
			triggerItemIds = append(triggerItemIds, interface{}(itemId))
		}

		table := s.tableName + tableSuffix
		builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
		builder.Select(s.keyField, s.embeddingField)
		builder.From(table)
		builder.Where(builder.In(s.keyField, triggerItemIds...))

		sqlQuery, args := builder.Build()
		ctx.LogDebug("module=SSDSort\tsqlquery=" + sqlQuery)
		rows, err := s.db.Query(sqlQuery, args...)
		if err != nil {
			ctx.LogError(fmt.Sprintf("module=SSDSort\terror=%v", err))
			return err
		}
		defer rows.Close()
		rowNum := 0
		itemID := &sql.NullString{}
		itemEmb := &sql.NullString{}
		for rows.Next() {
			if err := rows.Scan(itemID, itemEmb); err != nil {
				ctx.LogError(fmt.Sprintf("module=Scan SSDSort\terror=%v\tProductID=%s",
					err, itemID.String))
				continue
			}
			elements := strings.Split(strings.Trim(itemEmb.String, "{}"), s.embSeparator)
			embedSize = len(elements)
			vector := make([]float64, len(elements), len(elements)+1)
			for i, e := range elements {
				if val, err := strconv.ParseFloat(e, 64); err != nil {
					ctx.LogError(fmt.Sprintf("parse embedding value failed\terr=%v", err))
				} else {
					vector[i] = val
				}
			}
			if s.normalizeEmb {
				normV := floats.Norm(vector, 2)
				floats.Scale(1/normV, vector)
			}
			if s.ensurePosSimilarity {
				vector = append(vector, 1)
			}
			s.embCache.Put(itemID.String, vector)
			if item, ok := itemMap[itemID.String]; ok {
				item.Embedding = vector
			} else {
				return errors.New("item id is not in map")
			}
			rowNum = rowNum + 1
		}
		lenAbsentItems = len(absentItemIds) - rowNum
		if (float64(lenAbsentItems) / float64(len(items))) > s.embMissThreshold {
			return errors.New("the number of items missing embedding is above threshold")
		}
		if lenAbsentItems > 0 {
			for id, item := range itemMap {
				if len(item.Embedding) == 0 {
					ctx.LogWarning(fmt.Sprintf("not find embedding of item id:%s", id))
					item.Embedding = make([]float64, 0, embedSize+1)
					for i := 0; i < embedSize; i++ {
						item.Embedding = append(item.Embedding, rand.NormFloat64())
					}
					normV := floats.Norm(item.Embedding, 2)
					floats.Scale(1/normV, item.Embedding)
					if s.ensurePosSimilarity {
						item.Embedding = append(item.Embedding, 1)
					}
				}
			}
		}
	}
	if ctx.Debug {
		ctx.LogDebug(fmt.Sprintf("ctx_size=%d\tlen_items=%d\tlen_absent_items=%d\tlen_emb=%d",
			ctx.Size, len(items), lenAbsentItems, embedSize))
	}
	return nil
}

func (s *SSDSort) doSort(items []*module.Item, ctx *context.RecommendContext) []*module.Item {
	if len(items) == 0 {
		return items
	}
	params := ctx.ExperimentResult.GetExperimentParams()
	gamma := params.GetFloat("ssd_gamma", s.gamma)
	if gamma == 0 {
		gosort.Sort(gosort.Reverse(ItemScoreSlice(items)))
		ctx.LogDebug("ssd gamma=0, skip")
		return items
	}
	candidateCnt := params.GetInt("ssd_candidate_count", s.candidateCnt)
	minScorePercent := params.GetFloat("ssd_min_score_percent", s.minScorePercent)

	if (candidateCnt > 0 || minScorePercent > 0) && len(items) > ctx.Size {
		gosort.Sort(gosort.Reverse(ItemScoreSlice(items)))
		if candidateCnt > 0 {
			cnt := utils.MaxInt(ctx.Size, candidateCnt)
			if cnt < len(items) {
				items = items[:cnt]
			}
		}
		if minScorePercent > 0 && len(items) > ctx.Size {
			idx := ctx.Size
			maxScore := items[0].Score
			for ; idx < len(items); idx++ {
				percent := items[idx].Score / maxScore
				if percent < minScorePercent {
					break
				}
			}
			items = items[:idx]
		}
		ctx.LogInfo(fmt.Sprintf("module=SSDSort\tcandidate count=%d", len(items)))
	}

	if len(s.tableName) > 0 {
		if err := s.loadEmbeddingCache(ctx, items); err != nil {
			ctx.LogError(fmt.Sprintf("load embedding table cache failed %v", err))
			return items
		}
		return s.SSDWithSlidingWindow(items, ctx)
	} else {
		ctx.LogWarning("no embedding table and hooks")
	}
	return items
}

// SSDWithSlidingWindow paper: https://arxiv.org/pdf/2107.05204
func (s *SSDSort) SSDWithSlidingWindow(items []*module.Item, ctx *context.RecommendContext) []*module.Item {
	params := ctx.ExperimentResult.GetExperimentParams()
	gamma := params.GetFloat("ssd_gamma", s.gamma)
	windowSize := params.GetInt("ssd_window_size", s.windowSize)
	if windowSize <= 1 {
		ctx.LogWarning("SSD sliding window size must > 1, set to 5")
		windowSize = 5
	}
	N := len(items)
	// ensure all relevance score are positive and not in a large range
	relevanceScore := make([]float64, N)
	for i, item := range items {
		relevanceScore[i] = item.Score
	}
	doNorm := params.GetInt("ssd_norm_quality_score", 0)
	if doNorm == 1 {
		mean, variance := stat.PopMeanVariance(relevanceScore, nil)
		std := math.Sqrt(variance)
		for i, x := range relevanceScore {
			relevanceScore[i] = stat.StdScore(x, mean, std)
			items[i].AddAlgoScore("ssd_quality_score", relevanceScore[i])
		}
	} else if doNorm == 2 {
		maxScore := relevanceScore[0]
		minScore := relevanceScore[len(items)-1]
		scoreSpan := maxScore - minScore
		epsilon := 1e-6
		for i, x := range relevanceScore {
			relevanceScore[i] = ((x-minScore)/scoreSpan)*(1-epsilon) + epsilon
			items[i].AddAlgoScore("ssd_quality_score", relevanceScore[i])
		}
	}

	t := 1
	idx := floats.MaxIdx(relevanceScore)
	T := utils.MinInt(N, ctx.Size)
	dim := len(items[idx].Embedding)
	selected := make(map[int]bool, T)
	selected[idx] = true
	indices := make([]int, 0, T)
	indices = append(indices, idx)
	volume := gamma
	if !s.useSSDStar {
		volume *= floats.Norm(items[idx].Embedding, 2)
	}
	B := utils.NewCycleQueue(windowSize)
	P := utils.NewCycleQueue(windowSize)
	for t < T {
		if t > windowSize {
			i := B.Pop().(int)
			embI := mat.NewVecDense(dim, items[i].Embedding)
			projections := P.Pop().([]float64)
			for j := 0; j < N; j++ {
				if _, ok := selected[j]; ok {
					continue
				}
				scaledEmbI := mat.NewVecDense(dim, nil)
				scaledEmbI.ScaleVec(projections[j], embI)
				floats.Add(items[j].Embedding, scaledEmbI.RawVector().Data)
			}
		}
		if !B.Push(idx) {
			ctx.LogError(fmt.Sprintf("module=SSDSort\tpush index %d into queue failed", idx))
		} else {
			ctx.LogDebug(fmt.Sprintf("module=SSDSort\tpush index %d into queue", idx))
		}
		projections := make([]float64, N)
		embI := mat.NewVecDense(dim, items[idx].Embedding)
		for j := 0; j < N; j++ {
			if _, ok := selected[j]; ok {
				continue
			}
			projections[j] = floats.Dot(items[j].Embedding, items[idx].Embedding)
			projections[j] /= floats.Dot(items[idx].Embedding, items[idx].Embedding)
			scaledEmbI := mat.NewVecDense(dim, nil)
			scaledEmbI.ScaleVec(projections[j], embI)
			floats.Sub(items[j].Embedding, scaledEmbI.RawVector().Data)
		}
		if !P.Push(projections) {
			ctx.LogError(fmt.Sprintf("module=SSDSort\tpush projections %d into queue failed", idx))
		}
		t++
		qualities := make([]float64, len(relevanceScore))
		for i, r := range relevanceScore {
			if _, ok := selected[i]; ok {
				qualities[i] = -math.MaxFloat64
			} else {
				qualities[i] = r + volume*floats.Norm(items[i].Embedding, 2)
			}
		}
		idx = floats.MaxIdx(qualities)
		selected[idx] = true
		indices = append(indices, idx)
		if !s.useSSDStar {
			volume *= floats.Norm(items[idx].Embedding, 2)
		}
	}
	result := make([]*module.Item, 0, T)
	for _, index := range indices {
		result = append(result, items[index])
	}
	return result
}
