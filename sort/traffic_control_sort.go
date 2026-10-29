package sort

import (
	gocontext "context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	randv2 "math/rand/v2"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/abtest"
	"github.com/alibaba/pairec/v2/constants"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
	"github.com/expr-lang/expr/vm"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/experiments"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/sampleuv"
)

type TrafficControlSort struct {
	name           string
	config         *recconf.PIDControllerConfig
	controllersMap map[string]*PIDController // key: targetId
	controllerLock sync.RWMutex
	cloneInstances map[string]*TrafficControlSort
}

var positionWeight []float64
var expTable []float64
var tanhTable []float64
var sigmoidTable []float64
var experimentClient *experiments.ExperimentClient

// TrafficControlHitsProperty is the item property holding the traffic control tasks
// and targets that actually adjusted the position of the item, rendered as
// "taskId1:targetId1,targetId2|taskId2:targetId3". It is always set (an empty string
// when no task took effect), so it can be exposed as a required output field through
// SceneConfs.OutputFields with "item:traffic_control_hits".
const TrafficControlHitsProperty = "traffic_control_hits"

// maxTrafficControlParallel 限制单个请求在大规格机器上的并发占用，
// 同时允许小规格容器充分使用其 CPU 配额。
const maxTrafficControlParallel = 8

// macroControl and microControl traverse the same items concurrently, so each writes
// the tasks and targets it hits to its own property key, as a map of task id to target
// ids. Sort merges them once both finish.
const (
	macroTrafficControlHitsProperty = "__macro_traffic_control_hits__"
	microTrafficControlHitsProperty = "__micro_traffic_control_hits__"
)

func init() {
	positionWeight = make([]float64, 500)
	for i := 0; i < 500; i++ {
		positionWeight[i] = math.Exp(-0.01 * float64(i))
	}

	expTable = make([]float64, 1000) // 值域范围 [1, 2.71+]
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
		log.Warning("module=TrafficControlSort\tget experiment client failed.")
	}
	conf := config.PIDConf
	trafficControlSort := TrafficControlSort{
		config:         &conf,
		controllersMap: make(map[string]*PIDController),
		name:           config.Name,
		cloneInstances: make(map[string]*TrafficControlSort),
	}

	go func() {
		for {
			trafficControlSort.loadTrafficControllersMap()
			time.Sleep(time.Minute)
		}
	}()

	return &trafficControlSort
}

func (p *TrafficControlSort) Sort(sortData *SortData) error {
	items, ok := sortData.Data.([]*module.Item)
	if !ok {
		msg := "sort data type error"
		log.Error(fmt.Sprintf("module=TrafficControlSort\terror=%s", msg))
		return errors.New(msg)
	}
	if len(items) == 0 {
		return nil
	}

	user := sortData.User

	start := time.Now()
	ctx := sortData.Context

	layerExperimentParams := ctx.ExperimentResult.GetExperimentParams()
	experimentParams := newExperimentParams(layerExperimentParams)
	if ctx.Debug {
		d, _ := json.Marshal(experimentParams)
		ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\texperiment traffic control params: %s", string(d)))
	}

	sort.Sort(sort.Reverse(ItemScoreSlice(items)))
	for i, item := range items {
		item.AddProperty("__traffic_control_id__", i+1)
		item.AddProperty("_ORIGIN_POSITION_", i+1)
		// default to an empty string so items hit by no task still return the field
		item.AddProperty(TrafficControlHitsProperty, "")
	}

	// 如果服务启动时，没有加载成功，这里再次尝试
	var allControllersMap map[string]*PIDController
	p.controllerLock.RLock()
	controllersMap := p.controllersMap
	p.controllerLock.RUnlock()
	if len(controllersMap) == 0 {
		allControllersMap = p.loadTrafficControllersMap()
	} else {
		allControllersMap = controllersMap
	}

	validControllersMap := filterValidControllers(ctx, user, experimentParams, allControllersMap)
	sampleCount := targetSampleCount(p.config, experimentParams)
	selectedControllers := sampleControllers(validControllersMap, sampleCount, randv2.IntN)
	if ctx.Debug {
		ids := make([]string, 0, len(selectedControllers))
		for id := range selectedControllers {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tselected target ids=%v", ids))
	}

	globalControls, singleControls := splitControllers(selectedControllers)
	if len(globalControls) == 0 && len(singleControls) == 0 {
		ctx.LogWarning(fmt.Sprintf("module=TrafficControlSort\tboth global traffic control and single traffic control are zero"))
		sortData.Data = items
		return nil
	}

	runEnv := os.Getenv("PAIREC_ENVIRONMENT")
	traffic := selectTrafficControlTraffic(experimentClient.ListTrafficControlTasks(runEnv), runEnv, ctx.ExperimentResult.GetExpId(), validControllersMap, selectedControllers, items)
	itemFeatures := snapshotItemFeatures(items)
	macroParallel, microParallel := trafficControlParallelism(len(items), len(globalControls), len(singleControls), runtime.GOMAXPROCS(0))
	if ctx.Debug {
		ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tglobal control num %d, single control num %d, macro parallel %d, micro parallel %d", len(globalControls), len(singleControls), macroParallel, microParallel))
	}
	wgCtrl := sync.WaitGroup{}
	if len(singleControls) > 0 {
		wgCtrl.Add(1)
		go microControl(ctx, singleControls, items, itemFeatures, microParallel, &wgCtrl, experimentParams, traffic.Items)
	}
	if len(globalControls) > 0 {
		wgCtrl.Add(1)
		go macroControl(ctx, globalControls, items, itemFeatures, macroParallel, &wgCtrl, experimentParams, traffic)
	}
	wgCtrl.Wait()
	//pageNumber := utils.ToInt(ctx.GetParameter("page_number"), 1)
	//pageSize := ctx.Size
	//if pageNumber < 1 {
	//	pageNumber = 1
	//}
	//var candidateCount int
	//if experimentParams.CandidateCountAfterFirstPage != nil {
	//	candidateCount = *experimentParams.CandidateCountAfterFirstPage
	//	ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tlimit_uplift_at_first_page=%d", candidateCount))
	//} else {
	//	candidateCount = 0
	//}
	for i, item := range items {
		mergeTrafficControlHits(item)
		finalDeltaRank := item.GetAlgoScore("__delta_rank__")
		if finalDeltaRank != 0.0 {
			rank := float64(i+1) - finalDeltaRank
			//if pageNumber <= 1 && candidateCount != 0 {
			//	if i < pageSize {
			//		item.AddProperty("_NEW_POSITION_", i+1)
			//	} else {
			//		if rank <= float64(pageSize) { // 保证第一页流量调控的结果仅作为打散的候补出现
			//			rank = float64(pageSize) + 1 + tanh(0.001*rank) // rank > pageSize
			//		}
			//		item.AddProperty("_NEW_POSITION_", rank)
			//	}
			//} else {
			//
			//}
			item.AddProperty("_NEW_POSITION_", rank)
		} else {
			item.AddProperty("_NEW_POSITION_", i+1)
		}
	}
	sort.Sort(ItemRankSlice(items))
	ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\tcount=%d\tcost=%d\tvalid_target_count=%d\tselected_target_count=%d\ttarget_sample_count=%d", len(items), utils.CostTime(start), len(validControllersMap), len(selectedControllers), sampleCount))
	sortData.Data = items
	return nil
}

const defaultTargetSampleCount = 50

func targetSampleCount(config *recconf.PIDControllerConfig, params *ExperimentParams) int {
	count := 0
	if config != nil {
		count = config.TargetSampleCount
	}
	if params != nil && params.PidTargetSampleCount != nil {
		count = *params.PidTargetSampleCount
	}
	if count <= 0 {
		return defaultTargetSampleCount
	}
	return count
}

// sampleControllers 对任意遍历顺序执行等概率、无放回抽样，不修改共享控制器和状态。
func sampleControllers(controllers map[string]*PIDController, count int, intN func(int) int) map[string]*PIDController {
	if count <= 0 {
		count = defaultTargetSampleCount
	}
	if len(controllers) <= count {
		return controllers
	}
	ids := make([]string, 0, count)
	seen := 0
	for id := range controllers {
		seen++
		if len(ids) < count {
			ids = append(ids, id)
		} else if index := intN(seen); index < count {
			ids[index] = id
		}
	}
	selected := make(map[string]*PIDController, count)
	for _, id := range ids {
		selected[id] = controllers[id]
	}
	return selected
}

func splitControllers(controllers map[string]*PIDController) (map[string]*PIDController, map[string]*PIDController) {
	wholeCtrls := make(map[string]*PIDController)
	singleCtrls := make(map[string]*PIDController)
	if controllers == nil || len(controllers) == 0 {
		return wholeCtrls, singleCtrls
	}
	for targetId, controller := range controllers {
		if controller.task.ControlGranularity == constants.TrafficControlTaskControlGranularitySingle {
			singleCtrls[targetId] = controller
		} else {
			wholeCtrls[targetId] = controller
		}
	}
	return wholeCtrls, singleCtrls
}

// trafficControlParallelism 分配单请求的 worker 预算。
// macro 当前会遍历每个 item-target 组合两次，因此工作量权重按两倍计算。
func trafficControlParallelism(itemCount, macroCount, microCount, cpuBudget int) (macroParallel, microParallel int) {
	if itemCount <= 0 || (macroCount <= 0 && microCount <= 0) {
		return 0, 0
	}
	if cpuBudget < 1 {
		cpuBudget = 1
	}
	budget := min(cpuBudget, maxTrafficControlParallel, itemCount)

	if microCount <= 0 {
		return budget, 0
	}
	if macroCount <= 0 {
		return 0, budget
	}
	if budget == 1 {
		return 1, 1
	}

	macroWork := 2 * macroCount
	microWork := microCount
	macroParallel = budget * macroWork / (macroWork + microWork)
	macroParallel = max(1, min(macroParallel, budget-1))
	microParallel = budget - macroParallel
	return macroParallel, microParallel
}

// activeMacroControllers 只保留当前输出非零 alpha 的宏观调控目标，
// 避免已达标或被采样关闭的目标继续执行物品条件表达式。
func snapshotItemFeatures(items []*module.Item) []map[string]interface{} {
	features := make([]map[string]interface{}, len(items))
	for i, item := range items {
		features[i] = item.GetFeatures()
	}
	return features
}

func matchMacroTargets(items []*module.Item, itemFeatures []map[string]interface{}, activeControllers map[string]*PIDController, weightedScores []float64, parallel int) ([][]string, map[string]float64) {
	matchedTargets := make([][]string, len(items))
	targetScore := make(map[string]float64)
	if len(items) == 0 || len(activeControllers) == 0 {
		return matchedTargets, targetScore
	}

	parallel = max(1, min(parallel, len(items)))
	batchSize := (len(items) + parallel - 1) / parallel
	localTargetScores := make([]map[string]float64, parallel)
	var wg sync.WaitGroup
	for workerIndex := 0; workerIndex < parallel; workerIndex++ {
		begin := workerIndex * batchSize
		if begin >= len(items) {
			break
		}
		end := min(begin+batchSize, len(items))
		wg.Add(1)
		go func(workerIndex, begin, end int) {
			defer wg.Done()
			localScores := make(map[string]float64)
			var machine vm.VM
			for i := begin; i < end; i++ {
				for targetId, controller := range activeControllers {
					if isControlledItemFeaturesWithVM(controller, items[i].Id, itemFeatures[i], &machine) {
						localScores[targetId] += weightedScores[i]
						matchedTargets[i] = append(matchedTargets[i], targetId)
					}
				}
			}
			localTargetScores[workerIndex] = localScores
		}(workerIndex, begin, end)
	}
	wg.Wait()

	for _, localScores := range localTargetScores {
		for targetId, score := range localScores {
			targetScore[targetId] += score
		}
	}
	return matchedTargets, targetScore
}

func activeMacroControllers(controllerMap map[string]*PIDController, targetAlphaMap map[string]float64) map[string]*PIDController {
	activeControllers := make(map[string]*PIDController, len(targetAlphaMap))
	for targetId, alpha := range targetAlphaMap {
		if alpha == 0 {
			continue
		}
		if controller, ok := controllerMap[targetId]; ok {
			activeControllers[targetId] = controller
		}
	}
	return activeControllers
}

func upliftItemLimit(experimentParams *ExperimentParams) int {
	if experimentParams != nil && experimentParams.PidSingleControlLimitCount != nil {
		return *experimentParams.PidSingleControlLimitCount
	}
	return 5
}

// selectMacroQuantityUpliftItems 按原始排名为所有按量全局目标选择共享的上提物品名额。
// 同一物品命中多个正向按量目标时只占一个名额。
func selectMacroQuantityUpliftItems(matchedTargets [][]string, activeControllers map[string]*PIDController, targetAlphaMap map[string]float64, limit int) []bool {
	selected := make([]bool, len(matchedTargets))
	if limit <= 0 {
		return selected
	}

	selectedCount := 0
	for itemIndex, targetIds := range matchedTargets {
		for _, targetId := range targetIds {
			controller, ok := activeControllers[targetId]
			if !ok || controller.task.ControlType != constants.TrafficControlTaskControlTypeQuantity || targetAlphaMap[targetId] <= 0 {
				continue
			}
			selected[itemIndex] = true
			selectedCount++
			break
		}
		if selectedCount >= limit {
			break
		}
	}
	return selected
}

func shouldApplyMacroTarget(controller *PIDController, alpha float64, quantityUpliftSelected bool) bool {
	return alpha <= 0 || controller.task.ControlType != constants.TrafficControlTaskControlTypeQuantity || quantityUpliftSelected
}

// 宏观调控，针对目标整体
func macroControl(ctx *context.RecommendContext, controllerMap map[string]*PIDController, items []*module.Item, itemFeatures []map[string]interface{}, parallel int, wgCtrl *sync.WaitGroup, experimentParams *ExperimentParams, traffic trafficControlTraffic) {
	defer wgCtrl.Done()
	begin := time.Now()
	targetAlphaMap := computeMacroControlAlpha(ctx, controllerMap, experimentParams, parallel, traffic)
	if len(targetAlphaMap) == 0 {
		ctx.LogWarning(fmt.Sprintf("module=TrafficControlSort\tmacro control\ttraffic control task output is zero"))
		return
	}
	if ctx.Debug {
		d, _ := json.Marshal(targetAlphaMap)
		ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tmacro control\tthe relationship between target and alpha: %s", string(d)))
	}
	activeControllers := activeMacroControllers(controllerMap, targetAlphaMap)

	itemScores := make([]float64, len(items))
	weightedScores := make([]float64, len(items))
	// 预计算与 target 无关的分数，再并行计算 item-target 命中关系。
	totalScore := 0.0
	maxScore := 0.0 // item 列表中的最大分
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
		weightedScores[i] = score * posWeight
		totalScore += weightedScores[i]
	}
	matchedTargets, targetScore := matchMacroTargets(items, itemFeatures, activeControllers, weightedScores, parallel)
	for targetId, score := range targetScore {
		targetScore[targetId] = score / totalScore
	}

	var maxUpliftTargetCnt int
	if experimentParams.PidMaxUpliftItemCnt != nil {
		maxUpliftTargetCnt = *experimentParams.PidMaxUpliftItemCnt
	} else {
		maxUpliftTargetCnt = len(controllerMap)
	}
	if maxUpliftTargetCnt < len(controllerMap) {
		// 按照偏好分采样 `maxUpliftTargetCnt` 个需要上提的目标，未被选中的上提目标调控力度置为0
		sampleControlTargetsByScore(ctx, maxUpliftTargetCnt, targetScore, targetAlphaMap)
		activeControllers = activeMacroControllers(activeControllers, targetAlphaMap)
	}

	var (
		pidGamma = 1.0
		pidBeta  = 1.0
		pidEta   = 1.6
	)
	if experimentParams.PidGamma != nil {
		pidGamma = *experimentParams.PidGamma
	}
	if experimentParams.PidBeta != nil {
		pidBeta = *experimentParams.PidBeta
	}
	if experimentParams.PidEta != nil {
		pidEta = *experimentParams.PidEta
	}
	// preprocess, adjust control signal
	for targetId, alpha := range targetAlphaMap {
		if alpha > 0 { // uplift
			scoreWeight := targetScore[targetId]
			rho := 1.0 + pidGamma*tanh(pidBeta*scoreWeight) // 给更感兴趣的目标更大的提权，用来区分不同的调控目标
			alpha *= rho
			targetAlphaMap[targetId] = alpha
		}
	}
	if ctx.Debug {
		d, _ := json.Marshal(targetAlphaMap)
		ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tmacro control\tthe relationship between target and alpha after increase weight: %s", string(d)))
	}
	quantityUpliftSelected := selectMacroQuantityUpliftItems(matchedTargets, activeControllers, targetAlphaMap, upliftItemLimit(experimentParams))
	// compute delta rank
	pageNo := utils.ToInt(ctx.GetParameter("page_number"), 1)
	if pageNo < 1 {
		pageNo = 1
	}

	ctrlParams := &controlParams{
		targetScore: targetScore,
		itemScores:  itemScores,
		eta:         pidEta,
		pageNo:      pageNo,
	}

	// 统计每个 target 实际调控的物品数，key: targetId; value: controlled item count
	targetControlledNum := make(map[string]int, len(controllerMap))
	workerControlledNums := make([]map[string]int, 0, parallel)

	// compute delta rank
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
		localControlledNum := make(map[string]int)
		workerControlledNums = append(workerControlledNums, localControlledNum)
		wg.Add(1)
		go func(b int, items []*module.Item, localControlledNum map[string]int) {
			defer wg.Done()
			for idx, item := range items {
				i := b + idx
				finalDeltaRank := 0.0
				var hitTargets map[string][]string
				for _, targetId := range matchedTargets[i] {
					controller, ok := activeControllers[targetId]
					if !ok {
						continue
					}
					alpha := targetAlphaMap[targetId]
					if !shouldApplyMacroTarget(controller, alpha, quantityUpliftSelected[i]) {
						continue
					}
					deltaRank := computeDeltaRank(controller, item, i, alpha, ctrlParams, ctx)
					finalDeltaRank += deltaRank // 形成合力
					if deltaRank != 0.0 {
						localControlledNum[targetId]++
						if hitTargets == nil {
							hitTargets = make(map[string][]string)
						}
						taskId := controller.task.TrafficControlTaskId
						hitTargets[taskId] = append(hitTargets[taskId], targetId)
					}
				}

				if finalDeltaRank != 0.0 {
					item.IncrAlgoScore("__delta_rank__", finalDeltaRank)
					controlId, _ := item.IntProperty("__traffic_control_id__")
					if controlId == 0 && finalDeltaRank < 1.0 {
						item.AddProperty("__traffic_control_id__", item.GetProperty("_ORIGIN_POSITION_"))
					}
					// only record when the delta rank is really applied to the item
					if len(hitTargets) > 0 {
						item.AddProperty(macroTrafficControlHitsProperty, hitTargets)
					}
				}
			}
			<-ch
		}(b, candidates, localControlledNum)
	}
	wg.Wait()
	for _, localControlledNum := range workerControlledNums {
		for targetId, count := range localControlledNum {
			targetControlledNum[targetId] += count
		}
	}
	ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\tmacro control\tcount=%d\tcost=%d\tcontrolled item num=%v", len(items), utils.CostTime(begin), targetControlledNum))
}

// computeMacroControlAlpha 非单品（整体）目标流量调控，返回各个目标的调控力度。
func computeMacroControlAlpha(ctx *context.RecommendContext, controllerMap map[string]*PIDController, experimentParams *ExperimentParams, parallel int, traffic trafficControlTraffic) map[string]float64 {
	targetAlphaMap := make(map[string]float64)
	if len(controllerMap) == 0 {
		return targetAlphaMap
	}

	expId := ctx.ExperimentResult.GetExpId()
	if ctx.Debug {
		data, _ := json.Marshal(struct {
			All, Experiment map[string]*experiments.TrafficControlTargetTraffic
		}{traffic.All, traffic.Experiment})
		ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tmacro control\treport traffics: %s", string(data)))
	}
	allTrafficMap, expTrafficMap := traffic.All, traffic.Experiment

	controllerParamsMap := precomputeControllerParams(controllerMap, experimentParams)
	workerCount := max(1, min(parallel, len(controllerMap)))
	jobs := make(chan macroAlphaJob, len(controllerMap))
	results := make(chan targetAlpha, workerCount)
	gCtx, cancel := gocontext.WithTimeout(gocontext.Background(), time.Millisecond*12)
	defer cancel()

	for targetId, controller := range controllerMap {
		jobs <- macroAlphaJob{
			TargetId:   targetId,
			Controller: controller,
			Params:     controllerParamsMap[targetId],
		}
	}
	close(jobs)

	var wg sync.WaitGroup
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for {
				if gCtx.Err() != nil {
					return
				}
				select {
				case <-gCtx.Done():
					return
				case job, ok := <-jobs:
					if !ok {
						return
					}
					result, ok := computeMacroTargetAlpha(ctx, expId, allTrafficMap, expTrafficMap, job)
					if !ok {
						continue
					}
					select {
					case results <- result:
					case <-gCtx.Done():
						return
					}
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	timedOut := false
	collectResult := func(result targetAlpha) {
		if result.Alpha != 0 {
			targetAlphaMap[result.TargetId] = result.Alpha
		}
	}

Collect:
	for {
		select {
		case result, ok := <-results:
			if !ok {
				break Collect
			}
			collectResult(result)
		case <-gCtx.Done():
			timedOut = errors.Is(gCtx.Err(), gocontext.DeadlineExceeded)
			for result := range results {
				collectResult(result)
			}
			break Collect
		}
	}
	if timedOut {
		ctx.LogWarning(fmt.Sprintf("module=TrafficControlSort\tmacro control\ttraffic controller timeout: %v", gCtx.Err()))
	}
	if !traffic.HasTraffic {
		ctx.LogWarning(fmt.Sprintf("module=TrafficControlSort\tmacro control\tno traffic data detected, maybe flink job is not running"))
		clear(targetAlphaMap)
	}
	return targetAlphaMap
}

func computeMacroTargetAlpha(ctx *context.RecommendContext, expId string, allTrafficMap, expTrafficMap map[string]*experiments.TrafficControlTargetTraffic, job macroAlphaJob) (result targetAlpha, ok bool) {
	controller := job.Controller
	targetId := job.TargetId
	result.TargetId = targetId
	defer func() {
		if err := recover(); err != nil {
			log.Warning(fmt.Sprintf("module=TrafficControlSort\tmacro control\tcompute alpha panic\terr=%v\t%s", err, trafficIdentity(controller.task.TrafficControlTaskId, targetId, controller.target.Name)))
			ok = false
		}
	}()

	taskId := controller.task.TrafficControlTaskId
	var targetTraffic, taskTraffic, alpha, aimValue float64
	experimentId := ""
	var measureTime time.Time
	if controller.task.ControlType == constants.TrafficControlTaskControlTypePercent {
		targetTrafficMap := allTrafficMap
		if controller.IsAllocateExpWise() {
			targetTrafficMap = expTrafficMap
			experimentId = expId
		}
		if trafficInfo, exists := targetTrafficMap[targetId]; exists {
			targetTraffic = trafficInfo.TargetTraffic
			taskTraffic = trafficInfo.TaskTraffic
			measureTime = trafficInfo.RecordTime
		} else {
			taskTraffic = 1
			measureTime = time.Now().Truncate(time.Second)
		}
		if controller.IsAllocateExpWise() && targetTraffic < controller.GetMinExpTraffic() {
			if ctx.Debug {
				ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tmacro control\t%s\texp=%s\ttargetTraffic=%.0f change to global targetTraffic", trafficIdentity(taskId, targetId, controller.target.Name), expId, targetTraffic))
			}
			experimentId = ""
			if trafficInfo, exists := allTrafficMap[targetId]; exists {
				targetTraffic = trafficInfo.TargetTraffic
				taskTraffic = trafficInfo.TaskTraffic
				measureTime = trafficInfo.RecordTime
			} else {
				targetTraffic = 0
				taskTraffic = 1
				measureTime = time.Now().Truncate(time.Second)
			}
		}
		trafficPercentage := targetTraffic / taskTraffic
		controller.setMeasurement(experimentId, "", trafficPercentage, measureTime, job.Params)
		alpha, aimValue = controller.compute(ctx, experimentId, "", job.Params)
		if ctx.Debug {
			ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tmacro control\t%s\tpercentage=%f, aimValue=%f, alpha=%f", trafficIdentity(taskId, targetId, controller.target.Name), trafficPercentage, aimValue/100, alpha))
		}
	} else {
		if trafficInfo, exists := allTrafficMap[targetId]; exists {
			targetTraffic = trafficInfo.TargetTraffic
			measureTime = trafficInfo.RecordTime
		} else {
			measureTime = time.Now().Truncate(time.Second)
		}
		controller.setMeasurement("", "", targetTraffic, measureTime, job.Params)
		alpha, aimValue = controller.compute(ctx, "", "", job.Params)
		if ctx.Debug {
			ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\tmacro control\t%s\taimValue=%f, alpha=%f", trafficIdentity(taskId, targetId, controller.target.Name), aimValue, alpha))
		}
	}
	result.Alpha = alpha
	return result, true
}

// SampleControlTargetsByScore 按照偏好分权重选择n个上提目标，未被选中的目标调控值置0
func sampleControlTargetsByScore(ctx *context.RecommendContext, maxUpliftTargetCnt int, targetScore, targetAlpha map[string]float64) {
	if maxUpliftTargetCnt >= len(targetScore) || maxUpliftTargetCnt <= 0 {
		return
	}
	targetIds := make([]string, 0, len(targetScore))
	scores := make([]float64, 0, len(targetScore))
	sum := 0.0
	for targetId, score := range targetScore {
		if targetAlpha[targetId] > 0 { // only affect targets to be uplifted
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
	if ctx.Debug {
		ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tsample control target\tselected uplift target ids %v score", selected))
	}
	for _, targetId := range targetIds {
		if targetAlpha[targetId] <= 0 {
			continue
		}
		if _, ok := selected[targetId]; !ok {
			targetAlpha[targetId] = 0
		}
	}
}

type controlParams struct {
	targetScore map[string]float64
	itemScores  []float64
	eta         float64
	pageNo      int
}

// computeDeltaRank 计算位置偏移值
func computeDeltaRank(c *PIDController, item *module.Item, itemIndex int, alpha float64, args *controlParams, ctx *context.RecommendContext) float64 {
	scoreWeight := args.targetScore[c.target.TrafficControlTargetId]
	itemScore := args.itemScores[itemIndex]
	var deltaRank = alpha
	if alpha < 0.0 { // pull down
		rho := args.eta * (1.0 - tanh(scoreWeight))
		deltaRank *= sigmoid(float64(itemIndex), rho)
		if ctx.Debug {
			ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tmacro control\tcompute delta rank\titem_id[%s]\t%s\tscore_weight=%.3f, rho=%.3f, alpha=%.6f, origin_pos=%d, delta_rank=%.6f [pull down]", item.Id, trafficIdentity(c.target.TrafficControlTaskId, c.target.TrafficControlTargetId, c.target.Name), scoreWeight, rho, alpha, itemIndex+1, deltaRank))
		}
	} else { // uplift
		deltaRank *= itemScore // item.Score 越大，提权越多；用来在不同提取目标间竞争
		distinctStartPos := ctx.Size
		if args.pageNo > 1 {
			multiple := (scoreWeight - 0.3) * 10
			distinctStartPos += int(multiple * float64(ctx.Size))
		}
		if itemIndex >= distinctStartPos && deltaRank >= 1.0 {
			if c.task.ControlType == constants.TrafficControlTaskControlTypePercent {
				targetId, _ := strconv.Atoi(c.target.TrafficControlTargetId)
				item.AddProperty("__traffic_control_id__", -targetId) //改成负的
			} else {
				controlId, _ := item.IntProperty("__traffic_control_id__")
				if controlId > 0 { // 已经被别的controller置为负数时不再更新为0
					item.AddProperty("__traffic_control_id__", 0)
				}
			}
		}
		if ctx.Debug {
			controlId, _ := item.IntProperty("__traffic_control_id__")
			ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tmacro control\tcompute delta rank\titem_id[%s]\t%s\tscore_weight=%.3f,item_score=%.3f, alpha=%.6f, origin_pos=%d, delta_rank=%.6f, traffic_control_id=%d [uplift]", item.Id, trafficIdentity(c.target.TrafficControlTaskId, c.target.TrafficControlTargetId, c.target.Name), scoreWeight, itemScore, alpha, itemIndex+1, deltaRank, controlId))
		}
	}
	return deltaRank
}

// 微观调控，针对单个item
func microControl(ctx *context.RecommendContext, controllerMap map[string]*PIDController, items []*module.Item, itemFeatures []map[string]interface{}, parallel int, wg *sync.WaitGroup, experimentParams *ExperimentParams, actualTrafficsMap map[string][]*experiments.TrafficControlTargetTraffic) {
	defer wg.Done()
	start := time.Now()
	controllerParamsMap := precomputeControllerParams(controllerMap, experimentParams)
	targetItemActualTrafficMap := getItemActualTraffic(controllerMap, actualTrafficsMap, controllerParamsMap) // key1: targetId
	if ctx.Debug {
		data, _ := json.Marshal(targetItemActualTrafficMap)
		ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tmicro control\treport traffics:%s", string(data)))
	}
	limitCount := upliftItemLimit(experimentParams)
	maxScore := 0.0
	for _, item := range items {
		score := item.Score
		if score == 0 {
			score = 1e-8
		}
		if score > maxScore {
			maxScore = score
		}
	}

	// Calculate the number of items controlled by each target, key: targetId; value: controlled item sum.
	// Workers write to local maps and reduce after completion to avoid a lock for every item-target hit.
	controlNumberMap := make(map[string]int, len(controllerMap))
	controlNumberCh := make(chan map[string]int, parallel)

	var countMu sync.Mutex
	upliftCount := 0

	// 控制并发的组数量
	sem := make(chan int, parallel)
	defer close(sem)
	batchSize := len(items) / parallel
	if len(items)%parallel != 0 {
		batchSize++
	}
	if batchSize < 1 {
		batchSize = 1
	}
	var innerWg sync.WaitGroup
	for begin, end := 0, batchSize; begin < len(items); begin, end = end, end+batchSize {
		var candidates []*module.Item
		if end < len(items) {
			candidates = items[begin:end]
		} else {
			candidates = items[begin:]
		}
		sem <- begin
		innerWg.Add(1)
		go func(begin int, items []*module.Item) {
			defer innerWg.Done()
			localControlNumberMap := make(map[string]int, len(controllerMap))
			var machine vm.VM
			for j, item := range items {
				i := begin + j
				deltaRank := 0.0
				var hitTargets map[string][]string
				for targetId, controller := range controllerMap {
					if !isControlledItemFeaturesWithVM(controller, item.Id, itemFeatures[i], &machine) {
						continue
					}
					localControlNumberMap[targetId]++
					if ctx.Debug {
						ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tmicro control\titem_id[%v] is controlled by traffic control target: %s", item.Id, trafficIdentity(controller.target.TrafficControlTaskId, controller.target.TrafficControlTargetId, controller.target.Name)))
					}

					params := controllerParamsMap[targetId]
					alpha, aimValue := controller.compute(ctx, "", string(item.Id), params)
					delta := alpha
					originPosition, _ := item.IntProperty("_ORIGIN_POSITION_")
					if alpha > 0 { // uplift
						if i == 0 {
							delta *= math.E
						} else {
							v := item.Score / maxScore // 归一化 rank score
							idx := int(v * 1000)
							if idx < 0 {
								idx = 0
							}
							if idx >= 1000 {
								idx = 999
							}
							delta *= expTable[idx]
						}
					}
					deltaRank += delta // 多个目标调控方向不一致时，需要扳手腕看谁力气大
					if delta != 0.0 {
						if hitTargets == nil {
							hitTargets = make(map[string][]string)
						}
						taskId := controller.task.TrafficControlTaskId
						hitTargets[taskId] = append(hitTargets[taskId], targetId)
					}
					if ctx.Debug {
						ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tmicro control\t%s\titem_id[%s], origin_pos=%d, aim_value=%f, alpha=%f, delta_rank=%f", trafficIdentity(controller.task.TrafficControlTaskId, targetId, controller.target.Name), item.Id, originPosition, aimValue, alpha, delta))
					}
				}
				if deltaRank != 0.0 {
					if ctx.Debug {
						ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tmicro control\titem_id[%v]\tfinal delta rank:%v", item.Id, deltaRank))
					}
					// only record when the delta rank is really applied to the item, the uplift
					// branch can be skipped once the limit count is reached
					if deltaRank < 0 {
						item.IncrAlgoScore("__delta_rank__", deltaRank)
						if len(hitTargets) > 0 {
							item.AddProperty(microTrafficControlHitsProperty, hitTargets)
						}
					} else if upliftCount < limitCount { // uplift
						item.IncrAlgoScore("__delta_rank__", deltaRank)
						countMu.Lock()
						upliftCount++
						countMu.Unlock()
						pos, _ := item.IntProperty("_ORIGIN_POSITION_")
						if pos > ctx.Size {
							item.AddProperty("__traffic_control_id__", 0)
						}
						if len(hitTargets) > 0 {
							item.AddProperty(microTrafficControlHitsProperty, hitTargets)
						}
					}
				}
			}
			controlNumberCh <- localControlNumberMap
			<-sem
		}(begin, candidates)
	}
	innerWg.Wait()
	close(controlNumberCh)
	for localControlNumberMap := range controlNumberCh {
		for targetId, count := range localControlNumberMap {
			controlNumberMap[targetId] += count
		}
	}
	ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\tmicro control\tcount=%d\tcost=%d\tcontrolled item num=%v", len(items), utils.CostTime(start), controlNumberMap))
}

// 获取线上真实的曝光流量
func getItemActualTraffic(controllerMap map[string]*PIDController, actualTrafficsMap map[string][]*experiments.TrafficControlTargetTraffic, controllerParamsMap map[string]controllerParams) map[string][]*experiments.TrafficControlTargetTraffic {
	for targetId, traffics := range actualTrafficsMap {
		controller, ok := controllerMap[targetId]
		if ok {
			params := controllerParamsMap[targetId]
			for _, traffic := range traffics {
				controller.setMeasurement("", traffic.ItemOrExpId, traffic.TargetTraffic, traffic.RecordTime, params)
			}
		}
	}
	return actualTrafficsMap
}

// trafficControlTraffic 是一次请求的只读流量视图，不持有额外的跨请求缓存。
type trafficControlTraffic struct {
	All        map[string]*experiments.TrafficControlTargetTraffic
	Experiment map[string]*experiments.TrafficControlTargetTraffic
	Items      map[string][]*experiments.TrafficControlTargetTraffic
	HasTraffic bool
}

// selectTrafficControlTraffic 只查找请求所需明细，避免 SDK 展开全部 target×item 数据。
func selectTrafficControlTraffic(tasks []*model.TrafficControlTask, env, expId string, valid, selected map[string]*PIDController, items []*module.Item) trafficControlTraffic {
	result := trafficControlTraffic{
		All:        make(map[string]*experiments.TrafficControlTargetTraffic),
		Experiment: make(map[string]*experiments.TrafficControlTargetTraffic),
		Items:      make(map[string][]*experiments.TrafficControlTargetTraffic),
	}
	if env != constants.RunEnvironmentOfProduction && env != constants.RunEnvironmentOfPrepub {
		return result
	}
	needMacro, needMicro := false, false
	for _, controller := range selected {
		if controller.task.ControlGranularity == constants.TrafficControlTaskControlGranularitySingle {
			needMicro = true
		} else {
			needMacro = true
		}
	}
	itemIds := make(map[string]struct{})
	if needMicro {
		for _, item := range items {
			itemIds[string(item.Id)] = struct{}{}
		}
	}
	neededTasks := make(map[string]struct{})
	for id, controller := range valid {
		if controller.task.ControlGranularity == constants.TrafficControlTaskControlGranularitySingle {
			if selected[id] == nil {
				continue
			}
		} else if !needMacro {
			continue
		}
		neededTasks[controller.task.TrafficControlTaskId] = struct{}{}
	}
	for _, task := range tasks {
		if task == nil {
			continue
		}
		if _, ok := neededTasks[task.TrafficControlTaskId]; !ok {
			continue
		}
		// 仅使用 SDK 返回的有效 target，不能把残留的流量明细当作有效目标。
		needed := make(map[string]*PIDController)
		for _, target := range task.TrafficControlTargets {
			if target == nil {
				continue
			}
			id := target.TrafficControlTargetId
			controller := valid[id]
			if controller == nil || controller.task.TrafficControlTaskId != task.TrafficControlTaskId {
				continue
			}
			if controller.task.ControlGranularity == constants.TrafficControlTaskControlGranularitySingle {
				if selected[id] == nil {
					continue
				}
			} else if !needMacro {
				continue
			}
			needed[id] = controller
		}
		for _, targetTraffic := range task.ActualTraffic.TargetTraffics {
			if targetTraffic == nil {
				continue
			}
			id := targetTraffic.TrafficControlTargetId
			controller := needed[id]
			if controller == nil {
				continue
			}
			// 与 SDK 的线性查找一致，同一 target 只使用第一条明细集合。
			delete(needed, id)
			lookup := func(key string) *experiments.TrafficControlTargetTraffic {
				detail, ok := targetTraffic.Data[key]
				if !ok {
					return nil
				}
				return &experiments.TrafficControlTargetTraffic{
					ItemOrExpId: key, TrafficControlTaskId: task.TrafficControlTaskId,
					TrafficControlTargetId: id, TargetTraffic: detail.Traffic,
					TaskTraffic: task.ActualTraffic.TaskTraffics[key].Traffic,
					RecordTime:  time.Unix(detail.RecordTime, 0),
				}
			}
			if controller.task.ControlGranularity == constants.TrafficControlTaskControlGranularitySingle {
				for itemId := range itemIds {
					if record := lookup(itemId); record != nil {
						result.Items[id] = append(result.Items[id], record)
					}
				}
			} else {
				if record := lookup("ER_ALL"); record != nil {
					result.All[id] = record
				}
				if expId != "ER_ALL" {
					if record := lookup(expId); record != nil {
						result.Experiment[id] = record
					}
				}
			}
		}
	}
	// 健康判断覆盖完整有效宏观目标，不让随机抽到零流量样本变成全局故障。
	if needMacro {
		for id, controller := range valid {
			if controller.task.ControlGranularity == constants.TrafficControlTaskControlGranularitySingle {
				continue
			}
			record := result.All[id]
			if controller.task.ControlType == constants.TrafficControlTaskControlTypePercent && controller.IsAllocateExpWise() {
				record = result.Experiment[id]
				measurement := 0.0
				if record != nil {
					measurement = record.TargetTraffic
				}
				if measurement < controller.GetMinExpTraffic() {
					record = result.All[id]
				}
			}
			if record != nil && record.TargetTraffic > 0 {
				result.HasTraffic = true
				break
			}
		}
	}
	return result
}

func (p *TrafficControlSort) loadTrafficControllersMap() map[string]*PIDController {
	runEnv := os.Getenv("PAIREC_ENVIRONMENT")

	tasks := experimentClient.ListTrafficControlTasks(runEnv)
	if len(tasks) == 0 {
		controllerMap := make(map[string]*PIDController)
		p.controllerLock.Lock()
		p.controllersMap = controllerMap
		p.controllerLock.Unlock()
		log.Info(fmt.Sprintf("module=TrafficControlSort\tthere are no running tasks."))
		return controllerMap
	}
	oldControllerMap := make(map[string]*PIDController, 0)
	p.controllerLock.RLock()
	for targetId, controller := range p.controllersMap {
		oldControllerMap[targetId] = controller
	}
	p.controllerLock.RUnlock()

	controllerMap := make(map[string]*PIDController, 0)
	for i, task := range tasks {
		for _, value := range task.TrafficControlTargets {
			target := *value
			if oldControllerMap != nil {
				pidController, ok := oldControllerMap[target.TrafficControlTargetId]
				if ok {
					// update meta info
					pidController.task = tasks[i]
					pidController.target = &target
					controllerMap[target.TrafficControlTargetId] = pidController
					continue
				}
			}
			controller := NewPIDController(tasks[i], &target, p.config)
			if controller != nil {
				controllerMap[target.TrafficControlTargetId] = controller
			}
		}
	}

	p.controllerLock.Lock()
	p.controllersMap = controllerMap
	p.controllerLock.Unlock()
	log.Info(fmt.Sprintf("module=TrafficControlSort\tload %d traffic control target.", len(controllerMap)))
	return controllerMap
}

// 过滤有效的调控任务
func filterValidControllers(ctx *context.RecommendContext, user *module.User, superParam *ExperimentParams, controllerMap map[string]*PIDController) map[string]*PIDController {
	resultControllersMap := make(map[string]*PIDController)

	tmpSceneControllersMap := isControlledByScene(ctx, controllerMap)

	tmpUserControllerMap := isControlledByUser(user, tmpSceneControllersMap)

	tmpParamControllersMap := isControlledByParams(ctx, superParam, tmpUserControllerMap)

	resultControllersMap = isControlledByExperimentId(ctx, tmpParamControllersMap)
	return resultControllersMap
}

// 通过当前场景过滤调控任务
func isControlledByScene(ctx *context.RecommendContext, controllerMap map[string]*PIDController) map[string]*PIDController {
	newControllerMap := make(map[string]*PIDController, 0)
	scene := ctx.GetParameter("scene").(string)
	for targetId, controller := range controllerMap {
		// 当前场景是否和主场景匹配
		if controller.task.SceneName == scene {
			newControllerMap[targetId] = controller
			continue
		}
		// 当前场景是否在有效场景列表中
		effectiveScenes := controller.task.EffectiveSceneNames
		exist := isItemInArray(scene, effectiveScenes)
		if exist {
			newControllerMap[targetId] = controller
		}
	}
	return newControllerMap
}

// 通过当前用户的属性过滤调控任务
func isControlledByUser(user *module.User, controllerMap map[string]*PIDController) map[string]*PIDController {
	controllerNewMap := make(map[string]*PIDController)
	var properties map[string]interface{}
	var machine vm.VM

	for targetId, controller := range controllerMap {
		if controller.userExprProg == nil {
			controllerNewMap[targetId] = controller
			continue
		}

		if properties == nil {
			properties = user.MakeUserFeatures2()
		}
		result, err := machine.Run(controller.userExprProg, properties)
		if err != nil {
			log.Warning(fmt.Sprintf("module=PIDController\tcompute user expression field, err:%v", err))
			return controllerNewMap
		}
		if result.(bool) { // 这里一定是 bool
			controllerNewMap[targetId] = controller
		}
	}
	return controllerNewMap
}

// 通过实验参数过滤调控任务
func isControlledByParams(ctx *context.RecommendContext, superParams *ExperimentParams, controllerMap map[string]*PIDController) map[string]*PIDController {
	newControllerMap := make(map[string]*PIDController)

	// 判断页数
	if superParams.StartPageNumber != nil {
		pageNumber := utils.ToInt(ctx.GetParameter("page_number"), 1)
		if pageNumber <= *superParams.StartPageNumber {
			return newControllerMap
		}
	}

	for targetId, controller := range controllerMap {
		if taskParams, ok := superParams.PidTaskParams[controller.task.TrafficControlTaskId]; ok && taskParams.TurnOff != nil && *taskParams.TurnOff {
			continue
		}
		if targetParams, ok := superParams.PidTargetParams[controller.target.TrafficControlTargetId]; ok && targetParams.TurnOff != nil && *targetParams.TurnOff {
			continue
		}
		newControllerMap[targetId] = controller
	}
	return newControllerMap
}

// 通过当前实验ID 过滤调控任务
func isControlledByExperimentId(ctx *context.RecommendContext, controllerMap map[string]*PIDController) map[string]*PIDController {
	newControllerMap := make(map[string]*PIDController)
	runEnv := os.Getenv("PAIREC_ENVIRONMENT")
	for targetId, controller := range controllerMap {
		expId := ctx.ExperimentResult.GetExpId()
		var taskExpIds string
		if runEnv == constants.RunEnvironmentOfProduction {
			taskExpIds = controller.task.ProdExperimentIds
		} else {
			taskExpIds = controller.task.PreExperimentIds
		}

		// 当调控任务绑定的实验ID为空，则表示该调控任务不需要过滤，对所有的实验都是生效的
		if taskExpIds == "" {
			newControllerMap[targetId] = controller
			continue
		}

		expIdArr := strings.Split(taskExpIds, ",")
		for _, value := range expIdArr {
			subStr := fmt.Sprintf("E%s", value)
			exist := strings.Contains(expId, subStr)
			if exist {
				newControllerMap[targetId] = controller
			}
		}
	}
	return newControllerMap
}

// 单品调控时，通过 item 属性过滤调控物品集
func isControlledItem(controller *PIDController, item *module.Item) bool {
	return isControlledItemFeatures(controller, item.Id, item.GetFeatures())
}

func isControlledItemFeatures(controller *PIDController, itemId module.ItemId, properties map[string]interface{}) bool {
	var machine vm.VM
	return isControlledItemFeaturesWithVM(controller, itemId, properties, &machine)
}

func isControlledItemFeaturesWithVM(controller *PIDController, itemId module.ItemId, properties map[string]interface{}, machine *vm.VM) bool {
	if controller.itemExprProg == nil {
		// 物品表达式未编译成功（未配置条件或解析/编译失败）时拦截所有物品，
		// 避免错误配置把全量物品当成调控目标；警告日志每个控制器只打一次。
		controller.itemExprNilWarnOnce.Do(func() {
			log.Warning(fmt.Sprintf("module=TrafficControlSort\titem expression is not compiled, all items are excluded\t%s", trafficIdentity(controller.task.TrafficControlTaskId, controller.target.TrafficControlTargetId, controller.target.Name)))
		})
		return false
	}
	result, err := machine.Run(controller.itemExprProg, properties)
	if err != nil {
		log.Warning(fmt.Sprintf("module=TrafficControlSort\titem_id:%v\tcompute item expression field, err:%v", itemId, err))
		return false
	}

	return ToBool(result, false)
}

// lessTrafficControlId orders ids numerically when both of them are numeric, so the
// rendered output keeps a natural order, and falls back to a lexicographic order to
// stay deterministic for any other id format.
func lessTrafficControlId(a, b string) bool {
	ai, aErr := strconv.ParseInt(a, 10, 64)
	bi, bErr := strconv.ParseInt(b, 10, 64)
	if aErr == nil && bErr == nil {
		return ai < bi
	}
	return a < b
}

// formatTrafficControlHits renders the hit tasks and their targets as
// "taskId1:targetId1,targetId2|taskId2:targetId3". Both task ids and target ids are
// ordered so that the output stays stable across requests.
func formatTrafficControlHits(hits map[string][]string) string {
	if len(hits) == 0 {
		return ""
	}
	taskIds := make([]string, 0, len(hits))
	for taskId := range hits {
		taskIds = append(taskIds, taskId)
	}
	sort.Slice(taskIds, func(i, j int) bool { return lessTrafficControlId(taskIds[i], taskIds[j]) })

	var sb strings.Builder
	for i, taskId := range taskIds {
		if i > 0 {
			sb.WriteByte('|')
		}
		sb.WriteString(taskId)
		sb.WriteByte(':')
		targetIds := hits[taskId]
		sort.Slice(targetIds, func(i, j int) bool { return lessTrafficControlId(targetIds[i], targetIds[j]) })
		for j, targetId := range targetIds {
			if j > 0 {
				sb.WriteByte(',')
			}
			sb.WriteString(targetId)
		}
	}
	return sb.String()
}

// mergeTrafficControlHits merges the hits collected by macroControl and microControl
// into the final output property. It must be called after both controls finished, so it
// runs free of their concurrent writes.
func mergeTrafficControlHits(item *module.Item) {
	macroHits, _ := item.GetProperty(macroTrafficControlHitsProperty).(map[string][]string)
	microHits, _ := item.GetProperty(microTrafficControlHitsProperty).(map[string][]string)
	if len(macroHits) == 0 && len(microHits) == 0 {
		return
	}
	item.DeleteProperty(macroTrafficControlHitsProperty)
	item.DeleteProperty(microTrafficControlHitsProperty)

	hits := make(map[string][]string, len(macroHits)+len(microHits))
	for _, source := range []map[string][]string{macroHits, microHits} {
		for taskId, targetIds := range source {
			for _, targetId := range targetIds {
				if !isItemInArray(targetId, hits[taskId]) {
					hits[taskId] = append(hits[taskId], targetId)
				}
			}
		}
	}
	item.AddProperty(TrafficControlHitsProperty, formatTrafficControlHits(hits))
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

type macroAlphaJob struct {
	TargetId   string
	Controller *PIDController
	Params     controllerParams
}

type targetAlpha struct {
	TargetId string
	Alpha    float64
}

type ExperimentParams struct {
	PidTargetSampleCount         *int     `json:"pid_target_sample_count,omitempty"` // 单请求所有调控目标共享的抽样数量，默认50
	StartPageNumber              *int     `json:"start_page_number,omitempty"`
	MinExpTraffic                *float64 `json:"min_exp_traffic,omitempty"`
	LimitCountOnFirstPage        *int     `json:"limit_count_on_first_page,omitempty"`
	CandidateCountAfterFirstPage *int     `json:"candidate_count_after_first_page,omitempty"` // 启用此参数，第一页不会放调控的item，会排在第一页后，如果有其他重排，可能会被调控到第一页
	PidSingleControlLimitCount   *int     `json:"pid_single_control_limit_count,omitempty"`   // 单品调控和按量全局调控分别适用的单次上提物品数上限，默认为5
	PidMaxUpliftItemCnt          *int     `json:"pid_max_uplift_item_cnt,omitempty"`

	PidGamma             *float64             `json:"pid_gamma,omitempty"`
	PidBeta              *float64             `json:"pid_beta,omitempty"`
	PidEta               *float64             `json:"pid_eta,omitempty"`
	PidErrDiscount       *float64             `json:"pid_err_discount,omitempty"`
	PidIntegralThreshold *float64             `json:"pid_integral_threshold,omitempty"`
	PidTaskParams        map[string]PidParams `json:"pid_task_params,omitempty"`   // key: task_id
	PidTargetParams      map[string]PidParams `json:"pid_target_params,omitempty"` // key: target_id
}

type PidParams struct {
	TurnOff           *bool    `json:"turn_off,omitempty"`
	Kp                *float64 `json:"kp,omitempty"`
	Ki                *float64 `json:"ki,omitempty"`
	Kd                *float64 `json:"kd,omitempty"`
	IntegralThreshold *float64 `json:"integral_threshold,omitempty"`
	ErrThreshold      *float64 `json:"err_threshold,omitempty"`
	AllocateExpWise   *bool    `json:"allocate_exp_wise,omitempty"`
}

func newExperimentParams(layerParams model.LayerParams) *ExperimentParams {
	superParams := &ExperimentParams{}
	params := layerParams.ListParams()
	data, err := json.Marshal(params)
	if err != nil {
		log.Error(fmt.Sprintf("module=TrafficControlSort\tevent=NewSuperParams\terr=%v", err))
		return superParams
	}

	err = json.Unmarshal(data, superParams)
	if err != nil {
		log.Error(fmt.Sprintf("module=TrafficControlSort\tevent=NewSuperParams\terr=%v", err))
		return superParams
	}

	return superParams
}

type controllerParams struct {
	Kp                   *float64
	Ki                   *float64
	Kd                   *float64
	PidErrDiscount       *float64
	PidIntegralThreshold *float64
}

func precomputeControllerParams(controllerMap map[string]*PIDController, experimentParams *ExperimentParams) map[string]controllerParams {
	paramsMap := make(map[string]controllerParams, len(controllerMap))
	for targetId, controller := range controllerMap {
		paramsMap[targetId] = parseControllerParams(controller.task.TrafficControlTaskId, targetId, experimentParams)
	}
	return paramsMap
}

func parseControllerParams(taskId, targetId string, experimentParams *ExperimentParams) controllerParams {
	params := controllerParams{}

	// 先取全局
	if experimentParams.PidErrDiscount != nil {
		params.PidErrDiscount = experimentParams.PidErrDiscount
	}
	if experimentParams.PidIntegralThreshold != nil {
		params.PidIntegralThreshold = experimentParams.PidIntegralThreshold
	}

	// 再取 task 的实验参数
	if experimentParams.PidTaskParams != nil {
		if taskParams, ok := experimentParams.PidTaskParams[taskId]; ok {
			params.Kp = taskParams.Kp
			params.Ki = taskParams.Ki
			params.Kd = taskParams.Kd
		}
	}
	// 再用更细粒度的 target 参数覆盖 task 参数
	if experimentParams.PidTargetParams != nil {
		if targetParams, ok := experimentParams.PidTargetParams[targetId]; ok {
			if targetParams.Kp != nil {
				params.Kp = targetParams.Kp
			}
			if targetParams.Ki != nil {
				params.Ki = targetParams.Ki
			}
			if targetParams.Kd != nil {
				params.Kd = targetParams.Kd
			}
		}
	}
	return params
}

func ToBool(i interface{}, defaultVal bool) bool {
	switch value := i.(type) {
	case bool:
		return value
	case string:
		if "true" == strings.ToLower(value) || "y" == strings.ToLower(value) {
			return true
		}
	default:
		return defaultVal
	}
	return defaultVal
}

func isItemInArray(element string, array []string) bool {
	for _, value := range array {
		if element == value {
			return true
		}
	}
	return false
}

// trafficIdentity 渲染流量调控日志的统一身份标记，如 [task=2|target=3|name="string_in"]，
// name 加引号避免目标名中含 = 或空格时破坏解析。
func trafficIdentity(taskId, targetId, targetName string) string {
	return fmt.Sprintf("[task=%s|target=%s|name='%s']", taskId, targetId, targetName)
}
