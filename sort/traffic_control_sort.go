package sort

import (
	gocontext "context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/alibaba/pairec/v2/abtest"
	"github.com/alibaba/pairec/v2/constants"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

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
	name            string
	config          *recconf.PIDControllerConfig
	exp2controllers map[string]map[string]*PIDController // key1: expId; key2: targetId
	controllerLock  sync.RWMutex
	cloneInstances  map[string]*TrafficControlSort
	boostScoreSort  *BoostScoreSort
	context         *context.RecommendContext
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
		log.Warning("module=TrafficControlSort\tget experiment client failed.")
	}
	conf := config.PIDConf
	trafficControlSort := TrafficControlSort{
		config:          &conf,
		exp2controllers: make(map[string]map[string]*PIDController),
		name:            config.Name,
		cloneInstances:  make(map[string]*TrafficControlSort),
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
			tmpExpControllers := make(map[string]map[string]*PIDController)

			trafficControlSort.controllerLock.RLock()
			for expId, controllers := range trafficControlSort.exp2controllers {
				tmpExpControllers[expId] = controllers
			}
			trafficControlSort.controllerLock.RUnlock()

			for expId := range tmpExpControllers {
				trafficControlSort.loadTrafficControlTaskMetaData(expId)
			}
			time.Sleep(time.Minute) // 这里需要更新频繁一点，不然web页面上meta信息的修改不能及时反应出来
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

	if p.context == nil {
		p.context = ctx
	}

	params := ctx.ExperimentResult.GetExperimentParams()

	if p.boostScoreSort != nil && params.Get("pid_boost_score", true).(bool) {
		err := p.boostScoreSort.Sort(sortData)
		if err != nil {
			ctx.LogError(fmt.Sprintf("module=TrafficControlSort\tBoostScore\terror=%v", err))
		}
		ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\tBoostScoreSort\tcount=%d\tcost=%d", len(items), utils.CostTime(start)))
	}

	sort.Sort(sort.Reverse(ItemScoreSlice(items)))
	for i, item := range items {
		item.AddProperty("__traffic_control_id__", i+1)
		item.AddProperty("_ORIGIN_POSITION_", i+1)
	}

	controllerInfo := p.getPidControllers(ctx)

	controllerMap := isControlUser(user, controllerInfo)

	globalControls, singleControls := splitController(controllerMap, ctx)
	if len(globalControls) == 0 && len(singleControls) == 0 {
		ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\tcount=%d\tboth global traffic control and single traffic control are zero", len(items)))
		sortData.Data = items
		return nil
	}

	if enable := setHyperParams(controllerMap, ctx); !enable {
		ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\tcount=%d\ttraffic control hyper params turn off", len(items)))
		sortData.Data = items
		return nil
	}
	ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tglobal control num %d, single control num %d", len(globalControls), len(singleControls)))
	wgCtrl := sync.WaitGroup{}
	if len(singleControls) > 0 {
		wgCtrl.Add(1)
		go microControl(singleControls, items, ctx, &wgCtrl)
	}
	if len(globalControls) > 0 {
		wgCtrl.Add(1)
		go macroControl(globalControls, items, ctx, &wgCtrl)
	}
	wgCtrl.Wait()

	pageNo := utils.ToInt(ctx.GetParameter("pageNum"), 1)
	pageSize := ctx.Size // utils.ToInt(ctx.GetParameter("pageSize"), 10)
	if pageNo < 1 {
		pageNo = 1
	}
	limitFirstPage := params.GetInt("limit_uplift_at_first_page", 0)
	ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tlimit_uplift_at_first_page=%d", limitFirstPage))
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

func (p *TrafficControlSort) loadTrafficControlTaskMetaData(expId string) map[string]*PIDController {
	// 调用 SDK 获取调控计划的元信息, 创建 FlowControllers
	runEnv := os.Getenv("PAIREC_ENVIRONMENT")
	timestamp := p.config.Timestamp
	tasks := experimentClient.GetTrafficControlTaskMetaData(runEnv, timestamp)
	if len(tasks) == 0 {
		log.Info(fmt.Sprintf("module=TrafficControlSort\tcurrent timestamp=%d\tnot find traffic control task.", timestamp))
		return nil
	}
	var oldControllerMap map[string]*PIDController
	p.controllerLock.RLock()
	if controllerMap, ok := p.exp2controllers[expId]; ok {
		oldControllerMap = controllerMap
	}
	p.controllerLock.RUnlock()

	controllerMap := make(map[string]*PIDController, 0)
	for i, task := range tasks {
		taskUserExpress, err := ParseExpression(task.UserConditionArray, task.UserConditionExpress)
		if err != nil {
			log.Error(fmt.Sprintf("module=TrafficControlSort\tparse user condition field, please check %s or %s",
				task.UserConditionArray, task.UserConditionExpress))
		}
		for _, value := range task.TrafficControlTargets {
			target := value
			if target.TrafficControlTargetId == "" {
				continue
			}
			if target.Status == constants.TrafficControlTargetStatusClosed {
				continue
			}
			params := experimentClient.GetSceneParams(task.SceneName)
			freeze := params.GetInt(fmt.Sprintf("pid_freeze_target_%s_minutes", target.TrafficControlTargetId), 0)
			run := params.GetString(fmt.Sprintf("pid_run_with_zero_input_%s", task.TrafficControlTaskId), "true")
			runWithZeroInput := strings.ToLower(run) == "true"

			if oldControllerMap != nil {
				pidController, ok := oldControllerMap[target.TrafficControlTargetId]
				if ok {
					// update meta info
					pidController.task = &tasks[i]
					pidController.target = &target
					controllerMap[target.TrafficControlTargetId] = pidController
					if (taskUserExpress) != "" {
						pidController.SetUserExpress(taskUserExpress)
					}
					pidController.GenerateItemExpress()
					pidController.SetFreezeMinutes(freeze)
					pidController.SetRunWithZeroInput(runWithZeroInput)
					continue
				}
			}
			controller := NewPIDController(&tasks[i], &target, p.config, expId)
			if controller != nil {
				if taskUserExpress != "" {
					controller.SetUserExpress(taskUserExpress)
				}
				controller.SetFreezeMinutes(freeze)
				controller.SetRunWithZeroInput(runWithZeroInput)
				controllerMap[target.TrafficControlTargetId] = controller
			}
		}
	}

	p.controllerLock.Lock()
	p.exp2controllers[expId] = controllerMap
	p.controllerLock.Unlock()
	log.Info(fmt.Sprintf("module=TrafficControlSort\tcurrent timestamp=%d\tload %d Traffic Control Task for exp=%s.",
		timestamp, len(controllerMap), expId))
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

	itemIds := make([]string, len(items), len(items))
	for i, item := range items {
		itemIds[i] = string(item.Id)
	}

	// sdk 可能会返回已过期的Target下Item的历史流量，这样的话取最大值就是不对的
	result := make(map[string]map[string]float64) // key1: targetId, key2:expId, value: traffic
	runEnv := os.Getenv("PAIREC_ENVIRONMENT")
	traffics := experimentClient.GetTrafficControlTargetTraffic(runEnv, scene, itemIds...)
	hasTraffic := false
	measureTime := time.Now().Truncate(time.Second)
	for _, traffic := range traffics {
		if ctrl, ok := controllerMap[traffic.TrafficControlTargetId]; ok {
			ctrl.SetMeasurement(traffic.ItemOrExpId, traffic.TargetTraffic, measureTime)
		} else {
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

	return p.loadTrafficControlTaskMetaData(experiment)
}

func splitController(controllers map[string]*PIDController, ctx *context.RecommendContext) (map[string]*PIDController, map[string]*PIDController) {
	wholeCtrls := make(map[string]*PIDController)
	singleCtrls := make(map[string]*PIDController)
	if nil == controllers || len(controllers) == 0 {
		return wholeCtrls, singleCtrls
	}
	for targetId, controller := range controllers {
		if !controller.IsControlledTraffic(ctx) {
			continue
		}
		if controller.task.ControlGranularity == constants.TrafficControlTaskControlGranularitySingle {
			singleCtrls[targetId] = controller
		} else {
			wholeCtrls[targetId] = controller
		}
	}
	return wholeCtrls, singleCtrls
}

// 宏观调控，针对目标整体
func macroControl(controllerMap map[string]*PIDController, items []*module.Item, ctx *context.RecommendContext, wgCtrl *sync.WaitGroup) {
	defer wgCtrl.Done()
	begin := time.Now()
	var targetOutput map[string]float64
	var count int
	targetOutput, count = FlowControl(controllerMap, ctx)
	if len(targetOutput) == 0 || count == 0 {
		ctx.LogWarning(fmt.Sprintf("module=TrafficControlSort\tmacro control\ttraffic control task output is zero"))
		return
	}
	ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tmacro control\ttarget out put: %v", targetOutput))
	begin = time.Now()
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
			if alpha, ok := targetOutput[targetId]; ok && alpha != 0 && controller.IsControlledItem(item) {
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
		SampleControlTargetsByScore(maxUpliftTargetCnt, targetScore, targetOutput, ctx)
	}

	pidGamma := params.GetFloat("pid_gamma", 1.0)
	pidBeta := params.GetFloat("pid_beta", 1.0)
	// preprocess, adjust control signal
	for targetId, alpha := range targetOutput {
		if alpha > 0 { // uplift
			scoreWeight := targetScore[targetId]
			rho := 1.0 + pidGamma*tanh(pidBeta*scoreWeight) // 给更感兴趣的目标更大的提权，用来区分不同的调控目标
			alpha *= rho
			targetOutput[targetId] = alpha
		}
	}
	ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tmacro control\ttarget output: %v", targetOutput))
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

	targetControlledNum := make(map[string]int, len(controllerMap))
	mu := sync.Mutex{}

	for targetId := range controllerMap {
		targetControlledNum[targetId] = len(items)
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
				for targetId, controller := range controllerMap {
					if !controller.IsControlledItem(item) {
						if ctx.Debug {
							mu.Lock()
							targetControlledNum[targetId] = targetControlledNum[targetId] - 1
							mu.Unlock()
						}
						continue
					}
					if alpha, ok := targetOutput[targetId]; ok && alpha != 0 {
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
func FlowControl(controllerMap map[string]*PIDController, ctx *context.RecommendContext) (map[string]float64, int) {
	// 获取(granularity="Global")类型的调控目标 当前已累计完成的流量
	targetOutput := make(map[string]float64)

	var scene string
	var good bool
	s := ctx.GetParameter("scene")
	if scene, good = s.(string); !good {
		ctx.LogError("failed to get scene name")
		return targetOutput, 0
	}
	// 获取流量实时统计值
	runEnv := os.Getenv("PAIREC_ENVIRONMENT")
	expId := ctx.ExperimentResult.GetExpId()
	traffics := experimentClient.GetTrafficControlTargetTraffic(runEnv, scene, expId, "ER_ALL")
	if ctx.Debug {
		data, _ := json.Marshal(traffics)
		ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tflow control\texpId:%s\ttraffics:%s", expId, string(data)))
	}
	allTargetTrafficMap := make(map[string]experiments.TrafficControlTargetTraffic)
	expTargetTrafficMap := make(map[string]experiments.TrafficControlTargetTraffic)
	for _, traffic := range traffics {
		if traffic.ItemOrExpId == "ER_ALL" {
			allTargetTrafficMap[traffic.TrafficControlTargetId] = traffic
		} else {
			expTargetTrafficMap[traffic.TrafficControlTargetId] = traffic
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
					log.Warning(fmt.Sprintf("traffic control timeout in background: <taskId:%s/targetId:%s>[targetName:%s]",
						controller.task.TrafficControlTaskId, targetId, controller.target.Name))
				}
			}()
			taskId := controller.task.TrafficControlTaskId
			var targetTraffic, taskTraffic, output, setValue float64
			var binId = ""
			var trafficMap map[string]experiments.TrafficControlTargetTraffic
			if controller.task.ControlType == constants.TrafficControlTaskControlTypePercent {
				if controller.IsAllocateExpWise() {
					trafficMap = expTargetTrafficMap
					binId = expId
				} else {
					trafficMap = allTargetTrafficMap
				}
				if input, ok := trafficMap[targetId]; ok {
					targetTraffic = input.TargetTraffic
					taskTraffic = input.TaskTraffic
				} else {
					targetTraffic = float64(0)
					taskTraffic = float64(1)
				}
				if controller.IsAllocateExpWise() && targetTraffic < controller.GetMinExpTraffic() {
					// 用全局流量代替冷启动的实验流量
					ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\t<taskId:%s/targetId:%s>[targetName:%s]"+
						"\texp=%s\ttargetTraffic=%.0f change to global targetTraffic",
						taskId, targetId, controller.target.Name, expId, targetTraffic))
					binId = ""
					if input, ok := allTargetTrafficMap[targetId]; ok {
						targetTraffic = input.TargetTraffic
						taskTraffic = input.TaskTraffic
					} else {
						targetTraffic = float64(0)
						taskTraffic = float64(1)
					}
				}

				trafficPercentage := targetTraffic / taskTraffic
				measureTime := time.Now().Truncate(time.Second)
				controller.SetMeasurement(binId, trafficPercentage, measureTime)
				output, setValue = controller.Compute(binId, ctx)
				ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\t<taskId:%s/targetId:%s>[targetName:%s]\t"+
					"traffic=%.0f, percentage=%f, setValue=%f, output=%f, exp=%s", taskId, targetId,
					controller.target.Name, targetTraffic, trafficPercentage, setValue/100, output, binId))
				if targetTraffic > 0 {
					hasTraffic = true
				}
			} else {
				if input, ok := allTargetTrafficMap[targetId]; ok {
					targetTraffic = input.TargetTraffic
				} else {
					targetTraffic = float64(0)
				}
				measureTime := time.Now().Truncate(time.Second)
				controller.SetMeasurement("", targetTraffic, measureTime)
				output, setValue = controller.Compute("", ctx)
				ctx.LogInfo(fmt.Sprintf("module=TrafficControlSort\t<taskId:%s/targetId:%s>[targetName:%s]\t"+
					"traffic=%.0f, setValue=%f, output=%f",
					taskId, targetId, controller.target.Name, targetTraffic, setValue, output))
				if targetTraffic > 0 {
					hasTraffic = true
				}
			}
			select {
			case <-gCtx.Done():
				ctx.LogWarning(fmt.Sprintf("module=TrafficControlSort\ttimeout in goruntine: <taskId:%s/targetId:%s>[targetName:%s]",
					taskId, targetId, controller.target.Name))
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
				targetOutput[pair.string] = output
			}
		case <-gCtx.Done():
			if errors.Is(gCtx.Err(), gocontext.DeadlineExceeded) {
				ctx.LogWarning(fmt.Sprintf("module=TrafficControlSort\tflow control\ttraffic controller timeout: %v", gCtx.Err()))
			}
			break Loop
		}
	}
	if !hasTraffic {
		ctx.LogWarning(fmt.Sprintf("module=TrafficControlSort\tflow control\tno traffic data detected, maybe flink job is not running"))
		for k := range targetOutput {
			delete(targetOutput, k)
		}
		cnt = 0
	}
	return targetOutput, cnt
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
	ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tsample control target\tselected uplift target ids %v score", selected))
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
func microControl(controllerMap map[string]*PIDController, items []*module.Item, ctx *context.RecommendContext, wgCtrl *sync.WaitGroup) {
	defer wgCtrl.Done()
	itemTargetTraffic := loadTargetItemTraffic(ctx, items, controllerMap) // key1: targetId, key2: itemId, value: traffic
	if ctx.Debug {
		data, _ := json.Marshal(itemTargetTraffic)
		ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tmicro control\titem target traffic:%s", string(data)))
	}
	params := ctx.ExperimentResult.GetExperimentParams()
	maxUpliftCnt := params.GetInt("pid_max_uplift_item_cnt", 5)
	ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tmicro control\tpid_max_uplift_item_cnt=%d", maxUpliftCnt))
	upliftCnt := 0
	maxScore := 0.0

	//Calculate the number of items controlled by each target, key:targetId; value:control item sum
	controlNumberMap := make(map[string]int)

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
			if !controller.IsControlledItem(item) {
				continue
			}
			controlNumberMap[targetId] = controlNumberMap[targetId] + 1
			traffic := float64(0)
			if dict, ok := itemTargetTraffic[targetId]; ok {
				if value, okay := dict[string(item.Id)]; okay {
					traffic = value
				}
			}
			alpha, setValue := controller.Compute(string(item.Id), ctx)
			delta := alpha
			pos, _ := item.IntProperty("_ORIGIN_POSITION_")
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
			}
			deltaRank += delta // 多个目标调控方向不一致时，需要扳手腕看谁力气大
			ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\t[targetId:%s/targetName:%s], itemId:%s, "+
				"origin pos=%d, traffic=%.0f, setValue=%f, percentage=%f, alpha=%f, delta rank=%f",
				targetId, controller.target.Name, item.Id, pos, traffic, setValue, traffic/setValue, alpha, delta))
		}
		if deltaRank != 0.0 {
			ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\titem:%v\tdelta rank:%v", item.Id, deltaRank))
			if deltaRank < 0 {
				item.IncrAlgoScore("__delta_rank__", deltaRank)
			} else if upliftCnt < maxUpliftCnt { // uplift
				item.IncrAlgoScore("__delta_rank__", deltaRank)
				upliftCnt++
				pos, _ := item.IntProperty("_ORIGIN_POSITION_")
				if pos > ctx.Size && deltaRank >= 1.0 {
					item.AddProperty("__traffic_control_id__", 0)
				}
			}
		}
	}
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
		ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tcompute delta rank\titem %s [%s/%s], "+
			"score proportion=%.3f, rho=%.3f, alpha=%.6f, origin pos=%d, delta rank=%.6f [pull down]",
			item.Id, c.target.TrafficControlTargetId, c.target.Name, scoreWeight, rho, alpha, rank+1, deltaRank))
	} else { // uplift
		deltaRank *= itemScore // item.Score 越大，提权越多；用来在不同提取目标间竞争
		distinctStartPos := ctx.Size
		if scoreWeight > args.keepCtrlIdScore && args.pageNo > 1 {
			multiple := (scoreWeight - 0.3) * 10
			distinctStartPos += int(multiple * float64(ctx.Size))
		}
		if rank > distinctStartPos && deltaRank >= 1.0 {
			needNewCtrlId := args.needNewCtrlId[c.target.TrafficControlTargetId] || c.target.SplitParts.SetValues[0]/100 > int64(args.newCtrlIdThreshold)
			if c.task.ControlType == constants.TrafficControlTaskControlTypePercent && needNewCtrlId {
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
		ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tcompute delta rank\titem:%s\t[targetId:%s/targetName:%s],"+
			"score proportion=%.3f,norm_score=%.3f, alpha=%.6f, origin pos=%d, delta rank=%.6f, traffic_control_id=%d [uplift]",
			item.Id, c.target.TrafficControlTargetId, c.target.Name, scoreWeight, itemScore, alpha, rank+1, deltaRank, controlId))
	}
	return deltaRank
}

func isControlUser(user *module.User, controllerMap map[string]*PIDController) map[string]*PIDController {
	controllerNewMap := make(map[string]*PIDController)

	for targetId, controller := range controllerMap {
		userExpression := controller.userExpression

		if userExpression == "" {
			controllerNewMap[targetId] = controller
			continue
		}

		expression, err := govaluate.NewEvaluableExpression(userExpression)
		if err != nil {
			log.Error(fmt.Sprintf("module=PIDController\tgenerate user expression field, expression:%s, err:%v", userExpression, err))
			return controllerNewMap
		}
		properties := user.Properties

		result, err := expression.Evaluate(properties)
		if err != nil {
			log.Error(fmt.Sprintf("module=PIDController\tcompute user expression field, err:%v", err))
			return controllerNewMap
		}
		ok := ToBool(result, false)
		if ok {
			controllerNewMap[targetId] = controller
		}
	}
	return controllerNewMap
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

func setHyperParams(controllers map[string]*PIDController, ctx *context.RecommendContext) bool {
	params := ctx.ExperimentResult.GetExperimentParams()
	on := params.GetInt("pid_control_enable", 1)
	if on == 0 {
		return false
	}

	offPrefix := params.GetString("pid_off_target_name_prefix", "")
	if offPrefix != "" {
		for _, controller := range controllers {
			if strings.HasPrefix(controller.target.Name, offPrefix) {
				controller.SetOnline(false)
				ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tset target turn off, targetName:%s", controller.target.Name))
			}
		}
	}
	onPrefix := params.GetString("pid_on_target_name_prefix", "")
	if onPrefix != "" {
		for _, controller := range controllers {
			if strings.HasPrefix(controller.target.Name, onPrefix) {
				controller.SetOnline(true)
				ctx.LogDebug(fmt.Sprintf("module=TrafficControlSort\tset target turn on, targetName:%s", controller.target.Name))
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
		var defaultKp, defaultKi, defaultKd, defaultErrDiscount float64
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
						c.SetParameters(defaultKp, defaultKi, defaultKd)
					}
					c.SetStartPageNum(defaultStartPageNo)
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
						c.SetParameters(defaultKp, defaultKi, defaultKd)
					}
					continue
				}
				var kp, ki, kd float64
				if _kp, exist := dict["kp"]; exist {
					kp = _kp.(float64)
				}
				if _ki, exist := dict["ki"]; exist {
					ki = _ki.(float64)
				}
				if _kd, exist := dict["kd"]; exist {
					kd = _kd.(float64)
				}
				c.SetParameters(kp, ki, kd)
				if threshold, exist := dict["integral_threshold"]; exist {
					c.SetIntegralThreshold(threshold.(float64))
				}
				if threshold, exist := dict["err_threshold"]; exist {
					c.SetErrorThreshold(threshold.(float64))
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
