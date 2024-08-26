package sort

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
	gosort "sort"

	"github.com/goburrow/cache"
	"github.com/huandu/go-sqlbuilder"
	"github.com/alibaba/pairec/abtest"
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/persist/holo"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/utils"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
)

type DPPSort struct {
	db                   *sql.DB
	tableName            string
	suffixParam          string
	keyField             string
	embeddingField       string
	embSeparator         string
	alpha                float64
	dbStmt               *sql.Stmt
	mu                   sync.RWMutex
	embCache             cache.Cache
	lastTableSuffixParam string
	embeddingHookNames   []string
	normalizeEmb         bool
	windowSize           int
	abortRunCnt          int
	candidateCnt         int
	minScorePercent      float64
	embMissThreshold     float64
	filterRetrieveIds    []string
	ensurePosSimilarity  bool
}

type EmbeddingHookFunc func(context *context.RecommendContext, item *module.Item) []float64

var embeddingHooks = make(map[string]EmbeddingHookFunc)

func RegisterEmbeddingHook(name string, fn EmbeddingHookFunc) {
	embeddingHooks[name] = fn
}

func NewDPPSort(config recconf.DPPSortConfig) *DPPSort {
	hologres, err := holo.GetPostgres(config.DaoConf.HologresName)
	if err != nil {
		panic(err)
	}
	cacheTime := time.Duration(360)
	if config.CacheTimeInMinutes > 0 {
		cacheTime = time.Duration(config.CacheTimeInMinutes)
	}
	dpp := DPPSort{
		db:                   hologres.DB,
		tableName:            config.TableName,
		suffixParam:          config.TableSuffixParam,
		keyField:             config.TablePKey,
		embeddingField:       config.EmbeddingColumn,
		embSeparator:         config.EmbeddingSeparator,
		alpha:                config.Alpha,
		embCache:             cache.New(cache.WithMaximumSize(500000), cache.WithExpireAfterAccess(cacheTime*time.Minute)),
		lastTableSuffixParam: "",
		embeddingHookNames:   config.EmbeddingHookNames,
		normalizeEmb:         true,
		windowSize:           config.WindowSize,
		abortRunCnt:          config.AbortRunCount,
		candidateCnt:         config.CandidateCount,
		minScorePercent:      config.MinScorePercent,
		embMissThreshold:     0.5,
		filterRetrieveIds:    config.FilterRetrieveIds,
		ensurePosSimilarity:  true,
	}
	if dpp.windowSize <= 0 {
		dpp.windowSize = 10
	}
	if dpp.embSeparator == "" {
		dpp.embSeparator = ","
	}
	if strings.ToLower(config.NormalizeEmb) == "false" {
		dpp.normalizeEmb = false
	}
	if strings.ToLower(config.EnsurePositiveSim) == "false" {
		dpp.ensurePosSimilarity = false
	}
	if config.EmbMissedThreshold > 0 {
		dpp.embMissThreshold = config.EmbMissedThreshold
	}

	return &dpp
}

func (s *DPPSort) Sort(sortData *SortData) error {
	candidates, ok := sortData.Data.([]*module.Item)
	if !ok {
		return errors.New("sort data type error")
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx := sortData.Context
	if s.abortRunCnt > 0 && len(candidates) <= s.abortRunCnt {
		ctx.LogInfo(fmt.Sprintf("candidate cnt=%d, abort run cnt=%d", len(candidates), s.abortRunCnt))
		return nil
	}
	
	params := ctx.ExperimentResult.GetExperimentParams()
	names := params.Get("dpp_filter_retrieve_ids", nil)
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
		ctx.LogInfo(fmt.Sprintf("[dpp] filter retrieve ids = %v", filterRetrieveIds))
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
	ctx.LogInfo(fmt.Sprintf("module=DPPSort\tcount=%d\tcost_time=%d",
		len(result), utils.CostTime(start)))
	return nil
}

func (s *DPPSort) loadEmbeddingCache(ctx *context.RecommendContext, items []*module.Item) (int, error) {
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
		ctx.LogDebug("module=DPPSort\tsqlquery=" + sqlQuery)
		rows, err := s.db.Query(sqlQuery, args...)
		if err != nil {
			ctx.LogError(fmt.Sprintf("module=DPPSort\terror=%v", err))
			return -1, err
		}
		defer rows.Close()
		rowNum := 0
		itemID := &sql.NullString{}
		itemEmb := &sql.NullString{}
		for rows.Next() {
			if err := rows.Scan(itemID, itemEmb); err != nil {
				ctx.LogError(fmt.Sprintf("module=Scan DPPSort\terror=%v\tProductID=%s",
					err, itemID.String))
				continue
			}
			elements := strings.Split(strings.Trim(itemEmb.String, "{}"), s.embSeparator)
			embedSize = len(elements)
			vector := make([]float64, len(elements), len(elements)+1)
			for i, e := range elements {
				if val, err := strconv.ParseFloat(e, 64); err != nil {
					log.Error(fmt.Sprintf("parse embedding value failed\terr=%v", err))
				} else {
					vector[i] = val
				}
			}
			if s.normalizeEmb {
				normV := floats.Norm(vector, 2)
				floats.Scale(1/normV, vector)
			}
			s.embCache.Put(itemID.String, vector)
			if item, ok := itemMap[itemID.String]; ok {
				item.Embedding = vector
			} else {
				return -1, errors.New("item id is not in map")
			}
			rowNum = rowNum + 1
		}
		lenAbsentItems = len(absentItemIds) - rowNum
		if (float64(lenAbsentItems) / float64(len(items))) > s.embMissThreshold {
			return -1, errors.New("the number of items missing embedding is above threshold")
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
				}
			}
		}
	}
	if ctx.Debug {
		ctx.LogDebug(fmt.Sprintf("ctx_size=%d \tlen_items=%d \tlen_absent_items=%d \tlen_emb=%d",
			ctx.Size, len(items), lenAbsentItems, embedSize))
	}
	return embedSize, nil
}

func (s *DPPSort) doSort(items []*module.Item, ctx *context.RecommendContext) []*module.Item {
	if len(items) == 0 {
		return make([]*module.Item, 0)
	}
	params := ctx.ExperimentResult.GetExperimentParams()
	windowSize := params.GetInt("dpp_window_size", s.windowSize)
	candidateCnt := params.GetInt("dpp_candidate_count", s.candidateCnt)
	minScorePercent := params.GetFloat("dpp_min_score_percent", s.minScorePercent)

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
		ctx.LogInfo(fmt.Sprintf("module=DPPSort\tcandidate count=%d", len(items)))
	}

	if len(s.tableName) > 0 {
		lenEmb, err := s.loadEmbeddingCache(ctx, items)
		if err != nil {
			ctx.LogError(fmt.Sprintf("load embedding table cache failed %v", err))
			return items
		}
		hookEmb := s.GenerateEmbedding(ctx, items[0])
		if ctx.Debug {
			if len(hookEmb) != 0 {
				ctx.LogInfo(fmt.Sprintf("find embedding table: %s, find HooksEmb, len(hookEmb)=%d", s.tableName, len(hookEmb)))
			} else {
				ctx.LogInfo(fmt.Sprintf("find embedding table: %s, not find HooksEmb", s.tableName))
			}
		}
		kernelMatrix, err := s.KernelMatrix(ctx, items, lenEmb+len(hookEmb), true)
		if err != nil {
			ctx.LogError(fmt.Sprintf("build kernel matrix failed %v", err))
			return items
		}
		slice := DPPWithWindow(kernelMatrix, ctx.Size, windowSize)
		if ctx.Debug {
			ctx.LogDebug(fmt.Sprintf("the length of dpp-return items is %d and the ctx.size is %d, window=%d",
				len(slice), ctx.Size, windowSize))
		}
		retItems := make([]*module.Item, 0, ctx.Size)
		for _, i := range slice {
			retItems = append(retItems, items[i])
		}
		return retItems
	} else if s.hasHookFunc() {
		ctx.LogDebug("not find embedding table, find HooksEmb")
		hookEmb := s.GenerateEmbedding(ctx, items[0])
		ctx.LogDebug(fmt.Sprintf("first dpp hook emb: %v", hookEmb))
		kernelMatrix, err := s.KernelMatrix(ctx, items, len(hookEmb), false)
		if err != nil {
			ctx.LogError(fmt.Sprintf("build kernel matrix failed %v", err))
			return items
		}
		slice := DPPWithWindow(kernelMatrix, ctx.Size, windowSize)

		retItems := make([]*module.Item, 0, ctx.Size)
		for _, i := range slice {
			retItems = append(retItems, items[i])
		}
		return retItems
	} else {
		ctx.LogWarning("no embedding table and hooks")
	}
	return items
}

func (s *DPPSort) hasHookFunc() bool {
	for _, name := range s.embeddingHookNames {
		if _, ok := embeddingHooks[name]; ok {
			return true
		}
	}
	return false
}

func (s *DPPSort) GenerateEmbedding(context *context.RecommendContext, item *module.Item) []float64 {
	ret := make([]float64, 0)
	for _, name := range s.embeddingHookNames {
		if fn, ok := embeddingHooks[name]; ok {
			ret = append(ret, fn(context, item)...)
		}
	}
	return ret
}

func (s *DPPSort) KernelMatrix(context *context.RecommendContext, items []*module.Item, lenEmb int, hasTable bool) (*mat.Dense, error) {
	params := context.ExperimentResult.GetExperimentParams()
	alpha := params.GetFloat("dpp_alpha", s.alpha)
	context.LogDebug(fmt.Sprintf("dpp alpha: %f", alpha))

	// ensure all relevance score are positive and not in a large range
	relevanceScore := make([]float64, len(items))
	for i, item := range items {
		relevanceScore[i] = item.Score
	}
	doNorm := params.GetInt("dpp_norm_relevance_score", 0)
	if doNorm == 1 {
		mean, variance := stat.PopMeanVariance(relevanceScore, nil)
		std := math.Sqrt(variance)
		for i, x := range relevanceScore {
			relevanceScore[i] = stat.StdScore(x, mean, std)
		}
	} else if doNorm == 2 {
		maxScore := relevanceScore[0]
		minScore := relevanceScore[len(items) - 1]
		scoreSpan := maxScore - minScore
		epsilon := 1e-6
		for i, x := range relevanceScore {
			relevanceScore[i] = ((x - minScore) / scoreSpan) * (1 - epsilon) + epsilon
		}
	}

	itemSize := len(items)
	rawScores := make([]float64, 0, itemSize)
	featureMat := mat.NewDense(itemSize, lenEmb+1, nil)
	for row, item := range items {
		item.AddAlgoScore("dpp_relevance_score", relevanceScore[row])
		embs := s.GenerateEmbedding(context, item)
		if hasTable {
			itemEmb := item.Embedding
			if len(itemEmb) == 0 {
				return nil, errors.New("embedding missed during constructing KernelMatrix")
			}
			if len(embs) != 0 {
				embs = append(embs, itemEmb...)
				norm := floats.Norm(embs, 2)
				floats.Scale(1/norm, embs)
			} else {
				embs = itemEmb
			}
			if len(embs) != lenEmb {
				return nil, errors.New("the length of user-defined function is not equal")
			}
			embs = append(embs, 1)
			floats.Scale(1/math.Sqrt2, embs)
			featureMat.SetRow(row, embs)
			rawScores = append(rawScores, math.Exp(alpha*relevanceScore[row]))
		} else if len(embs) > 0 {
			if len(embs) != lenEmb {
				return nil, errors.New("the length of user-defined function is not equal")
			}
			if s.normalizeEmb {
				norm := floats.Norm(embs, 2)
				floats.Scale(1/norm, embs)
			}
			if s.ensurePosSimilarity {
				embs = append(embs, 1)
				floats.Scale(1/math.Sqrt2, embs)
			} else {
				embs = append(embs, 0)
			}
			featureMat.SetRow(row, embs)
			rawScores = append(rawScores, math.Exp(alpha*relevanceScore[row]))
		} else {
			context.LogWarning(fmt.Sprintf("not find embedding of item id:%s, generate random embedding", item.Id))
			embedding := make([]float64, 0, lenEmb+1)
			for i := 0; i < lenEmb; i++ {
				embedding = append(embedding, rand.NormFloat64())
			}
			normV := floats.Norm(embedding, 2)
			floats.Scale(1/normV, embedding)
			embedding = append(embedding, 1)
			floats.Scale(1/math.Sqrt2, embedding)
			featureMat.SetRow(row, embedding)
			rawScores = append(rawScores, math.Exp(alpha*relevanceScore[row]))
		}
	}

	var similarities mat.Dense
	similarities.Mul(featureMat, featureMat.T())

	ruDiag := mat.NewDense(itemSize, itemSize, nil)
	for i, v := range rawScores {
		ruDiag.Set(i, i, v)
	}
	var kernelMat mat.Dense
	kernelMat.Mul(ruDiag, &similarities)
	kernelMat.Mul(&kernelMat, ruDiag)

	return &kernelMat, nil
}

func DPPWithWindow(L *mat.Dense, topN int, windowSize int) []int {
	result := make([]int, 0, topN)
	if topN <= windowSize {
		return DPP(L, topN, result)
	}
	for i := 0; i < topN/windowSize; i++ {
		subResult := DPP(L, windowSize, result)
		result = append(result, subResult...)
	}
	if topN%windowSize > 0 {
		subResult := DPP(L, topN%windowSize, result)
		result = append(result, subResult...)
	}
	return result
}

func DPP(L *mat.Dense, topN int, existed []int) []int {
	epsilon := 1e-10
	N, _ := L.Dims()
	if topN > N {
		topN = N
	}
	Y := make([]int, 0, topN)
	_d2 := make([]float64, N)
	for i := 0; i < N; i++ {
		if indexOf(existed, i) < 0 {
			_d2[i] = L.At(i, i)
		} else {
			_d2[i] = math.NaN()
		}
	}
	j := floats.MaxIdx(_d2)
	Y = append(Y, j)

	d2 := mat.NewVecDense(N, _d2)
	c := mat.NewDense(topN, N, nil)
	ss := mat.NewDense(1, N, nil)
	e := mat.NewDense(1, N, nil)
	for len(Y) < topN {
		dj := d2.AtVec(j)
		if dj < epsilon {
			break
		}
		dj = math.Sqrt(dj)
		k := len(Y) - 1
		Lj := L.Slice(j, j+1, 0, N)
		if k == 0 {
			e.Scale(1/dj, Lj)
		} else {
			cj := c.Slice(0, k, j, j+1)
			ss.Mul(cj.T(), c.Slice(0, k, 0, N)) // s = <cj, ci>
			e.Sub(Lj, ss)                       // e = Lj - s
			e.Scale(1/dj, e)                    // e = e / dj
		}
		c.SetRow(k, e.RawRowView(0)) // c = [c, e]
		e.MulElem(e, e)              // e = e^2
		d2.SubVec(d2, e.RowView(0))  // d2 = d2 - e2

		d2.SetVec(j, math.NaN()) // j has already be selected, set to NaN
		j = floats.MaxIdx(d2.RawVector().Data)
		Y = append(Y, j)
	}
	if len(Y) < topN {
		for i := 0; i < N; i++ {
			if indexOf(existed, i) < 0 && indexOf(Y, i) < 0 {
				Y = append(Y, i)
				if len(Y) == topN {
					break
				}
			}
		}
	}

	return Y
}

func indexOf(a []int, e int) int {
	n := len(a)
	var i = 0
	for ; i < n; i++ {
		if e == a[i] {
			return i
		}
	}
	return -1
}
