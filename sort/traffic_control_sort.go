package sort

import (
	gocontext "context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alibaba/pairec/v2/abtest"
	"github.com/alibaba/pairec/v2/persist/holo"
	"hash/crc32"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/alibaba/pairec/v2/utils/sqlutil"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/experiments"
	"github.com/goburrow/cache"
	"github.com/huandu/go-sqlbuilder"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/sampleuv"
)

type TrafficControlSort struct {
	name               string
	config             *recconf.PIDControllerConfig
	exp2controllers    map[string]map[string]*PIDController
	controllerLock     sync.RWMutex
	db                 *sql.DB
	feaMu              sync.RWMutex
	feaStmtMap         map[string]map[int]*sql.Stmt
	itemCache          cache.Cache
	loadItemFeature    bool
	preloadItemFeature bool
	cloneInstances     map[string]*TrafficControlSort
	boostScoreSort     *BoostScoreSort
	hologresTableName  string
	primaryKey         string

	context *context.RecommendContext
}

var positionWeight []float64
var expTable []float64
var tanhTable []float64
var sigmoidTable []float64
var experimentClient *experiments.ExperimentClient

func init() {
	positionWeight = make([]float64, 500)
	for i := 0; i < 500; i++ {
		positionWeight[i] = math.Exp(-0.01 * float64(i))
	}

	expTable = make([]float64, 1000)
	for i := 0; i < 1000; i++ {
		expTable[i] = math.Exp(float64(i) / 1000.0)
	}

	tanhTable = make([]float64, 3000)
	for i := 0; i < 3000; i++ {
		tanhTable[i] = math.Tanh(float64(i) / 1000.0) // 值域范围 [0, 3)
	}

	sigmoidTable = make([]float64, 10000)
	for i := 0; i < 10000; i++ {
		x := float64(i)/1000.0 + 5.0 // 范围 [5, 15)
		sigmoidTable[i] = 1.0 / (1.0 + math.Exp(10-x))
	}
}

func NewTrafficControlSort(config recconf.SortConfig) *TrafficControlSort {
	experimentClient = abtest.GetExperimentClient()
	if experimentClient == nil {
		log.Error("module=TrafficControlSort\tGetExperimentClient failed.")
		return nil
	}
	conf := config.PIDConf

	var db *sql.DB = nil
	if conf.LoadItemFeature {
		hologres, err := holo.GetPostgres(conf.HologresName)
		if err != nil {
			log.Error(fmt.Sprintf("module=TrafficControlSort\tget hologres error=%v, name=%s", err, conf.HologresName))
			return nil
		}
		db = hologres.DB
	}

	maxCacheSize := 100000
	if conf.MaxItemCacheSize > 0 {
		maxCacheSize = conf.MaxItemCacheSize
	}
	minCacheTime := 3600
	if conf.MaxItemCacheTime > 0 {
		minCacheTime = conf.MaxItemCacheTime
	}
	trafficControlSort := TrafficControlSort{
		config:             &conf,
		exp2controllers:    make(map[string]map[string]*PIDController),
		db:                 db,
		loadItemFeature:    conf.LoadItemFeature,
		preloadItemFeature: conf.PreloadItemFeature,
		feaStmtMap:         make(map[string]map[int]*sql.Stmt),
		itemCache:          cache.New(cache.WithMaximumSize(maxCacheSize), cache.WithExpireAfterWrite(time.Duration(minCacheTime)*time.Second)),
		name:               config.Name,
		cloneInstances:     make(map[string]*TrafficControlSort),
		hologresTableName:  conf.HologresTableName,
		primaryKey:         conf.PrimaryKey,
	}

	if len(conf.BoostScoreConditions) > 0 {
		boostConf := recconf.SortConfig{
			Debug:                config.Debug,
			BoostScoreConditions: conf.BoostScoreConditions,
		}
		trafficControlSort.boostScoreSort = NewBoostScoreSort(boostConf)
	}

	go func() {
		for {
			// check for new control tasks and new control targets
			// trafficControlSort.loadTrafficControlTask("")
			for expId := range trafficControlSort.exp2controllers {
				trafficControlSort.loadTrafficControlTask(expId)
			}
			if trafficControlSort.preloadItemFeature {
				trafficControlSort.loadAllItemFeature()
			}
			time.Sleep(time.Minute) // 这里需要更新频繁一点，不然web页面上meta信息的修改不能及时反应出来
		}
	}()

	return &trafficControlSort
}

func (p *TrafficControlSort) Sort(sortData *SortData) error {
	items, good := sortData.Data.([]*module.Item)
	if !good {
		return errors.New("sort data type error")
	}
	if len(items) == 0 {
		return nil
	}

	start := time.Now()
	ctx := sortData.Context

	if p.context == nil {
		p.context = ctx
	}

	params := ctx.ExperimentResult.GetExperimentParams()

	if p.boostScoreSort != nil && params.Get("pid_boost_score", true).(bool) {
		err := p.boostScoreSort.Sort(sortData)
		if err != nil {
			ctx.LogError(fmt.Sprintf("module=TrafficControlSort\terror=%v", err))
		}
		ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\\BoostScoreSort\tcount=%d\tcost=%d", len(items), utils.CostTime(start)))
	}

	sort.Sort(sort.Reverse(ItemScoreSlice(items)))
	for i, item := range items {
		item.AddProperty("__flow_control_id__", i+1)
		item.AddProperty("_ORIGIN_POSITION_", i+1)
	}

	controllerMap := p.getPidControllers(ctx)
	wholeCtrls, singleCtrls := splitController(controllerMap, ctx)
	if len(wholeCtrls) == 0 && len(singleCtrls) == 0 {
		ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\tcount=%d\tNo traffic control task", len(items)))
		sortData.Data = items
		return nil
	}

	if enable := setHyperParams(controllerMap, ctx); !enable {
		ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\tcount=%d\ttraffic control task turn off", len(items)))
		sortData.Data = items
		return nil
	}

	wgLoad := sync.WaitGroup{}
	if p.loadItemFeature { // 如果需要，从holo加载item特征
		wgLoad.Add(1)
		go p.loadItemFeatures(items, ctx, &wgLoad)
	}

	wgCtrl := sync.WaitGroup{}
	if len(singleCtrls) > 0 {
		wgCtrl.Add(1)
		go microControl(singleCtrls, items, ctx, &wgLoad, &wgCtrl)
	}
	if len(wholeCtrls) > 0 {
		wgCtrl.Add(1)
		go macroControl(wholeCtrls, items, ctx, &wgLoad, &wgCtrl)
	}
	wgCtrl.Wait()

	pageNo := utils.ToInt(ctx.GetParameter("pageNum"), 1)
	pageSize := ctx.Size // utils.ToInt(ctx.GetParameter("pageSize"), 10)
	if pageNo < 1 {
		pageNo = 1
	}
	limitFirstPage := params.GetInt("limit_uplift_at_first_page", 0)
	for i, item := range items {
		finalDeltaRank := item.GetAlgoScore("__delta_rank__")
		if finalDeltaRank != 0.0 {
			rank := float64(i+1) - finalDeltaRank
			if pageNo <= 1 && limitFirstPage != 0 {
				if i < pageSize {
					item.AddProperty("_NEW_POSITION_", i+1)
				} else {
					if rank <= float64(pageSize) { // 保证第一页流量调控的结果仅作为打散的候补出现
						rank = float64(pageSize) + 1 + tanh(0.001*rank) // rank > pageSize
					}
					item.AddProperty("_NEW_POSITION_", rank)
				}
			} else {
				item.AddProperty("_NEW_POSITION_", rank)
			}
		} else {
			item.AddProperty("_NEW_POSITION_", i+1)
		}
	}
	sort.Sort(ItemRankSlice(items))
	ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\tcount=%d\tcost=%d", len(items), utils.CostTime(start)))
	sortData.Data = items
	return nil
}

func (p *TrafficControlSort) loadAllItemFeature() {
	fields := make([]string, 0)
	for _, controllerMap := range p.exp2controllers {
		for _, controller := range controllerMap {
			itemTableMeta := controller.task.ItemTableMeta
			for _, field := range itemTableMeta.Fields {
				fields = append(fields, field.Name)
			}

			itemFields := getFilterFields(controller.task.ItemConditionArray)
			fields = append(fields, itemFields...)

			itemFields = getFilterFields(controller.target.ItemConditionArray)
			fields = append(fields, itemFields...)
		}
	}
	p.loadAllItemFeatureFromHolo(p.hologresTableName, p.primaryKey, fields)
}

func (p *TrafficControlSort) loadAllItemFeatureFromHolo(table string, idField string, fields []string) {
	selectFields := utils.UniqueStrings(fields)
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select(selectFields...)
	sb.From(table)
	sqlQuery, args := sb.Build()
	p.context.LogDebug("module=TrafficControlSort\tfn=loadAllItemFeatureFromHolo\tquery:" + sqlQuery)

	sort.Strings(selectFields)
	cols := strings.Join(selectFields, ",")
	hash := crc32.ChecksumIEEE([]byte(cols))
	tableKey := fmt.Sprintf("%s_%d_%d", table, len(selectFields), hash)
	numItems := 0

	p.feaMu.RLock()
	feaStmtMap := p.feaStmtMap[tableKey]
	feaStmt := feaStmtMap[numItems]
	p.feaMu.RUnlock()

	if feaStmt == nil {
		p.feaMu.Lock()
		if feaStmtMap == nil {
			feaStmtMap = p.feaStmtMap[tableKey]
			if feaStmtMap == nil {
				feaStmtMap = make(map[int]*sql.Stmt)
				p.feaStmtMap[tableKey] = feaStmtMap
			}
		}
		feaStmt = feaStmtMap[numItems]
		if feaStmt == nil {
			if stmt, err := p.db.Prepare(sqlQuery); err != nil {
				p.context.LogError(fmt.Sprintf("model=TrafficControlSort\tenent=loadAllItemFeatureFromHolo\terr=%v", err))
				p.feaMu.Unlock()
				return
			} else {
				feaStmt = stmt
				feaStmtMap[numItems] = stmt
			}
		}
		p.feaMu.Unlock()
	}

	cntx, cancel := gocontext.WithTimeout(gocontext.Background(), 300*time.Millisecond)
	defer cancel()
	rows, err := feaStmt.QueryContext(cntx, args...)
	if err != nil {
		if errors.Is(err, gocontext.DeadlineExceeded) {
			p.context.LogWarning("module=TrafficControlSort\tevent=loadAllItemFeatureFromHolo\ttimeout=200")
			return
		}
		p.context.LogError(fmt.Sprintf("module=TrafficControlSort\terror=hologres error(%v)", err))
		return
	}
	defer rows.Close()

	columns, err := rows.ColumnTypes()
	if err != nil {
		p.context.LogError(fmt.Sprintf("module=TrafficControlSort\terror=hologres error(%v)", err))
		return
	}
	values := sqlutil.ColumnValues(columns)

	count := 0
	for rows.Next() {
		if err = rows.Scan(values...); err == nil {
			features := make(map[string]interface{})
			for i, column := range columns {
				name := column.Name()
				val := values[i]
				switch v := val.(type) {
				case *sql.NullString:
					if v.Valid {
						features[name] = v.String
					}
				case *sql.NullInt32:
					if v.Valid {
						features[name] = v.Int32
					}
				case *sql.NullInt64:
					if v.Valid {
						features[name] = v.Int64
					}
				case *sql.NullFloat64:
					if v.Valid {
						features[name] = v.Float64
					}
				case *sql.NullBool:
					if v.Valid {
						features[name] = v.Bool
					}
				case *sql.NullTime:
					if v.Valid {
						features[name] = v.Time
					}
				default:
				}
			}
			if id, ok := features[idField]; ok {
				delete(features, idField)
				itemId := utils.ToString(id, "")
				cacheKey := table + "_" + itemId
				p.itemCache.Put(cacheKey, features)
				count++
			} else {
				p.context.LogError(fmt.Sprintf("module=TrafficControlSort\tevent=loadAllItemFeatureFromHolo\tNo item id found"))
			}
		} else {
			p.context.LogError(fmt.Sprintf("module=TrafficControlSort\terror=hologres error(%v)", err))
		}
	}
	p.context.LogInfo(fmt.Sprintf("module=TrafficControlSort\tevent=loadAllItemFeatureFromHolo\tload %d item feature from table %s", count, table))
}

func (p *TrafficControlSort) loadItemFeatures(items []*module.Item, ctx *context.RecommendContext, wg *sync.WaitGroup) {
	defer wg.Done()
	begin := time.Now()
	defer func() {
		p.context.LogInfo(fmt.Sprintf("module=PID_loadItemFeature\tcount=%d\tcost=%d", len(items), utils.CostTime(begin)))
	}()
	if len(items) == 0 {
		return
	}
	controllerMap := p.getPidControllers(ctx)
	if controllerMap == nil || len(controllerMap) == 0 {
		return
	}

	fields := make([]string, 0)
	for _, controller := range controllerMap {

		itemFields := getFilterFields(controller.task.ItemConditionArray)
		fields = append(fields, itemFields...)

		itemFields = getFilterFields(controller.target.ItemConditionArray)
		fields = append(fields, itemFields...)
	}
	p.loadItemFeatureFromHolo(p.hologresTableName, p.primaryKey, fields, items, ctx)
}

func (p *TrafficControlSort) loadItemFeatureFromHolo(table string, idField string, fields []string, items []*module.Item, ctx *context.RecommendContext) {
	itemIds := make([]interface{}, 0, len(items))
	itemMap := make(map[string]*module.Item, len(items))
	count := 0
	for _, item := range items {
		cacheKey := table + "_" + string(item.Id)
		if attr, exist := p.itemCache.GetIfPresent(cacheKey); exist {
			value := attr.(map[string]interface{})
			if value != nil {
				item.AddProperties(value)
			}
			count++
		} else {
			itemIds = append(itemIds, item.Id)
			itemMap[string(item.Id)] = item
		}
	}
	p.context.LogDebug(fmt.Sprintf("module=TrafficControlSort\tevent=loadItemFeatureFromHolo\tcache hit=%d", count))
	if len(itemIds) == 0 || p.preloadItemFeature {
		return
	}

	padNum := getPaddingSize(len(itemIds))
	for i := 0; i < padNum; i++ {
		itemIds = append(itemIds, "-1")
	}

	selectFields := utils.UniqueStrings(fields)
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select(selectFields...)
	sb.From(table)
	sb.Where(sb.In(idField, itemIds...))
	sqlQuery, args := sb.Build()
	p.context.LogDebug(fmt.Sprintf("module=TrafficControlSort\tevnet=loadItemFeatureFromHolo\tquery:%s\targs:%v", sqlQuery, args))

	sort.Strings(selectFields)
	cols := strings.Join(selectFields, ",")
	hash := crc32.ChecksumIEEE([]byte(cols))
	tableKey := fmt.Sprintf("%s_%d_%d", table, len(selectFields), hash)
	numItems := len(itemIds)

	p.feaMu.RLock()
	feaStmtMap := p.feaStmtMap[tableKey]
	feaStmt := feaStmtMap[numItems]
	p.feaMu.RUnlock()

	if feaStmt == nil {
		p.feaMu.Lock()
		if feaStmtMap == nil {
			feaStmtMap = p.feaStmtMap[tableKey]
			if feaStmtMap == nil {
				feaStmtMap = make(map[int]*sql.Stmt)
				p.feaStmtMap[tableKey] = feaStmtMap
			}
		}
		feaStmt = feaStmtMap[numItems]
		if feaStmt == nil {
			if stmt, err := p.db.Prepare(sqlQuery); err != nil {
				p.context.LogError(fmt.Sprintf("model=TrafficControlSort\tevent=loadItemFeatureFromHolo\terr=%v", err))
				p.feaMu.Unlock()
				return
			} else {
				feaStmt = stmt
				feaStmtMap[numItems] = stmt
			}
		}
		p.feaMu.Unlock()
	}

	cntx, cancel := gocontext.WithTimeout(gocontext.Background(), 200*time.Millisecond)
	defer cancel()
	rows, err := feaStmt.QueryContext(cntx, args...)
	if err != nil {
		if errors.Is(err, gocontext.DeadlineExceeded) {
			p.context.LogWarning("module=TrafficControlSort\tevent=loadItemFeatureFromHolo\ttimeout=200")
			return
		}
		p.context.LogError(fmt.Sprintf("module=TrafficControlSort\terror=hologres error(%v)", err))
		return
	}
	defer rows.Close()

	columns, err := rows.ColumnTypes()
	if err != nil {
		p.context.LogError(fmt.Sprintf("module=TrafficControlSort\terror=hologres error(%v)", err))
		return
	}
	values := sqlutil.ColumnValues(columns)

	selected := make(map[string]bool, 0)
	for rows.Next() {
		if err = rows.Scan(values...); err == nil {
			features := make(map[string]interface{})
			for i, column := range columns {
				name := column.Name()
				val := values[i]
				switch v := val.(type) {
				case *sql.NullString:
					if v.Valid {
						features[name] = v.String
					}
				case *sql.NullInt32:
					if v.Valid {
						features[name] = v.Int32
					}
				case *sql.NullInt64:
					if v.Valid {
						features[name] = v.Int64
					}
				case *sql.NullFloat64:
					if v.Valid {
						features[name] = v.Float64
					}
				case *sql.NullBool:
					if v.Valid {
						features[name] = v.Bool
					}
				case *sql.NullTime:
					if v.Valid {
						features[name] = v.Time
					}
				default:
				}
			}
			if id, ok := features[idField]; ok {
				delete(features, idField)
				itemId := utils.ToString(id, "")
				cacheKey := table + "_" + itemId
				p.itemCache.Put(cacheKey, features)
				if item, okay := itemMap[itemId]; okay {
					item.AddProperties(features)
					selected[itemId] = true
				} else {
					p.context.LogError(fmt.Sprintf("module=TrafficControlSort\tInvalid item id: %s", itemId))
				}
			} else {
				p.context.LogError(fmt.Sprintf("module=TrafficControlSort\tevent=loadItemFeatureFromHolo\tNo item id found"))
			}
		} else {
			p.context.LogError(fmt.Sprintf("module=TrafficControlSort\terror=hologres error(%v)", err))
		}
	}

	// add nil to cache, prevent to loading from holo again and again
	var emptyMap map[string]interface{}
	for id := range itemMap {
		if _, ok := selected[id]; !ok {
			cacheKey := table + "_" + id
			p.itemCache.Put(cacheKey, emptyMap)
		}
	}
	p.context.LogDebug(fmt.Sprintf("module=TrafficControlSort\tevent=loadItemFeatureFromHolo\tload %d item features", len(selected)))
}

func (p *TrafficControlSort) loadTrafficControlTask(expId string) map[string]*PIDController {
	// 调用 SDK 获取调控计划的元信息, 创建 FlowControllers
	runEnv := os.Getenv("PAIREC_ENVIRONMENT")
	timePoint := p.config.TestTimestamp
	tasks := experimentClient.GetTrafficControlTaskMetaData(runEnv, timePoint)
	if len(tasks) == 0 {
		p.context.LogInfo(fmt.Sprintf("module=TrafficControlSort\tcurrent timestamp=%d\tNo Traffic Control Task.", timePoint))
		return nil
	}
	var oldControllerMap map[string]*PIDController
	p.controllerLock.RLock()
	if controllerMap, ok := p.exp2controllers[expId]; ok {
		oldControllerMap = controllerMap
	}
	p.controllerLock.RUnlock()

	controllerMap := make(map[string]*PIDController)
	for _, task := range tasks {
		// load task item express conditions
		itemCondition := task.ItemConditionArray
		var conditions []*Expression
		if itemCondition != "" {
			err := json.Unmarshal([]byte(itemCondition), &conditions)
			if err != nil {
				p.context.LogError(fmt.Sprintf("module=TrafficControlSort\tUnmarshal '%s' failed. err=%v", itemCondition, err))
				continue
			}
		}
		for _, target := range task.TrafficControlTargets {
			if target.Status == "Closed" {
				continue
			}
			params := experimentClient.GetSceneParams(task.SceneName)
			freeze := params.GetInt(fmt.Sprintf("pid_freeze_target_%s_minutes", target.TrafficControlTargetId), 0)
			run := params.GetString(fmt.Sprintf("pid_run_with_zero_input_%s", task.TrafficControlTaskId), "true")
			runWithZeroInput := strings.ToLower(run) == "true"

			if oldControllerMap != nil {
				if c, ok := oldControllerMap[target.TrafficControlTargetId]; ok {
					// update meta info
					c.task = &task
					c.target = &target
					controllerMap[target.TrafficControlTargetId] = c
					if conditions != nil {
						c.SetConditions(conditions)
					}

					c.GenerateItemConditions()
					c.SetFreezeMinutes(freeze)
					c.SetRunWithZeroInput(runWithZeroInput)
					continue
				}
			}

			c := NewPIDController(&task, &target, p.config, expId)
			if c != nil {
				if conditions != nil {
					c.SetConditions(conditions)
				}

				c.SetFreezeMinutes(freeze)
				c.SetRunWithZeroInput(runWithZeroInput)
				controllerMap[target.TrafficControlTargetId] = c
			}
		}
	}

	p.controllerLock.Lock()
	p.exp2controllers[expId] = controllerMap
	p.controllerLock.Unlock()
	p.context.LogInfo(fmt.Sprintf("module=TrafficControlSort\tcurrent timestamp=%d\tload %d Traffic Control Task for exp=%s.", timePoint, len(controllerMap), expId))
	return controllerMap
}

func loadTargetItemTraffic(ctx *context.RecommendContext, items []*module.Item, controllerMap map[string]*PIDController) map[string]map[string]float64 {
	var scene string
	var good bool
	s := ctx.GetParameter("scene")
	if scene, good = s.(string); !good {
		ctx.LogError("module=TrafficControlSort\tfailed to get scene name")
		return nil
	}

	itemIds := make([]string, 0, len(items))
	for i, item := range items {
		itemIds[i] = string(item.Id)
	}

	targetIdMap := make(map[string]bool)
	for targetId, _ := range controllerMap {
		targetIdMap[targetId] = true
	}

	// sdk 可能会返回已过期的Target下Item的历史流量，这样的话取最大值就是不对的
	result := make(map[string]map[string]float64) // key1: targetId, key2:expId, value: traffic
	runEnv := os.Getenv("PAIREC_ENVIRONMENT")
	traffics := experimentClient.GetTrafficControlTargetTraffic(runEnv, scene, itemIds...)
	hasTraffic := false
	for _, traffic := range traffics {
		if !targetIdMap[traffic.TrafficControlTargetId] {
			continue
		}

		if traffic.TargetTraffic <= 0 {
			continue
		}
		hasTraffic = true
		if dict, ok := result[traffic.TrafficControlTargetId]; ok {
			dict[traffic.ItemOrExpId] = traffic.TargetTraffic
		} else {
			dict = make(map[string]float64)
			dict[traffic.ItemOrExpId] = traffic.TargetTraffic
			result[traffic.TrafficControlTargetId] = dict
		}
	}
	if hasTraffic {
		ctx.LogDebug(fmt.Sprintf("item traffic:%v", result))
		return result
	}
	return nil
}

func (p *TrafficControlSort) getPidControllers(ctx *context.RecommendContext) map[string]*PIDController {
	var experiment string
	params := ctx.ExperimentResult.GetExperimentParams()
	expId := params.Get("pid_experiment_id", nil)
	expLayer := params.Get("pid_experiment_layer", nil)
	if expId != nil {
		experiment = expId.(string)
	} else if expLayer != nil {
		layer := expLayer.(string)
		n := len(layer)
		if !strings.Contains(layer, "#") {
			ctx.LogWarning(fmt.Sprintf("pid experiment layer `%s` maybe a prefix of another layer", layer))
		}
		recExpId := ctx.ExperimentResult.GetExpId()
		expIds := strings.Split(recExpId, "_")
		for i, id := range expIds {
			if i == 0 || len(id) < n {
				continue
			}
			if id[:n] == layer {
				experiment = id
				break
			}
		}
		if experiment == "" && recExpId != "" {
			ctx.LogError(fmt.Sprintf("parse pid experiment layer failed: `%s`", expLayer))
		}
	}
	p.controllerLock.RLock()
	if controllers, ok := p.exp2controllers[experiment]; ok {
		p.controllerLock.RUnlock()
		return controllers
	}
	p.controllerLock.RUnlock()

	return p.loadTrafficControlTask(experiment)
}

func splitController(controllers map[string]*PIDController, ctx *context.RecommendContext) (map[string]*PIDController, map[string]*PIDController) {
	wholeCtrls := make(map[string]*PIDController)
	singleCtrls := make(map[string]*PIDController)
	if nil == controllers || len(controllers) == 0 {
		return wholeCtrls, singleCtrls
	}
	for t, c := range controllers {
		if !c.IsControlledTraffic(ctx) {
			continue
		}
		if c.task.ControlGranularity == "Single" {
			singleCtrls[t] = c
		} else {
			wholeCtrls[t] = c
		}
	}
	return wholeCtrls, singleCtrls
}

// 宏观调控，针对目标整体
func macroControl(controllerMap map[string]*PIDController, items []*module.Item, ctx *context.RecommendContext, wgLoad, wgCtrl *sync.WaitGroup) {
	defer wgCtrl.Done()
	begin := time.Now()
	var planOutput map[string]float64
	var count int
	planOutput, count = FlowControl(controllerMap, ctx)
	if len(planOutput) == 0 || count == 0 {
		ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\tcount=%d\tcost=%d\tMacro FlowControl Not Need", len(items), utils.CostTime(begin)))
		return
	}
	ctx.LogInfo(fmt.Sprintf("module=PID_macro_control_signal\tcount=%d\tcost=%d", len(items), utils.CostTime(begin)))

	begin = time.Now()
	itemScores := make([]float64, 0, len(items))
	// 计算各个目标的偏好分的全局占比
	totalScore := 0.0
	maxScore := 0.0
	targetScore := make(map[string]float64)
	for i, item := range items {
		score := item.Score
		if score == 0.0 {
			score = 1e-8
		}
		if i == 0 {
			maxScore = score
			itemScores[0] = math.E
		} else {
			v := score / maxScore // 归一化 rank score
			idx := int(v * 1000)
			if idx < 0 {
				idx = 0
			}
			if idx >= 1000 {
				idx = 999
			}
			itemScores[i] = expTable[idx]
		}
		posWeight := 0.006737946999
		if i < 500 {
			posWeight = positionWeight[i]
		}
		score *= posWeight
		totalScore += score
		for targetId, controller := range controllerMap {
			if alpha, ok := planOutput[targetId]; ok && alpha != 0 && controller.IsControlledItem(ctx, item) {
				targetScore[targetId] += score
			}
		}
	}
	for targetId, score := range targetScore {
		targetScore[targetId] = score / totalScore
	}

	params := ctx.ExperimentResult.GetExperimentParams()
	maxUpliftTargetCnt := params.GetInt("pid_max_uplift_target_cnt", len(controllerMap))
	if maxUpliftTargetCnt < len(controllerMap) {
		// 按照偏好分采样 `maxUpliftTargetCnt` 个需要上提的目标，未被选中的上提目标调控力度置为0
		SampleControlTargetsByScore(maxUpliftTargetCnt, targetScore, planOutput, ctx)
	}

	pidGamma := params.GetFloat("pid_gamma", 1.0)
	pidBeta := params.GetFloat("pid_beta", 1.0)
	// preprocess, adjust control signal
	for targetId, alpha := range planOutput {
		if alpha > 0 { // uplift
			scoreWeight := targetScore[targetId]
			rho := 1.0 + pidGamma*tanh(pidBeta*scoreWeight) // 给更感兴趣的目标更大的提权，用来区分不同的调控目标
			alpha *= rho
			planOutput[targetId] = alpha
		}
	}
	ctx.LogInfo(fmt.Sprintf("module=PID_compute_uplift_score\tcount=%d\tcost=%d", len(items), utils.CostTime(begin)))

	wgLoad.Wait()

	// compute delta rank
	begin = time.Now()
	pageNo := utils.ToInt(ctx.GetParameter("pageNum"), 1)
	if pageNo < 1 {
		pageNo = 1
	}
	keepCtrlIdScore := params.GetFloat("pid_keep_id_target_score_weight", 1.0)
	if keepCtrlIdScore < 0.3 {
		keepCtrlIdScore = 0.3
	}
	ctrlParams := &controlParams{
		targetScore:        targetScore,
		itemScores:         itemScores,
		eta:                params.GetFloat("pid_eta", 1.6),
		pageNo:             pageNo,
		keepCtrlIdScore:    keepCtrlIdScore,
		newCtrlIdThreshold: params.GetFloat("pid_new_id_target_threshold", 1.0),
		needNewCtrlId:      make(map[string]bool),
	}
	for targetId, _ := range controllerMap {
		newCtrlId := utils.GetExperimentParamByPath(params, fmt.Sprintf("pid_params.%s.new_ctrl_id", targetId), false)
		ctrlParams.needNewCtrlId[targetId] = newCtrlId.(bool)
	}

	// compute delta rank
	parallel := params.GetInt("pid_parallel", 10)
	ch := make(chan int, parallel)
	defer close(ch)
	var wg sync.WaitGroup
	batchSize := len(items) / parallel
	if len(items)%parallel != 0 {
		batchSize++
	}
	if batchSize < 1 {
		batchSize = 1
	}
	for b, e := 0, batchSize; b < len(items); b, e = e, e+batchSize {
		var candidates []*module.Item
		if e < len(items) {
			candidates = items[b:e]
		} else {
			candidates = items[b:]
		}
		ch <- b
		wg.Add(1)
		go func(b int, items []*module.Item) {
			defer wg.Done()
			for idx, item := range items {
				i := b + idx
				finalDeltaRank := 0.0
				for t, c := range controllerMap {
					if !c.IsControlledItem(ctx, item) {
						continue
					}
					if alpha, ok := planOutput[t]; ok && alpha != 0 {
						deltaRank := computeDeltaRank(c, item, i, alpha, ctrlParams, ctx)
						finalDeltaRank += deltaRank // 形成合力
					}
				}

				if finalDeltaRank != 0.0 {
					item.IncrAlgoScore("__delta_rank__", finalDeltaRank)
				}
			}
			<-ch
		}(b, candidates)
	}
	wg.Wait()
	ctx.LogInfo(fmt.Sprintf("module=PID_macro_compute_delta_rank\tcount=%d\tcost=%d", len(items), utils.CostTime(begin)))
}

// FlowControl 非单品（整体）目标流量调控，返回各个目标的调控力度
func FlowControl(controllerMap map[string]*PIDController, ctx *context.RecommendContext) (map[string]float64, int) {
	// 获取(granularity="Global")类型的调控目标 当前已累计完成的流量
	planOutput := make(map[string]float64)
	if controllerMap == nil || len(controllerMap) == 0 {
		return planOutput, 0
	}

	var scene string
	var good bool
	s := ctx.GetParameter("scene")
	if scene, good = s.(string); !good {
		ctx.LogError("failed to get scene name")
		return planOutput, 0
	}
	// 获取流量实时统计值
	runEnv := os.Getenv("PAIREC_ENVIRONMENT")
	expId := ctx.ExperimentResult.GetExpId()
	traffics := experimentClient.GetTrafficControlTargetTraffic(runEnv, scene, expId, "ER_ALL")
	allTrafficDict := make(map[string]experiments.TrafficControlTargetTraffic)
	expTrafficDict := make(map[string]experiments.TrafficControlTargetTraffic)
	for _, traffic := range traffics {
		if traffic.ItemOrExpId == "ER_ALL" {
			allTrafficDict[traffic.TrafficControlTargetId] = traffic
		} else {
			expTrafficDict[traffic.TrafficControlTargetId] = traffic
		}
	}

	hasTraffic := false
	retCh := make(chan struct {
		string
		float64
	}, utils.MinInt(len(controllerMap), 64))
	defer close(retCh)

	gCtx, cancel := gocontext.WithTimeout(gocontext.Background(), time.Millisecond*12)
	defer cancel()
	for targetId, controller := range controllerMap {
		go func(gCtx gocontext.Context, targetId string, controller *PIDController) {
			defer func() {
				if err := recover(); err != nil {
					//stack := string(debug.Stack())
					log.Warning(fmt.Sprintf("traffic control timeout in background: <taskId:%s/targetId:%s>[targetName:%s]", controller.task.TrafficControlTaskId, targetId, controller.target.Name))
				}
			}()
			taskId := controller.task.TrafficControlTaskId
			var traffic, taskTraffic, output, setValue float64
			var binId = ""
			var dict map[string]experiments.TrafficControlTargetTraffic
			if controller.task.ControlType == "Percent" {
				if controller.IsAllocateExpWise() {
					dict = expTrafficDict
					binId = expId
				} else {
					dict = allTrafficDict
				}
				if input, ok := dict[targetId]; ok {
					traffic = input.TargetTraffic
					taskTraffic = input.TaskTraffic
				} else {
					traffic = float64(0)
					taskTraffic = float64(1)
				}
				if controller.IsAllocateExpWise() && traffic < controller.GetMinExpTraffic() {
					// 用全局流量代替冷启动的实验流量
					ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\t<taskId:%s/targetId:%s>[targetName:%s]\texp=%s\ttraffic=%f change to global traffic", taskId, targetId, controller.target.Name, expId, traffic))
					binId = ""
					if input, ok := allTrafficDict[targetId]; ok {
						traffic = input.TargetTraffic
						taskTraffic = input.TaskTraffic
					} else {
						traffic = float64(0)
						taskTraffic = float64(1)
					}
				}

				trafficPercentage := traffic / taskTraffic
				output, setValue = controller.DoWithId(trafficPercentage, binId)
				ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\t<taskId:%s/targetId:%s>[targetName:%s]\ttraffic=%f,percentage=%f,setValue=%f,output=%f,exp=%s", taskId, targetId, controller.target.Name, traffic, trafficPercentage, setValue, output, binId))
				if traffic > 0 {
					hasTraffic = true
				}
			} else {
				if input, ok := allTrafficDict[targetId]; ok {
					traffic = input.TargetTraffic
				} else {
					traffic = float64(0)
				}
				output, setValue = controller.Do(traffic)
				ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\t<taskId:%s/targetId%s>[targetName:%s]\ttraffic=%f,setValue=%f,output=%f", taskId, targetId, controller.target.Name, traffic, setValue, output))
				if traffic > 0 {
					hasTraffic = true
				}
			}
			select {
			case <-gCtx.Done():
				ctx.LogWarning(fmt.Sprintf("traffic controller timeout in goruntine: <taskId:%s/targetId:%s>[targetName:%s]", taskId, targetId, controller.target.Name))
				return
			case retCh <- struct {
				string
				float64
			}{targetId, output}:
			}
		}(gCtx, targetId, controller)
	}
	cnt := 0
Loop:
	for range controllerMap {
		select {
		case pair := <-retCh:
			output := pair.float64
			if output != 0 {
				cnt++
				planOutput[pair.string] = output
			}
		case <-gCtx.Done():
			if errors.Is(gCtx.Err(), gocontext.DeadlineExceeded) {
				ctx.LogWarning(fmt.Sprintf("traffic controller timeout: %v", gCtx.Err()))
			}
			break Loop
		}
	}
	if !hasTraffic {
		for k := range planOutput {
			delete(planOutput, k)
		}
		cnt = 0
	}
	return planOutput, cnt
}

// SampleControlTargetsByScore 按照偏好分权重选择n个上提目标，未被选中的目标调控值置0
func SampleControlTargetsByScore(maxUpliftTargetCnt int, targetScore, alpha map[string]float64, ctx *context.RecommendContext) {
	if maxUpliftTargetCnt >= len(targetScore) || maxUpliftTargetCnt <= 0 {
		return
	}
	targetIds := make([]string, 0, len(targetScore))
	scores := make([]float64, 0, len(targetScore))
	sum := 0.0
	for targetId, score := range targetScore {
		if alpha[targetId] > 0 { // only affect targets to be uplifted
			targetIds = append(targetIds, targetId)
			scores = append(scores, score)
			sum += score
		}
	}
	num := len(scores)
	if num == 0 || maxUpliftTargetCnt >= num {
		return
	}
	// normalize
	for j := range scores {
		scores[j] /= sum
	}

	w := sampleuv.NewWeighted(
		scores,
		rand.New(rand.NewSource(uint64(time.Now().UnixNano()))))

	selected := make(map[string]bool)
	for j := 0; j < maxUpliftTargetCnt; j++ {
		if i, ok := w.Take(); ok {
			targetId := targetIds[i]
			selected[targetId] = true
		}
	}
	ctx.LogDebug(fmt.Sprintf("selected uplift target ids %v score", selected))
	for _, targetId := range targetIds {
		if alpha[targetId] <= 0 {
			continue
		}
		if _, ok := selected[targetId]; !ok {
			alpha[targetId] = 0
		}
	}
}

// 微观调控，针对单个item
func microControl(controllerMap map[string]*PIDController, items []*module.Item, ctx *context.RecommendContext, wgLoad, wgCtrl *sync.WaitGroup) {
	defer wgCtrl.Done()
	begin := time.Now()
	itemTargetTraffic := loadTargetItemTraffic(ctx, items, controllerMap) // key1: targetId, key2: itemId, value: traffic
	ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\tcount=%d\tcost=%d", len(items), utils.CostTime(begin)))

	params := ctx.ExperimentResult.GetExperimentParams()
	maxUpliftCnt := params.GetInt("pid_max_uplift_item_cnt", 5)
	upliftCnt := 0

	wgLoad.Wait()
	begin = time.Now()
	maxScore := 0.0
	for i, item := range items {
		score := item.Score
		if score == 0 {
			score = 1e-8
		}
		if i == 0 {
			maxScore = score
		}
		deltaRank := 0.0
		for targetId, controller := range controllerMap {
			if !controller.IsControlledItem(ctx, item) {
				continue
			}

			traffic := float64(0)
			if dict, ok := itemTargetTraffic[targetId]; ok {
				if value, okay := dict[string(item.Id)]; okay {
					traffic = value
				}
			}
			alpha, setValue := controller.DoWithId(traffic, string(item.Id))
			delta := alpha
			pos, _ := item.IntProperty("_ORIGIN_POSITION_")
			ctrlId := pos
			if alpha > 0 { // uplift
				if i == 0 {
					delta *= math.E
				} else {
					v := score / maxScore // 归一化 rank score
					idx := int(v * 1000)
					if idx < 0 {
						idx = 0
					}
					if idx >= 1000 {
						idx = 999
					}
					delta *= expTable[idx]
				}
				if pos > ctx.Size {
					item.AddProperty("__flow_control_id__", 0)
					ctrlId = 0
				}
			}
			deltaRank += delta // 多个目标调控方向不一致时，需要扳手腕看谁力气大
			ctx.LogDebug(fmt.Sprintf("item %s [%s/%s], origin pos=%d, traffic=%f, setValue=%f, percentage=%f,alpha=%f, delta rank=%f, ctrl_id=%d", item.Id, targetId, controller.target.Name, pos, traffic, setValue, traffic/setValue, alpha, delta, ctrlId))
		}

		if deltaRank != 0.0 {
			if deltaRank < 0 {
				item.IncrAlgoScore("__delta_rank__", deltaRank)
			} else if upliftCnt < maxUpliftCnt { // uplift
				item.IncrAlgoScore("__delta_rank__", deltaRank)
				upliftCnt++
			}
		}
	}
	ctx.LogInfo(fmt.Sprintf("module=PID_micro_compute_delta_rank\tcount=%d\tcost=%d", len(items), utils.CostTime(begin)))
}

type controlParams struct {
	targetScore        map[string]float64
	itemScores         []float64
	eta                float64
	pageNo             int
	newCtrlIdThreshold float64
	keepCtrlIdScore    float64
	needNewCtrlId      map[string]bool
}

// computeDeltaRank 计算位置偏移值
func computeDeltaRank(c *PIDController, item *module.Item, rank int, alpha float64, args *controlParams, ctx *context.RecommendContext) float64 {
	scoreWeight := args.targetScore[c.target.TrafficControlTargetId]
	itemScore := args.itemScores[rank]
	var deltaRank = alpha
	if alpha < 0.0 { // pull down
		rho := args.eta * (1.0 - tanh(scoreWeight))
		deltaRank *= sigmoid(float64(rank), rho)
		ctx.LogDebug(fmt.Sprintf("item %s [%s/%s], score proportion=%.3f, rho=%.3f, origin pos=%d, delta rank=%f", item.Id, c.target.TrafficControlTargetId, c.target.Name, scoreWeight, rho, rank+1, deltaRank))
	} else { // uplift
		deltaRank *= itemScore // item.Score 越大，提权越多；用来在不同提取目标间竞争

		distinctStartPos := ctx.Size
		if scoreWeight > args.keepCtrlIdScore && args.pageNo > 1 {
			multiple := (scoreWeight - 0.3) * 10
			distinctStartPos += int(multiple * float64(ctx.Size))
		}
		if rank > distinctStartPos {
			needNewCtrlId := args.needNewCtrlId[c.target.TrafficControlTargetId] || c.target.SplitParts.SetValues[0] > int64(args.newCtrlIdThreshold)
			if c.task.ControlType == "Percent" && needNewCtrlId {
				item.AddProperty("__flow_control_id__", c.target.TrafficControlTargetId)
			} else {
				ctrlId, _ := item.IntProperty("__flow_control_id__")
				if ctrlId > 0 { // 已经被别的controller置为负数时不再更新为0
					item.AddProperty("__flow_control_id__", 0)
				}
			}
		}
		ctrlId, _ := item.IntProperty("__flow_control_id__")
		ctx.LogDebug(fmt.Sprintf("item:%s\t[targetId:%s/targetName:%s], score proportion=%.3f,norm_score=%.3f, origin pos=%d, delta rank=%f, ctrl_id=%d", item.Id, c.target.TrafficControlTargetId, c.target.Name, scoreWeight, itemScore, rank+1, deltaRank, ctrlId))
	}
	return deltaRank
}

type ItemRankSlice []*module.Item

func (us ItemRankSlice) Len() int {
	return len(us)
}

func (us ItemRankSlice) Less(i, j int) bool {
	iRank, _ := us[i].FloatProperty("_NEW_POSITION_")
	jRank, _ := us[j].FloatProperty("_NEW_POSITION_")
	if iRank != jRank {
		return iRank < jRank
	}

	iOriRank, _ := us[i].IntProperty("_ORIGIN_POSITION_")
	jOriRank, _ := us[j].IntProperty("_ORIGIN_POSITION_")
	return iOriRank < jOriRank
}

func (us ItemRankSlice) Swap(i, j int) {
	tmp := us[i]
	us[i] = us[j]
	us[j] = tmp
}

func getPaddingSize(n int) int {
	var t = 32
	for i := 5; i < n; i++ {
		t = 1 << i
		if t >= n {
			break
		}
	}
	return t - n
}

func getFilterFields(express string) []string {
	fields := make([]string, 0)
	if express == "" {
		return fields
	}
	var conditions []*Expression
	err := json.Unmarshal([]byte(express), &conditions)
	if err != nil {
		log.Error(fmt.Sprintf("module=TrafficControlSort\tUnmarshal '%s' failed. err=%v", express, err))
		return nil
	}

	for _, expr := range conditions {
		fields = append(fields, expr.Field)
	}
	return fields
}

func tanh(x float64) float64 {
	idx := int(x * 1000)
	if idx < 0 {
		idx = 0
	} else if idx >= 3000 {
		idx = 2999
	}
	return tanhTable[idx]
}

func sigmoid(x, rho float64) float64 {
	idx := int(rho*x*1000.0) - 5000.0
	if idx < 0 {
		idx = 0
	} else if idx >= 10000 {
		return 1
	}
	return sigmoidTable[idx]
}

func setHyperParams(controllers map[string]*PIDController, ctx *context.RecommendContext) bool {
	if nil == controllers || len(controllers) == 0 {
		return false
	}
	params := ctx.ExperimentResult.GetExperimentParams()
	on := params.GetInt("pid_control_enable", 1)
	if on == 0 {
		return false
	}

	offPrefix := params.GetString("pid_off_target_name_prefix", "")
	if offPrefix != "" {
		for _, c := range controllers {
			if strings.HasPrefix(c.target.Name, offPrefix) {
				c.SetOnline(false)
			}
		}
	}
	onPrefix := params.GetString("pid_on_target_name_prefix", "")
	if onPrefix != "" {
		for _, c := range controllers {
			if strings.HasPrefix(c.target.Name, onPrefix) {
				c.SetOnline(true)
			}
		}
	}

	planParams := params.Get("pid_plan_params", nil)
	if planParams != nil {
		if values, ok := planParams.(map[string]interface{}); ok {
			for pid, args := range values {
				if dict, good := args.(map[string]interface{}); good {
					if _on, exist := dict["online"]; exist {
						for _, c := range controllers {
							if c.task.TrafficControlTaskId == pid {
								c.SetOnline(_on.(bool))
							}
						}
					}
				}
			}
		}
	}

	hyperParams := params.Get("pid_params", nil)
	if hyperParams == nil {
		return true
	}
	if values, ok := hyperParams.(map[string]interface{}); ok {
		hasDefaultValue := false
		var defaultKp, defaultKi, defaultKd, defaultSampleTime, defaultErrDiscount float64
		var defaultStartPageNo = 0
		if args, exist := values["default"]; exist {
			if dict, good := args.(map[string]interface{}); good {
				hasDefaultValue = true
				if _kp, okay := dict["kp"]; okay {
					defaultKp = _kp.(float64)
				}
				if _ki, okay := dict["ki"]; okay {
					defaultKi = _ki.(float64)
				}
				if _kd, okay := dict["kd"]; okay {
					defaultKd = _kd.(float64)
				}
				if _t, okay := dict["sample_time"]; okay {
					defaultSampleTime = _t.(float64)
				}
				if _d, okay := dict["err_discount"]; okay {
					defaultErrDiscount = _d.(float64)
				}
				if _s, okay := dict["start_page_num"]; okay {
					defaultStartPageNo = int(_s.(float64))
				}
			}
		}
		if hasDefaultValue {
			for _, c := range controllers {
				if _, okay := values[c.target.TrafficControlTargetId]; !okay {
					if defaultKp != 0 {
						c.SetParameters(float32(defaultKp), float32(defaultKi), float32(defaultKd))
					}
					c.SetStartPageNum(defaultStartPageNo)
					c.SetTimeWindow(int(defaultSampleTime))
					c.SetErrDiscount(defaultErrDiscount)
				}
			}
		}
		for pid, args := range values {
			if pid == "default" {
				continue
			}
			if c, okay := controllers[pid]; okay {
				dict, good := args.(map[string]interface{})
				if !good {
					if hasDefaultValue {
						c.SetParameters(float32(defaultKp), float32(defaultKi), float32(defaultKd))
					}
					continue
				}
				var kp, ki, kd, sampleTime float64
				if _kp, exist := dict["kp"]; exist {
					kp = _kp.(float64)
				}
				if _ki, exist := dict["ki"]; exist {
					ki = _ki.(float64)
				}
				if _kd, exist := dict["kd"]; exist {
					kd = _kd.(float64)
				}
				c.SetParameters(float32(kp), float32(ki), float32(kd))
				if _sampleTime, exist := dict["sample_time"]; exist {
					sampleTime = _sampleTime.(float64)
					c.SetTimeWindow(int(sampleTime))
				}
				if discount, exist := dict["err_discount"]; exist {
					c.SetErrDiscount(discount.(float64))
				}
				if _exp, exist := dict["allocate_exp_wise"]; exist {
					c.SetAllocateExpWise(_exp.(bool))
				}
				if _s, exist := dict["start_page_num"]; exist {
					startPageNo := int(_s.(float64))
					c.SetStartPageNum(startPageNo)
				}
				if _s, exist := dict["min_exp_traffic"]; exist {
					c.SetMinExpTraffic(_s.(float64))
				}
				if _on, exist := dict["online"]; exist {
					c.SetOnline(_on.(bool))
				}
			}
		}
	}
	return true
}
