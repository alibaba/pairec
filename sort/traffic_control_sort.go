package sort

import (
	gocontext "context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/abtest"
	"github.com/alibaba/pairec/v2/constants"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
	"github.com/expr-lang/expr"

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
	items, good := sortData.Data.([]*module.Item)
	if !good {
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

	layerParams := ctx.ExperimentResult.GetExperimentParams()
	experimentParams := newExperimentParams(layerParams)
	if ctx.Debug {
		d, _ := json.Marshal(experimentParams)
		ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\texperiment params: %s", string(d)))
	}

	sort.Sort(sort.Reverse(ItemScoreSlice(items)))
	for i, item := range items {
		item.AddProperty("__traffic_control_id__", i+1)
		item.AddProperty("_ORIGIN_POSITION_", i+1)
	}

	// 如果服务启动时，没有加载成功，这里再次尝试
	var allControllersMap map[string]*PIDController
	if len(p.controllersMap) == 0 {
		allControllersMap = p.loadTrafficControllersMap()
	} else {
		allControllersMap = p.controllersMap
	}

	validControllersMap := filterValidControllers(ctx, user, experimentParams, allControllersMap)

	globalControls, singleControls := splitControllers(validControllersMap)
	if len(globalControls) == 0 && len(singleControls) == 0 {
		ctx.LogWarning(fmt.Sprintf("module=TrafficControlSort\tboth global traffic control and single traffic control are zero"))
		sortData.Data = items
		return nil
	}

	ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tglobal control num %d, single control num %d", len(globalControls), len(singleControls)))
	wgCtrl := sync.WaitGroup{}
	if len(singleControls) > 0 {
		wgCtrl.Add(1)
		go microControl(ctx, singleControls, items, &wgCtrl, experimentParams)
	}
	if len(globalControls) > 0 {
		wgCtrl.Add(1)
		go macroControl(ctx, globalControls, items, &wgCtrl, experimentParams)
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
	ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\tcount=%d\tcost=%d", len(items), utils.CostTime(start)))
	sortData.Data = items
	return nil
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

// 宏观调控，针对目标整体
func macroControl(ctx *context.RecommendContext, controllerMap map[string]*PIDController, items []*module.Item, wgCtrl *sync.WaitGroup, experimentParams *ExperimentParams) {
	defer wgCtrl.Done()
	begin := time.Now()
	targetAlphaMap := computeMacroControlAlpha(ctx, controllerMap, experimentParams)
	if len(targetAlphaMap) == 0 {
		ctx.LogWarning(fmt.Sprintf("module=TrafficControlSort\tmacro control\ttraffic control task output is zero"))
		return
	}
	if ctx.Debug {
		d, _ := json.Marshal(targetAlphaMap)
		ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tmacro control\tthe relationship between target and alpha: %s", string(d)))
	}

	itemScores := make([]float64, len(items))
	// 计算各个目标的偏好分的全局占比
	totalScore := 0.0
	maxScore := 0.0 // item 列表中的最大分
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
			isControlled := isControlledItem(controller, item)
			if alpha, ok := targetAlphaMap[targetId]; ok && alpha != 0 && isControlled {
				targetScore[targetId] += score
			}
		}
	}
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
	ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tmacro control\ttarget alpha: %v", targetAlphaMap))
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

	targetControlledNum := make(map[string]int, len(controllerMap))
	mu := sync.RWMutex{}

	// compute delta rank
	parallel := 10
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
				for targetId, controller := range controllerMap {
					if !isControlledItem(controller, item) {
						if ctx.Debug {
							mu.Lock()
							targetControlledNum[targetId] = targetControlledNum[targetId] - 1
							mu.Unlock()
						}
						continue
					}
					if alpha, ok := targetAlphaMap[targetId]; ok && alpha != 0 {
						deltaRank := computeDeltaRank(controller, item, i, alpha, ctrlParams, ctx)
						finalDeltaRank += deltaRank // 形成合力
					}
				}

				if finalDeltaRank != 0.0 {
					item.IncrAlgoScore("__delta_rank__", finalDeltaRank)
					controlId, _ := item.IntProperty("__traffic_control_id__")
					if controlId == 0 && finalDeltaRank < 1.0 {
						item.AddProperty("__traffic_control_id__", item.GetProperty("_ORIGIN_POSITION_"))
					}
				}
			}
			<-ch
		}(b, candidates)
	}
	wg.Wait()
	ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\tmacro control\tcount=%d\tcost=%d\tcontrolled target=%v",
		len(items), utils.CostTime(begin), targetControlledNum))
}

// FlowControl 非单品（整体）目标流量调控，返回各个目标的调控力度
func computeMacroControlAlpha(ctx *context.RecommendContext, controllerMap map[string]*PIDController, experimentParams *ExperimentParams) map[string]float64 {
	// 获取ControlGranularity="Global"(调控粒度是全局)类型的调控目标 当前已累计完成的流量
	targetAlphaMap := make(map[string]float64)

	// 获取流量实时统计值
	runEnv := os.Getenv("PAIREC_ENVIRONMENT")
	expId := ctx.ExperimentResult.GetExpId()
	actualTargetTrafficsMap := experimentClient.GetTrafficControlActualTraffic(runEnv, expId, "ER_ALL")
	if ctx.Debug {
		data, _ := json.Marshal(actualTargetTrafficsMap)
		ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tactual traffics:%s", string(data)))
	}
	allTrafficMap := make(map[string]*experiments.TrafficControlTargetTraffic)
	expTrafficMap := make(map[string]*experiments.TrafficControlTargetTraffic)
	for _, targetTraffics := range actualTargetTrafficsMap {
		for _, trafficInfo := range targetTraffics {
			if trafficInfo.ItemOrExpId == "ER_ALL" {
				allTrafficMap[trafficInfo.TrafficControlTargetId] = trafficInfo
			} else if trafficInfo.ItemOrExpId == expId {
				expTrafficMap[trafficInfo.TrafficControlTargetId] = trafficInfo
			}
		}
	}
	hasTraffic := false
	retCh := make(chan targetAlpha, utils.MinInt(len(controllerMap), 64))
	defer close(retCh)

	gCtx, cancel := gocontext.WithTimeout(gocontext.Background(), time.Millisecond*12)
	defer cancel()
	for targetId, controller := range controllerMap {
		go func(gCtx gocontext.Context, targetId string, controller *PIDController, experimentParams *ExperimentParams) {
			defer func() {
				if err := recover(); err != nil {
					log.Warning(fmt.Sprintf("traffic control timeout in background: <taskId:%s/targetId:%s>[targetName:%s]",
						controller.task.TrafficControlTaskId, targetId, controller.target.Name))
				}
			}()
			taskId := controller.task.TrafficControlTaskId
			var targetTraffic, taskTraffic, alpha, aimValue float64
			var experimentId = ""
			var targetTrafficMap map[string]*experiments.TrafficControlTargetTraffic
			var measureTime time.Time
			if controller.task.ControlType == constants.TrafficControlTaskControlTypePercent {
				if controller.IsAllocateExpWise() {
					targetTrafficMap = expTrafficMap
					experimentId = expId
				} else {
					targetTrafficMap = allTrafficMap
				}
				if trafficInfo, ok := targetTrafficMap[targetId]; ok {
					targetTraffic = trafficInfo.TargetTraffic
					taskTraffic = trafficInfo.TaskTraffic
					measureTime = trafficInfo.RecordTime
				} else {
					targetTraffic = float64(0)
					taskTraffic = float64(1)
					measureTime = time.Now().Truncate(time.Second)
				}
				if controller.IsAllocateExpWise() && targetTraffic < controller.GetMinExpTraffic() {
					// 用全局流量代替冷启动的实验流量
					ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\t<taskId:%s/targetId:%s>[targetName:%s]\texp=%s\ttargetTraffic=%.0f change to global targetTraffic", taskId, targetId, controller.target.Name, expId, targetTraffic))
					experimentId = ""
					if trafficInfo, ok := allTrafficMap[targetId]; ok {
						targetTraffic = trafficInfo.TargetTraffic
						taskTraffic = trafficInfo.TaskTraffic
						measureTime = trafficInfo.RecordTime
					} else {
						targetTraffic = float64(0)
						taskTraffic = float64(1)
						measureTime = time.Now().Truncate(time.Second)
					}
				}
				params := parseControllerParams(controller.task.TrafficControlTaskId, targetId, experimentParams)
				trafficPercentage := targetTraffic / taskTraffic
				controller.setMeasurement(experimentId, "", trafficPercentage, measureTime, params)
				alpha, aimValue = controller.compute(ctx, experimentId, "", params)
				ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\t<taskId:%s/targetId:%s>[targetName:%s]\ttrafficInfo=%.0f, percentage=%f, controlAimValue=%f, alpha=%f, exp=%s", taskId, targetId, controller.target.Name, targetTraffic, trafficPercentage, aimValue/100, alpha, experimentId))
				if targetTraffic > 0 {
					hasTraffic = true
				}
			} else {
				if trafficInfo, ok := allTrafficMap[targetId]; ok {
					targetTraffic = trafficInfo.TargetTraffic
					measureTime = trafficInfo.RecordTime
				} else {
					targetTraffic = float64(0)
					measureTime = time.Now().Truncate(time.Second)
				}
				params := parseControllerParams(controller.task.TrafficControlTaskId, targetId, experimentParams)
				controller.setMeasurement("", "", targetTraffic, measureTime, params)
				alpha, aimValue = controller.compute(ctx, "", "", params)
				ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\t<taskId:%s/targetId:%s>[targetName:%s]\t"+
					"traffic=%.0f, controlAimValue=%f, alpha=%f",
					taskId, targetId, controller.target.Name, targetTraffic, aimValue, alpha))
				if targetTraffic > 0 {
					hasTraffic = true
				}
			}
			select {
			case <-gCtx.Done():
				ctx.LogWarning(fmt.Sprintf("module=TrafficControlSort\ttimeout in goruntine: <taskId:%s/targetId:%s>[targetName:%s]",
					taskId, targetId, controller.target.Name))
				return
			case retCh <- targetAlpha{
				TargetId: targetId,
				Alpha:    alpha,
			}:
			}
		}(gCtx, targetId, controller, experimentParams)
	}
Loop:
	for range controllerMap {
		select {
		case pair := <-retCh:
			alpha := pair.Alpha
			if alpha != 0 {
				targetAlphaMap[pair.TargetId] = alpha
			}
		case <-gCtx.Done():
			if errors.Is(gCtx.Err(), gocontext.DeadlineExceeded) {
				ctx.LogWarning(fmt.Sprintf("module=TrafficControlSort\ttraffic controller timeout: %v", gCtx.Err()))
			}
			break Loop
		}
	}
	if !hasTraffic {
		ctx.LogWarning(fmt.Sprintf("module=TrafficControlSort\tno traffic data detected, maybe flink job is not running"))
		for k := range targetAlphaMap {
			delete(targetAlphaMap, k)
		}
	}
	return targetAlphaMap
}

// SampleControlTargetsByScore 按照偏好分权重选择n个上提目标，未被选中的目标调控值置0
func sampleControlTargetsByScore(ctx *context.RecommendContext, maxUpliftTargetCnt int, targetScore, targetAlpha map[string]float64) {
	if maxUpliftTargetCnt >= len(targetScore) || maxUpliftTargetCnt <= 0 {
		return
	}
	targetIds := make([]string, len(targetScore))
	scores := make([]float64, len(targetScore))
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
	ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tsample control target\tselected uplift target ids %v score", selected))
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
		ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tcompute delta itemIndex\titem %s [%s/%s], "+
			"score proportion=%.3f, rho=%.3f, alpha=%.6f, origin pos=%d, delta itemIndex=%.6f [pull down]",
			item.Id, c.target.TrafficControlTargetId, c.target.Name, scoreWeight, rho, alpha, itemIndex+1, deltaRank))
	} else { // uplift
		deltaRank *= itemScore // item.Score 越大，提权越多；用来在不同提取目标间竞争
		distinctStartPos := ctx.Size
		if args.pageNo > 1 {
			multiple := (scoreWeight - 0.3) * 10
			distinctStartPos += int(multiple * float64(ctx.Size))
		}
		if itemIndex > distinctStartPos && deltaRank >= 1.0 {
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
		controlId, _ := item.IntProperty("__traffic_control_id__")
		ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tcompute delta itemIndex\titem:%s\t[targetId:%s/targetName:%s],"+
			"score proportion=%.3f,norm_score=%.3f, alpha=%.6f, origin pos=%d, delta itemIndex=%.6f, traffic_control_id=%d [uplift]",
			item.Id, c.target.TrafficControlTargetId, c.target.Name, scoreWeight, itemScore, alpha, itemIndex+1, deltaRank, controlId))
	}
	return deltaRank
}

// 微观调控，针对单个item
func microControl(ctx *context.RecommendContext, controllerMap map[string]*PIDController, items []*module.Item, wg *sync.WaitGroup, experimentParams *ExperimentParams) {
	defer wg.Done()
	targetItemActualTrafficMap := getItemActualTraffic(controllerMap, items, experimentParams) // key1: targetId
	if ctx.Debug {
		data, _ := json.Marshal(targetItemActualTrafficMap)
		ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tevent=microControl\titem traffic:%s", string(data)))
	}
	var limitCount int
	if experimentParams.PidSingleControlLimitCount != nil {
		limitCount = *experimentParams.PidSingleControlLimitCount
	} else {
		limitCount = 5
	}
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

	//Calculate the number of items controlled by each target, key:targetId; value:controlled item sum
	var mapMu sync.Mutex
	controlNumberMap := make(map[string]int)

	var countMu sync.Mutex
	upliftCount := 0

	parallel := 10
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
			for j, item := range items {
				i := begin + j
				deltaRank := 0.0
				for targetId, controller := range controllerMap {
					if !isControlledItem(controller, item) {
						ctx.LogDebug(fmt.Sprintf("item id:%v is not controller", item.Id))
						continue
					}
					mapMu.Lock()
					controlNumberMap[targetId] = controlNumberMap[targetId] + 1
					mapMu.Unlock()

					params := parseControllerParams(controller.task.TrafficControlTaskId, targetId, experimentParams)
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
					if ctx.Debug {
						ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\t[targetId:%s/targetName:%s], itemId:%s,origiPosition=%d, aimValue=%f, alpha=%f, deltaRank=%f", targetId, controller.target.Name, item.Id, originPosition, aimValue, alpha, delta))
					}
				}
				if deltaRank != 0.0 {
					if ctx.Debug {
						ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\titem:%v\tfinal delta rank:%v", item.Id, deltaRank))
					}
					if deltaRank < 0 {
						item.IncrAlgoScore("__delta_rank__", deltaRank)
					} else if upliftCount < limitCount { // uplift
						item.IncrAlgoScore("__delta_rank__", deltaRank)
						countMu.Lock()
						upliftCount++
						countMu.Unlock()
						pos, _ := item.IntProperty("_ORIGIN_POSITION_")
						if pos > ctx.Size {
							item.AddProperty("__traffic_control_id__", 0)
						}
					}
				}
			}
			<-sem
		}(begin, candidates)
	}
	innerWg.Wait()
}

// 获取线上真实的曝光流量
func getItemActualTraffic(controllerMap map[string]*PIDController, items []*module.Item, experimentParams *ExperimentParams) map[string][]*experiments.TrafficControlTargetTraffic {
	itemIds := make([]string, len(items))
	for i, item := range items {
		itemIds[i] = string(item.Id)
	}
	runEnv := os.Getenv("PAIREC_ENVIRONMENT")
	actualTrafficsMap := experimentClient.GetTrafficControlActualTraffic(runEnv, itemIds...)

	for targetId, traffics := range actualTrafficsMap {
		controller, ok := controllerMap[targetId]
		if ok {

			params := parseControllerParams(controller.task.TrafficControlTaskId, controller.target.TrafficControlTargetId, experimentParams)

			for _, traffic := range traffics {
				controller.setMeasurement("", traffic.ItemOrExpId, traffic.TargetTraffic, traffic.RecordTime, params)
			}
		}
	}
	return actualTrafficsMap
}

func (p *TrafficControlSort) loadTrafficControllersMap() map[string]*PIDController {
	runEnv := os.Getenv("PAIREC_ENVIRONMENT")

	tasks := experimentClient.ListTrafficControlTasks(runEnv)
	if len(tasks) == 0 {
		log.Info(fmt.Sprintf("module=TrafficControlSort\tthere are no running tasks."))
		return nil
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

	for targetId, controller := range controllerMap {
		if controller.userExprProg == nil {
			controllerNewMap[targetId] = controller
			continue
		}

		properties := user.MakeUserFeatures2()

		result, err := expr.Run(controller.userExprProg, properties)
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
		newControllerMap[targetId] = controller
	}
	for targetId, controller := range controllerMap {
		// 判断 task param
		if superParams.PidTaskParams != nil {
			for taskId, taskParams := range superParams.PidTaskParams {
				if taskId == controller.task.TrafficControlTaskId {
					if taskParams.TurnOff != nil && *taskParams.TurnOff == true {
						delete(newControllerMap, targetId)
						continue
					}
				}
			}
		}
		// 判断 target param
		if superParams.PidTargetParams != nil {
			for paramTargetId, targetParams := range superParams.PidTargetParams {
				if paramTargetId == controller.target.TrafficControlTargetId {
					if targetParams.TurnOff != nil && *targetParams.TurnOff == true {
						delete(newControllerMap, targetId)
					}
				}
			}
		}
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
	if controller.itemExprProg == nil {
		return true
	}
	properties := item.GetFeatures()

	result, err := expr.Run(controller.itemExprProg, properties)
	if err != nil {
		log.Warning(fmt.Sprintf("module=TrafficControlSort\titem_id:%v\tcompute item expression field, err:%v", item.Id, err))
		return false
	}

	return ToBool(result, false)
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

type targetAlpha struct {
	TargetId string
	Alpha    float64
}

type ExperimentParams struct {
	StartPageNumber              *int     `json:"start_page_number,omitempty"`
	MinExpTraffic                *float64 `json:"min_exp_traffic,omitempty"`
	LimitCountOnFirstPage        *int     `json:"limit_count_on_first_page,omitempty"`
	CandidateCountAfterFirstPage *int     `json:"candidate_count_after_first_page,omitempty"` // 启用此参数，第一页不会放调控的item，会排在第一页后，如果有其他重排，可能会被调控到第一页
	PidSingleControlLimitCount   *int     `json:"pid_single_control_limit_count,omitempty"`   // 单品调控时，限制单品调控数量，默认为5
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
