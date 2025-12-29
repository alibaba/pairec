package sort

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/constants"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
	"github.com/goburrow/cache"
)

type PIDController struct {
	task              *model.TrafficControlTask   // the meta info of current task
	target            *model.TrafficControlTarget // the meta info of current target
	kp                float64                     // The value for the proportional gain
	ki                float64                     // The value for the integralSum gain
	kd                float64                     // The value for the derivative gain
	globalStatus      *PIDStatus                  //for global control status
	expIdStatusMap    sync.Map                    // for experiment wisely task, key is exp id, value is pidStatus
	singleStatusMap   sync.Map                    // for single granularity or experiment wisely task, key is itemId or exp id, value is pidStatus
	allocateExpWise   bool                        // whether to allocate traffic experiment wisely。如果为是，每个实验都达成目标，如果为否，整个链路达成目标
	minExpTraffic     float64                     // the minimum traffic to activate the experimental control
	freezeMinutes     int                         // use last output when time less than this at everyday morning
	integralMin       float64                     // 积分项最小值, 通常 integralMin = -integralMax
	integralMax       float64                     // 积分项最大值, 根据最大控制量需求计算： integralMax = (MaxOutput - Kp*MaxError) / Ki
	integralThreshold float64                     // 激活积分项的误差阈值，通常设为目标值的10%-20%
	errDiscount       float64                     // err discount
	errThreshold      float64                     // 变速积分阈值, 初始值设为目标值的20%-30%; 快速响应系统：较大阈值; 慢速系统：较小阈值
	memberCache       cache.Cache

	userExprProg *vm.Program // the compiled expression of valid user
	itemExprProg *vm.Program // the compiled expression of candidate items
}

func NewPIDController(task *model.TrafficControlTask, target *model.TrafficControlTarget, conf *recconf.PIDControllerConfig) *PIDController {
	controller := PIDController{
		task:              task,
		target:            target,
		kp:                conf.DefaultKp,
		ki:                conf.DefaultKi,
		kd:                conf.DefaultKd,
		errDiscount:       1.0,
		globalStatus:      &PIDStatus{integralActive: true},
		allocateExpWise:   conf.AllocateExperimentWise,
		integralMin:       -100.0,
		integralMax:       100.0,
		integralThreshold: conf.IntegralThreshold,
		errThreshold:      conf.ErrThreshold,
		freezeMinutes:     conf.FreezeMinutes,
		minExpTraffic:     conf.MinExpTraffic,
	}
	if conf.MembershipCacheSeconds > 0 {
		controller.memberCache = cache.New(cache.WithMaximumSize(100000),
			cache.WithExpireAfterWrite(time.Second*time.Duration(conf.MembershipCacheSeconds)))
	} else {
		controller.memberCache = cache.New(cache.WithMaximumSize(100000),
			cache.WithExpireAfterWrite(time.Second*time.Duration(60)))
	}
	if conf.DefaultKi == 0 {
		controller.ki = 10.0
	}
	if conf.DefaultKd == 0 {
		controller.kd = 10.0
	}
	if conf.DefaultKp == 0 {
		controller.kp = 1000.0
	}
	if conf.ErrDiscount > 0 {
		controller.errDiscount = conf.ErrDiscount
	}
	if conf.IntegralMin < 0 {
		controller.integralMin = conf.IntegralMin
	}
	if conf.IntegralMax > 0 {
		controller.integralMax = conf.IntegralMax
	}
	controller.GenerateItemExpress()
	controller.GenerateUserExpress()
	log.Info(fmt.Sprintf("NewPIDController:\ttaskId:%s\ttaskName=%s\ttargetId:%s\ttargetName:%s", controller.task.TrafficControlTaskId, controller.task.Name, controller.target.TrafficControlTargetId, controller.target.Name))
	return &controller
}

// Compute 测量值更新与实际控制分离设计，控制计算始终使用最新可用测量值和实时分解的目标值
func (p *PIDController) compute(ctx *context.RecommendContext, expId, itemId string, controllerParams controllerParams) (alpha float64, aimValue float64) {
	status := p.getPIDStatus(expId, itemId)
	//status.mu.Lock()
	//defer status.mu.Unlock()

	isPercentageTask := p.task.ControlType == constants.TrafficControlTaskControlTypePercent
	lastMeasurement := status.GetLastMeasurement()
	if isPercentageTask && lastMeasurement > 1.0 {
		ctx.LogError(fmt.Sprintf("module=PIDController\tinvalid traffic percentage <taskId:%s/targetId:%s>[targetName:%s] value=%f", p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, lastMeasurement))
		return 0, 0
	}

	if isPercentageTask || time.Now().Sub(status.lastTime) < time.Duration(30)*time.Second {
		aimValue = status.controlAimValue
	} else {
		value, enabled := p.getControlTargetAimValue()
		if !enabled {
			ctx.LogDebug(fmt.Sprintf("module=PIDController\t<taskId:%s/targetId:%s>[targetName:%s] disable", p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name))
			return 0, value
		}
		if value < 1 {
			aimValue = 1
		} else {
			aimValue = value
		}
	}

	if p.task.ControlLogic == constants.TrafficControlTaskControlLogicGuaranteed || p.task.ControlLogic == "" {
		// 调控类型为保量，并且当前时刻目标已达成的情况下，alpha值直接返回 0
		if isPercentageTask {
			if lastMeasurement >= (aimValue / 100) {
				ctx.LogDebug(fmt.Sprintf("module=PIDController\t<taskId:%s/targetId:%s>[targetName:%s] expId=%s, lastMeasurement=%.6f, achieved", p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, expId, lastMeasurement))
				return 0, aimValue
			}
		} else {
			if lastMeasurement >= aimValue {
				ctx.LogDebug(fmt.Sprintf("module=PIDController\t<taskId:%s/targetId:%s>[targetName:%s] item=%s, lastMeasurement=%.2f, achieved", p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, itemId, lastMeasurement))
				return 0, aimValue
			}
		}
	}
	// 处理容错区域
	if isPercentageTask {
		if p.target.ToleranceValue != 0 {
			if isInToleranceRange(lastMeasurement, aimValue/100, float64(p.target.ToleranceValue)) {
				// when current input is between `controlAimValue` and `controlAimValue+SetVale Range`, turn off controller
				ctx.LogDebug(fmt.Sprintf("module=PIDController\t<taskId:%s/targetId:%s>[targetName:%s] expId=%s, between slack interval, achieved", p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, expId))
				return 0, aimValue
			}
		}
	} else {
		if p.target.ToleranceValue != 0 {
			if isInToleranceRange(lastMeasurement, aimValue, float64(p.target.ToleranceValue)) {
				ctx.LogDebug(fmt.Sprintf("module=PIDController\t<taskId:%s/targetId:%s>[targetName:%s] itemId=%s, between slack interval, achieved", p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, itemId))
				return 0, aimValue
			}
		}
	}

	if p.freezeMinutes > 0 {
		// 流量占比型任务凌晨刚开始的时候流量占比统计值不置信，直接输出前一天最后一次的调控信号
		location, err := time.LoadLocation("Asia/Shanghai") // 北京采用Asia/Shanghai时区
		if err != nil {                                     // 如果无法加载时区，默认使用本地时区
			location = time.Local
		}
		// 获取当前时间
		now := time.Now().In(location)
		// 获取当天的零点时间
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
		// 计算当前时间与零点的差值
		durationSinceStartOfDay := int(now.Sub(startOfDay).Minutes())
		if durationSinceStartOfDay < p.freezeMinutes {
			ctx.LogDebug(fmt.Sprintf("module=PIDController\titemId=%s,expId=%s\texit within the freezing time", itemId, expId))
			return status.lastOutput, aimValue
		}
	}

	var currentError float64
	if isPercentageTask {
		currentError = aimValue/100.0 - lastMeasurement
	} else {
		currentError = 1.0 - lastMeasurement/aimValue
	}
	var kp, ki, kd float64
	if controllerParams.Kp != nil {
		kp = *controllerParams.Kp
	} else {
		kp = p.kp
	}
	if controllerParams.Ki != nil {
		ki = *controllerParams.Ki
	} else {
		ki = p.ki
	}
	if controllerParams.Kd != nil {
		kd = *controllerParams.Kd
	} else {
		kd = p.kd
	}

	pTerm := kp * currentError
	dTerm := kd * status.derivative
	var iTerm float64
	if status.integralActive { // 积分项（根据激活状态决定是否使用）
		iTerm = ki * status.integralSum
	}
	status.lastOutput = pTerm + iTerm + dTerm
	if status.lastOutput == 0 {
		ctx.LogDebug(fmt.Sprintf("module=PIDController\t<taskId:%s/targetId:%s>[targetName:%s] itemId=%s, expId=%s lastMeasurement=%f, err=%.6f, output=0", p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, itemId, expId, lastMeasurement, currentError))
	}
	return status.lastOutput, aimValue
}

// SetMeasurement 处理积分饱和问题: 当系统长时间偏离目标值时，积分项会累积很大的值，导致控制信号过大，系统恢复时会有较大的超调。
// 1. 在每次测量更新时才进行积分项限制，防止长期误差累积；处理每分钟一次的输入延迟和实时调用的矛盾: 根据测量时间间隔来计算积分项和微分项。
// 2. 通过设置integralMin和integralMax对积分项进行限幅
// 3. 积分分离法机制: 在误差较大时关闭积分项，避免积分累积过快导致超调
// 4. 变速积分机制: 根据误差的大小动态调整积分项的系数，从而平滑地控制积分的作用
func (p *PIDController) setMeasurement(expId, itemId string, measurement float64, measureTime time.Time, controllerParams controllerParams) {
	var status = p.getPIDStatus(expId, itemId)
	// measureTime 是流量上报时间（测量时间）
	if !measureTime.After(status.lastTime) {
		return
	}
	aimValue, enabled := p.getControlTargetAimValue()
	if !enabled {
		return
	}
	if aimValue < 1 {
		aimValue = 1
	}

	isPercentageTask := p.task.ControlType == constants.TrafficControlTaskControlTypePercent
	var achieved bool
	if p.task.ControlLogic == constants.TrafficControlTaskControlLogicGuaranteed || p.task.ControlLogic == "" {
		// 调控类型为保量，并且当前时刻目标已达成的情况下，直接返回
		if isPercentageTask {
			if measurement >= (aimValue / 100) {
				achieved = true
			}
		} else if measurement >= aimValue {
			achieved = true
		}
	}
	var currentError float64
	if !achieved {
		if isPercentageTask {
			currentError = aimValue/100.0 - measurement
		} else {
			currentError = 1.0 - measurement/aimValue
		}
	}

	if achieved || status.lastTime.IsZero() {
		status.lastTime = measureTime
		status.lastMeasurement = measurement
		status.lastError = currentError
		status.controlAimValue = aimValue
		return
	}

	// 计算时间差（秒）
	dt := measureTime.Sub(status.lastTime).Seconds()

	var errDiscount float64
	if controllerParams.PidErrDiscount != nil {
		errDiscount = *controllerParams.PidErrDiscount
	} else {
		errDiscount = p.errDiscount
	}

	if p.errDiscount != 1.0 {
		status.integralSum *= errDiscount
	}
	// 计算变速积分系数
	factor := p.variableIntegralFactor(currentError)
	// 积分分离逻辑：仅在误差小于阈值时累积积分
	var integralThreshold float64
	if controllerParams.PidIntegralThreshold != nil {
		integralThreshold = *controllerParams.PidIntegralThreshold
	} else {
		integralThreshold = p.integralThreshold
	}

	if integralThreshold > 0 {
		if math.Abs(currentError) <= integralThreshold {
			status.integralSum += factor * currentError * dt
			status.integralActive = true
		} else { // 进入积分分离区，暂停积分累积
			status.integralActive = false
		}
	} else {
		status.integralSum += factor * currentError * dt
	}
	// 计算并限制积分项
	if status.integralSum > p.integralMax {
		status.integralSum = p.integralMax
	} else if status.integralSum < p.integralMin {
		status.integralSum = p.integralMin
	}

	// 避免除零错误
	if dt < 1e-9 {
		dt = 1e-9
	}
	// 计算微分项
	status.derivative = (currentError - status.lastError) / dt

	logProb := 1.0
	if p.task.ControlGranularity == constants.TrafficControlTaskControlGranularitySingle {
		logProb = 0.1
	}
	if rand.Float64() < logProb {
		log.Info(fmt.Sprintf("module=PIDController\ttarget=[%s/%s]\titemId=%s\terr=%f,lastErr=%f, derivative=%f,integralSum=%f,dt=%.2f,measure=%.6f,time=%v", p.target.TrafficControlTargetId, p.target.Name, itemId,
			currentError, status.lastError, status.derivative, status.integralSum, dt, measurement, measureTime))
	}

	// 更新状态记录
	status.lastError = currentError
	status.lastMeasurement = measurement
	status.lastTime = measureTime
	status.controlAimValue = aimValue
}

// 变速积分函数（可根据需要修改插值算法）
// 自动适应系统不同工作状态: 快速响应阶段：侧重比例微分作用; 精细调节阶段：增强积分作用
func (p *PIDController) variableIntegralFactor(currentError float64) float64 {
	if p.errThreshold == 0 {
		return 1.0
	}
	absErr := math.Abs(currentError)
	// return math.Exp(-absErr/pid.errThreshold) // 指数衰减型积分系数
	if absErr > p.errThreshold {
		return 0.0 // 完全禁用积分
	}
	// 线性插值：误差越小，积分系数越大
	return 1.0 - absErr/p.errThreshold
}

// 获取调控任务被拆解的目标值
func (p *PIDController) getControlTargetAimValue() (float64, bool) {
	now := time.Now()
	n := len(p.target.SplitParts.SetValues)
	if n == 0 {
		log.Error(fmt.Sprintf("module=PIDController\tthe size of target set values array is 0, targetId:%s",
			p.target.TrafficControlTargetId))
		return float64(p.target.Value), false
	}
	if len(p.target.SplitParts.TimePoints) != n {
		log.Error(fmt.Sprintf("module=PIDController\tthe size of target time points array is not equal to the size of target set values array, targetId:%s",
			p.target.TrafficControlTargetId))
		return float64(p.target.Value), false
	}
	if p.task.ControlType == constants.TrafficControlTaskControlTypePercent {
		return float64(p.target.SplitParts.SetValues[n-1]), true // p.target.SetValues 不会被动态更新
	}

	// 获取当前时间点的目标值, 调控目标拆解到分钟级
	curHour := now.Hour()
	curMinute := now.Minute()
	elapseMinutes := curHour*60 + curMinute
	morning := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	if p.target.StatisPeriod == constants.TrafficControlTargetStatisPeriodDaily {
		tomorrow := morning.AddDate(0, 0, 1)
		start, end := morning, tomorrow
		duration := end.Sub(start)
		if duration < time.Hour*24 {
			// part day
			d := time.Since(start)
			progress := d.Seconds() / duration.Seconds()
			return float64(p.target.SplitParts.SetValues[n-1]) * progress, true
		}
		// whole day
		startTimePoint := 0
		timeSpan := p.target.SplitParts.TimePoints[0]
		for i, timePoint := range p.target.SplitParts.TimePoints {
			if timePoint >= elapseMinutes {
				base := 0.0
				if i > 0 {
					startTimePoint = p.target.SplitParts.TimePoints[i-1]
					timeSpan = timePoint - startTimePoint
					base = float64(p.target.SplitParts.SetValues[i-1])
					if base < 1.0 {
						base = 1.0
					}
				}
				timeProgress := float64(elapseMinutes-startTimePoint) / float64(timeSpan)
				//  每一个时间段内，目标随时间线性增长
				next := p.target.SplitParts.SetValues[i]
				if next < 1.0 {
					next = 1.0
				}
				return base + (float64(next)-base)*timeProgress, true
			}
		}
		return float64(p.target.SplitParts.SetValues[n-1]), true
	} else if p.target.StatisPeriod == "Hour" {
		return float64(p.target.Value) * float64(elapseMinutes) / 60.0, true
	} else {
		return float64(p.target.SplitParts.SetValues[n-1]), true // p.target.SetPoint 不会被动态更新
	}
}

// 实际值是否在目标值的误差范围内
func isInToleranceRange(actualValue, aimValue, tolerance float64) bool {
	if tolerance > 0 {
		upperBound := aimValue + tolerance/100*aimValue
		lowerBound := aimValue
		if actualValue <= upperBound && actualValue >= lowerBound {
			return true
		}
	} else if tolerance < 0 {
		upperBound := aimValue
		lowerBound := aimValue - tolerance/100*aimValue
		if actualValue <= upperBound && actualValue >= lowerBound {
			return true
		}
	}
	return false
}

func (p *PIDController) GetMinExpTraffic() float64 {
	return p.minExpTraffic
}

func (p *PIDController) IsAllocateExpWise() bool {
	return p.allocateExpWise
}

func (p *PIDController) GenerateItemExpress() {
	targetExpression, err := ParseExpression(p.target.ItemConditionArray, p.target.ItemConditionExpress)
	if err != nil {
		log.Error(fmt.Sprintf("module=PIDController\tparse item condition field, please check %s or %s\terr:%v",
			p.target.ItemConditionArray, p.target.ItemConditionExpress, err))
		return
	}

	taskExpression, err := ParseExpression(p.task.ItemConditionArray, p.target.ItemConditionExpress)
	if err != nil {
		log.Error(fmt.Sprintf("module=PIDController\tparse item condition field, please check %s or %s\terr:%v",
			p.task.UserConditionArray, p.task.UserConditionExpress, err))
		return
	}

	var itemExpression string
	if targetExpression != "" && taskExpression != "" {
		itemExpression = fmt.Sprintf("%s&&%s", taskExpression, targetExpression)
	} else if targetExpression != "" {
		itemExpression = targetExpression
	} else if taskExpression != "" {
		itemExpression = taskExpression
	}
	if itemExpression != "" {
		p.itemExprProg, err = expr.Compile(itemExpression, expr.AsBool())
		if err != nil {
			log.Error(fmt.Sprintf("module=PIDController\tcompile item expression field, expression:%s, err:%v", itemExpression, err))
			return
		}
		log.Info(fmt.Sprintf("module=PIDController\tcompile item expression success, expression:%s", itemExpression))
	}
}

func (p *PIDController) GenerateUserExpress() {
	taskUserExpression, err := ParseExpression(p.task.UserConditionArray, p.task.UserConditionExpress)
	if err != nil {
		log.Error(fmt.Sprintf("module=PIDControl\tparse user condition field, please check %s or %s",
			p.task.UserConditionArray, p.task.UserConditionExpress))
	}
	p.userExprProg, err = expr.Compile(taskUserExpression, expr.AsBool())
	if err != nil {
		log.Error(fmt.Sprintf("module=PIDController\tcompile user expression field, expression:%s, err:%v", taskUserExpression, err))
		return
	}
	log.Info(fmt.Sprintf("module=PIDController\tcompile user expression success, expression:%s", taskUserExpression))
}

type PIDStatus struct {
	controlAimValue float64   // 调控目标值
	integralSum     float64   // 积分项累积值
	lastError       float64   // 上次计算的误差
	lastMeasurement float64   // 上次测量的实际流量值
	lastTime        time.Time // 上次测量时间
	derivative      float64   // 微分项计算值
	lastOutput      float64   // 上次输出的值
	integralActive  bool      // 当前是否激活积分项
}

func (s *PIDStatus) GetLastMeasurement() float64 {
	if s.lastTime.IsZero() {
		return 0
	}
	location, err := time.LoadLocation("Asia/Shanghai") // 北京采用Asia/Shanghai时区
	if err != nil {                                     // 如果无法加载时区，默认使用本地时区
		location = time.Local
	}
	now := time.Now().In(location) // 获取当前时间
	last := s.lastTime.In(location)
	if now.Day() != last.Day() {
		return 0
	}
	if now.Sub(last) >= time.Hour*24 {
		return 0
	}
	return s.lastMeasurement
}

func (p *PIDController) getPIDStatus(expId, itemId string) *PIDStatus {
	var pidStatus *PIDStatus

	// 单品调控时，各个item的状态
	if itemId != "" {
		status, ok := p.singleStatusMap.Load(itemId)
		if ok {
			pidStatus = status.(*PIDStatus)
		} else {
			pidStatus = &PIDStatus{integralActive: true}
			p.singleStatusMap.Store(itemId, pidStatus)
		}
		return pidStatus
	}
	// 比例调控时，当想要每个实验达成目标，各个实验的状态
	if expId != "" {
		status, ok := p.expIdStatusMap.Load(expId)
		if ok {
			pidStatus = status.(*PIDStatus)
		} else {
			pidStatus = &PIDStatus{integralActive: true}
			p.expIdStatusMap.Store(expId, pidStatus)
		}
	}

	if expId == "" {
		return p.globalStatus
	}
	return pidStatus
}

func IsExprMatch(conditions []*Expression, item *module.Item) bool {
	for _, expression := range conditions {
		field := expression.Field
		op := expression.Option
		value := expression.Value
		v := item.GetProperty(field)
		if v == nil {
			return false
		}
		switch op {
		case "=":
			if utils.NotEqual(v, value) {
				return false
			}
		case "!=":
			if utils.Equal(v, value) {
				return false
			}
		case ">":
			if utils.LessEqual(v, value) {
				return false
			}
		case ">=":
			if utils.Less(v, value) {
				return false
			}
		case "<":
			if utils.GreaterEqual(v, value) {
				return false
			}
		case "<=":
			if utils.Greater(v, value) {
				return false
			}
		case "in":
			if !utils.In(v, value) {
				return false
			}
		}
	}
	return true
}

type Expression struct {
	Field  string      `json:"field"`
	Type   string      `json:"type"`
	Option string      `json:"option"`
	Value  interface{} `json:"value"`
}

func ParseExpression(conditionArray, conditionExpress string) (string, error) {
	var express string
	if conditionArray != "" {
		var conditions []*Expression
		err := json.Unmarshal([]byte(conditionArray), &conditions)
		if err != nil {
			return "", err
		}

		for _, condition := range conditions {
			if condition.Option == "=" {
				condition.Option = "=="

				switch condition.Value.(type) {
				case string:
					condition.Value = fmt.Sprintf("'%s'", condition.Value)
				}
			} else if condition.Option == "in" {
				if condition.Type == "STRING" {
					valueArr := strings.Split(condition.Value.(string), ",")
					for i, value := range valueArr {
						valueArr[i] = fmt.Sprintf("'%s'", value)
					}
					condition.Value = fmt.Sprintf("[%v]", strings.Join(valueArr, ","))
				} else {
					condition.Value = fmt.Sprintf("[%v]", condition.Value)
				}
			}
			conditionExpr := fmt.Sprintf("%s %s %v", condition.Field, condition.Option, condition.Value)
			if express == "" {
				express = conditionExpr
			} else {
				express = fmt.Sprintf("%s&&%s", express, conditionExpr)
			}
		}
	}
	if conditionExpress != "" {
		if express == "" {
			express = conditionExpress
		} else {
			express = fmt.Sprintf("%s&&%s", express, conditionExpress)
		}
	}
	return express, nil
}
