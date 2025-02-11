package sort

import (
	"encoding/json"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/alibaba/pairec/v2/constants"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
)

var targetMap map[string]model.TrafficControlTarget // key: targetId, value: target

type PIDController struct {
	task             *model.TrafficControlTask   // the meta info of current task
	target           *model.TrafficControlTarget // the meta info of current target
	kp               float64                     // The value for the proportional gain
	ki               float64                     // The value for the integral gain
	kd               float64                     // The value for the derivative gain
	errDiscount      float64                     // the discount of err sum
	status           *PIDStatus
	itemStatusMap    sync.Map // for single granularity or experiment wisely task, key is itemId or exp id, value is pidStatus
	timestamp        int64    // set timestamp to this value to get task meta info of that time
	allocateExpWise  bool     // whether to allocate traffic experiment wisely。如果为是，每个实验都达成目标，如果为否，整个链路达成目标
	minExpTraffic    float64  // the minimum traffic to activate the experimental control
	userExpression   string   // the conditions of user
	itemExpression   string   // the conditions of candidate items
	startPageNum     int      // turn off pid controller when current pageNum < startPageNum
	online           bool
	freezeMinutes    int // use last output when time less than this at everyday morning
	aheadMinutes     int // get dynamic target value ahead of this minutes
	runWithZeroInput bool
	integralMin      float64 // 积分项最小值
	integralMax      float64 // 积分项最大值
}

type PIDStatus struct {
	mu              sync.Mutex
	integral        float64   // 积分项累积值
	lastError       float64   // 上次计算的误差
	lastMeasurement float64   // 上次测量的实际流量值
	lastTime        time.Time // 上次测量时间
	derivative      float64   // 微分项计算值
	lastOutput      float64   // 上次输出的值
}

type Expression struct {
	Field  string      `json:"field"`
	Option string      `json:"option"`
	Value  interface{} `json:"value"`
}

func NewPIDController(task *model.TrafficControlTask, target *model.TrafficControlTarget, conf *recconf.PIDControllerConfig, expId string) *PIDController {
	loadTrafficControlTargetData(task.SceneName, conf.Timestamp)
	controller := PIDController{
		task:             task,
		target:           target,
		kp:               float64(conf.DefaultKp),
		ki:               float64(conf.DefaultKi),
		kd:               float64(conf.DefaultKd),
		errDiscount:      1.0,
		status:           &PIDStatus{},
		timestamp:        conf.Timestamp,
		allocateExpWise:  conf.AllocateExperimentWise,
		aheadMinutes:     conf.AheadMinutes,
		online:           true,
		runWithZeroInput: true,
		integralMin:      -100.0,
		integralMax:      100.0,
	}
	if conf.DefaultKi == 0 {
		controller.ki = 1.0
	}
	if conf.DefaultKd == 0 {
		controller.kd = 1.0
	}
	if conf.DefaultKp == 0 {
		controller.kp = 10.0
	}
	if conf.AheadMinutes < 1 {
		controller.aheadMinutes = 1
	}
	if conf.IntegralMin < 0 {
		controller.integralMin = conf.IntegralMin
	}
	if conf.IntegralMax > 0 {
		controller.integralMax = conf.IntegralMax
	}
	controller.GenerateItemExpress()
	log.Info(fmt.Sprintf("NewPIDController:\texp=%s\ttaskId:%s\ttaskName=%s\ttargetId:%s\ttargetName:%s",
		expId, controller.task.TrafficControlTaskId, controller.task.Name, controller.target.TrafficControlTargetId,
		controller.target.Name))
	return &controller
}

func loadTrafficControlTargetData(sceneName string, timestamp int64) {
	runEnv := os.Getenv("PAIREC_ENVIRONMENT")
	targetMap = experimentClient.GetTrafficControlTargetData(runEnv, sceneName, timestamp)
}

// SetMeasurement 处理积分饱和问题: 当系统长时间偏离目标值时，积分项会累积很大的值，导致控制信号过大，系统恢复时会有较大的超调。
// 处理每分钟一次的输入延迟和实时调用的矛盾: 根据测量时间间隔来计算积分项和微分项。
func (p *PIDController) SetMeasurement(itemOrExpId string, measurement float64, measureTime time.Time) {
	var status = p.getPIDStatus(itemOrExpId)
	if !measureTime.After(status.lastTime) {
		return
	}
	setValue, enabled := p.getTargetSetValue()
	if !enabled {
		return
	}
	if setValue < 1 {
		setValue = 1
	}

	var currentError float64
	if p.task.ControlType == constants.TrafficControlTaskControlTypePercent {
		currentError = setValue/100.0 - measurement
	} else {
		currentError = 1.0 - measurement/setValue
	}

	status.mu.Lock()
	defer status.mu.Unlock()

	if status.lastTime.IsZero() {
		status.lastTime = measureTime
		status.lastMeasurement = measurement
		status.lastError = currentError
		return
	}

	// 计算时间差（秒）
	dt := measureTime.Sub(status.lastTime).Seconds()

	// 计算并限制积分项
	if p.errDiscount != 1.0 {
		status.integral *= p.errDiscount
	}
	status.integral += currentError * dt
	if status.integral > p.integralMax {
		status.integral = p.integralMax
	} else if status.integral < p.integralMin {
		status.integral = p.integralMin
	}

	// 计算微分项
	status.derivative = (currentError - status.lastError) / dt

	// 更新状态记录
	status.lastError = currentError
	status.lastMeasurement = measurement
	status.lastTime = measureTime

	log.Info(fmt.Sprintf("module=PIDController\ttarget=[%s/%s]\titemIdOrExpId=%s\terr=%f,lastErr=%f,"+
		"derivative=%f,integral=%f,dt=%f,measure=%.6f", p.target.TrafficControlTargetId, p.target.Name, itemOrExpId,
		currentError, status.lastError, status.derivative, status.integral, dt, measurement))
}

func (p *PIDController) Compute(itemOrExpId string, ctx *context.RecommendContext) (float64, float64) {
	if !p.online {
		return 0, 0
	}
	var status = p.getPIDStatus(itemOrExpId)
	status.mu.Lock()
	defer status.mu.Unlock()

	measure := status.lastMeasurement
	if p.task.ControlType == constants.TrafficControlTaskControlTypePercent && measure > 1.0 {
		ctx.LogError(fmt.Sprintf("module=PIDController\tinvalid traffic percentage <taskId:%s/targetId%s>[targetName:%s] value=%f",
			p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, measure))
		return 0, 0
	}
	if measure == 0 && !p.runWithZeroInput {
		return 0, 0
	}

	setValue, enabled := p.getTargetSetValue()
	if !enabled {
		return 0, setValue
	}
	if setValue < 1 {
		setValue = 1
	}
	if p.task.ControlLogic == constants.TrafficControlTaskControlLogicGuaranteed {
		// 调控类型为保量，并且当前时刻目标已达成的情况下，直接返回 0
		if p.task.ControlType == constants.TrafficControlTaskControlTypePercent {
			if measure >= (setValue / 100) {
				return 0, setValue
			}
		} else {
			if measure >= setValue {
				return 0, setValue
			}
		}
	}
	if p.task.ControlType == constants.TrafficControlTaskControlTypePercent {
		if BetweenSlackInterval(measure, setValue/100, float64(p.target.ToleranceValue)/100) {
			// when current input is between `setValue` and `setValue+SetVale Range`, turn off controller
			return 0, setValue
		}
	} else {
		if BetweenSlackInterval(measure, setValue, float64(p.target.ToleranceValue)) {
			return 0, setValue
		}
	}

	if p.freezeMinutes > 0 {
		// 流量占比型任务凌晨刚开始的时候流量占比统计值不置信，直接输出前一天最后一次的调控信号
		// 定义东八区时区
		location, err := time.LoadLocation("Asia/Shanghai") // 北京采用Asia/Shanghai时区
		if err != nil {
			// 如果无法加载时区，默认使用本地时区
			location = time.Local
		}

		// 获取当前时间
		now := time.Now().In(location)

		// 获取当天的零点时间
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)

		// 计算当前时间与零点的差值
		durationSinceStartOfDay := int(now.Sub(startOfDay).Minutes())
		if durationSinceStartOfDay < p.freezeMinutes {
			ctx.LogDebug(fmt.Sprintf("module=PIDController\titemIdOrExpId=%s\texit within the freezing time", itemOrExpId))
			return status.lastOutput, setValue
		}
	}

	var currentError float64
	if p.task.ControlType == constants.TrafficControlTaskControlTypePercent {
		currentError = setValue/100.0 - measure
	} else {
		currentError = 1.0 - measure/setValue
	}

	pTerm := p.kp * currentError
	iTerm := p.ki * status.integral
	dTerm := p.kd * status.derivative
	status.lastOutput = pTerm + iTerm + dTerm
	return status.lastOutput, setValue
}

/*
func (p *PIDController) Do(trafficOrPercent float64, ctx *context.RecommendContext) (float64, float64) {
	return p.DoWithId(trafficOrPercent, "", ctx)
}

func (p *PIDController) DoWithId(trafficOrPercent float64, itemOrExpId string, ctx *context.RecommendContext) (float64, float64) {
	if !p.online {
		return 0, 0
	}
	if p.task.ControlType == constants.TrafficControlTaskControlTypePercent && trafficOrPercent > 1.0 {
		log.Error(fmt.Sprintf("module=PIDController\tinvalid traffic percentage <taskId:%s/targetId%s>[targetName:%s] value=%f",
			p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, trafficOrPercent))
		return 0, 0
	}
	if trafficOrPercent == 0 && !p.runWithZeroInput {
		return 0, 0
	}
	setValue, enabled := p.getTargetSetValue()
	if !enabled {
		return 0, setValue
	}
	if setValue < 1 {
		setValue = 1
	}

	if p.task.ControlLogic == constants.TrafficControlTaskControlLogicGuaranteed {
		// 调控类型为保量，并且当前时刻目标已达成的情况下，直接返回 0
		if p.task.ControlType == constants.TrafficControlTaskControlTypePercent {
			if trafficOrPercent >= (setValue / 100) {
				return 0, setValue
			}
		} else {
			if trafficOrPercent >= setValue {
				return 0, setValue
			}
		}
	}
	if p.task.ControlType == constants.TrafficControlTaskControlTypePercent {
		if BetweenSlackInterval(trafficOrPercent, setValue/100, float64(p.target.ToleranceValue)/100) {
			// when current input is between `setValue` and `setValue+SetVale Range`, turn off controller
			return 0, setValue
		}
	} else {
		if BetweenSlackInterval(trafficOrPercent, setValue, float64(p.target.ToleranceValue)) {
			return 0, setValue
		}
	}

	now := time.Now()
	curTime := now.Unix()
	var status = p.readPIDStatus(itemOrExpId, curTime)
	timeDiff := curTime - status.LastTime
	// 时间差还在一个时间窗口内
	if timeDiff < int64(p.timeWindow) && status.LastOutput != 0 {
		ctx.LogDebug(fmt.Sprintf("module=PIDController\titemIdOrExpId=%s\tuse last output", itemOrExpId))
		return float64(status.LastOutput), setValue
	}
	if p.freezeMinutes > 0 {
		// 流量占比型任务凌晨刚开始的时候流量占比统计值不置信，直接输出前一天最后一次的调控信号
		curHour := now.Hour()
		curMinute := now.Minute()
		elapseMinutes := curHour*60 + curMinute
		if elapseMinutes < p.freezeMinutes {
			ctx.LogDebug(fmt.Sprintf("module=PIDController\titemIdOrExpId=%s\texit within the freezing time", itemOrExpId))
			return float64(status.LastOutput), setValue
		}
	}
	if timeDiff < int64(p.timeWindow) {
		timeDiff = int64(p.timeWindow)
	}
	dt := float32(timeDiff) / float32(p.timeWindow)

	// 时间过了几个时间窗口，要加上误差衰减系数
	if p.errDiscount != 1.0 {
		status.ErrSum *= float32(math.Pow(p.errDiscount, float64(dt)))
	}

	var err float32
	if p.task.ControlType == constants.TrafficControlTaskControlTypePercent {
		err = float32(setValue/100.0 - trafficOrPercent)
	} else {
		err = float32(1.0 - trafficOrPercent/setValue)
	}
	status.ErrSum += err * dt
	dErr := (err - status.LastError) / dt

	// Compute final output
	output := p.kp*err + p.ki*status.ErrSum + p.kd*dErr
	ctx.LogInfo(fmt.Sprintf("module=PIDController\ttarget=[%s/%s]\titemIdOrExpId=%s\terr=%f,lastErr=%f,dErr=%f,"+
		"ErrSum=%f,dt=%f,input=%.6f,output=%v", p.target.TrafficControlTargetId, p.target.Name, itemOrExpId, err,
		status.LastError, dErr, status.ErrSum, dt, trafficOrPercent, output))

	// Keep track of state
	status.LastOutput = output
	status.LastError = err
	status.LastTime = curTime
	if p.syncStatus {
		go p.writePIDStatus(itemOrExpId) // 通过外部存储同步中间状态
	}
	return float64(output), setValue
}
*/
// 获取被拆解的目标值
func (p *PIDController) getTargetSetValue() (float64, bool) {
	endTime, _ := time.Parse("2006-01-02T15:04:05+08:00", p.target.EndTime)
	startTime, _ := time.Parse("2006-01-02T15:04:05+08:00", p.target.StartTime)
	if endTime.Unix() < startTime.Unix() {
		log.Warning(fmt.Sprintf("module=PIDController\tinvalid target end time and start time, targetId:%s\tstartTime:%v\tendTime:%v",
			p.target.TrafficControlTargetId, startTime, endTime))
		return 0, false
	}
	now := time.Now().UTC().Add(time.Hour * 8)
	if now.Unix() < startTime.Unix() {
		log.Warning(fmt.Sprintf("module=PIDController\tcurrent time is before target start time, targetId:%s\tcurrentTime:%v\tstartTime:%v",
			p.target.TrafficControlTargetId, now, startTime))
		return 0, false
	}
	if now.Unix() > endTime.Unix() {
		log.Warning(fmt.Sprintf("module=PIDController\tcurrent time is after target end time, targetId:%s\tcurrentTime:%v\tendTime:%v",
			p.target.TrafficControlTargetId, now, endTime))
		return 0, false
	}
	loadTrafficControlTargetData(p.task.SceneName, p.timestamp)
	if target, ok := targetMap[p.target.TrafficControlTargetId]; ok {
		p.target = &target
	} else {
		return 0, false
	}
	n := len(p.target.SplitParts.SetValues)
	if n == 0 {
		log.Error(fmt.Sprintf("module=PIDController\tthe size of target set values array is 0, targetId:%s",
			p.target.TrafficControlTargetId))
		return p.target.Value, false
	}
	if len(p.target.SplitParts.TimePoints) != n {
		log.Error(fmt.Sprintf("module=PIDController\tthe size of target time points array is not equal to the size of target set values array, targetId:%s",
			p.target.TrafficControlTargetId))
		return p.target.Value, false
	}
	if p.task.ControlType == constants.TrafficControlTaskControlTypePercent {
		return float64(p.target.SplitParts.SetValues[n-1]), true // p.target.SetValues 不会被动态更新
	}

	// 获取当前时间点的setValue, 调控目标拆解到分钟级
	curHour := now.Hour()
	curMinute := now.Minute()
	elapseMinutes := curHour*60 + curMinute + p.aheadMinutes
	morning := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	if p.target.StatisPeriod == constants.TrafficControlTargetStatisPeriodDaily {
		tomorrow := morning.AddDate(0, 0, 1)
		start, end := morning, tomorrow
		if startTime.After(morning) {
			start = startTime
		}
		if endTime.Before(tomorrow) {
			end = endTime
		}
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
		if startTime.After(morning) {
			beginHour := startTime.Hour()
			beginMinute := startTime.Minute()
			elapseMinutes -= beginHour*60 + beginMinute
		}
		return p.target.Value * float64(elapseMinutes) / 60.0, true
	} else {
		return float64(p.target.SplitParts.SetValues[n-1]), true // p.target.SetPoint 不会被动态更新
	}
}

func (p *PIDController) getPIDStatus(itemOrExpId string) *PIDStatus {
	var pidStatus *PIDStatus
	if itemOrExpId == "" {
		pidStatus = p.status
	} else if status, ok := p.itemStatusMap.Load(itemOrExpId); ok {
		pidStatus = status.(*PIDStatus)
	} else {
		pidStatus = &PIDStatus{}
		p.itemStatusMap.Store(itemOrExpId, pidStatus)
	}
	return pidStatus
}

/*
func (p *PIDController) readPIDStatus(itemOrExpId string, currentTimestamp int64) *PIDStatus {
	var pidStatus *PIDStatus
	if itemOrExpId == "" {
		pidStatus = p.status
	} else if status, ok := p.itemStatusMap.Load(itemOrExpId); ok {
		pidStatus = status.(*PIDStatus)
	} else {
		pidStatus = &PIDStatus{
			LastTime:   time.Now().Unix(),
			LastOutput: 0.0,
			LastError:  0.0,
			ErrSum:     0.0,
		}
		p.itemStatusMap.Store(itemOrExpId, pidStatus)
		return pidStatus
	}

	if !p.syncStatus {
		return pidStatus
	}

	timeInterval := int(currentTimestamp - pidStatus.LastTime)

	if timeInterval < p.timeWindow {
		return pidStatus
	}

	if currentTimestamp-serviceStartTimeStamp < 600 {
		return pidStatus // 服务启动的10分钟内不读状态, 重启任务需清空状态
	}

	cacheKey := p.cachePrefix + p.target.TrafficControlTargetId + "_" + itemOrExpId
	value := p.cache.Get(cacheKey)
	if value != nil {
		status := value.([]byte)
		pid := PIDStatus{}
		if err := json.Unmarshal(status, &pid); err != nil {
			log.Error(fmt.Sprintf("module=PIDController\tread PID status <taskId:%s/targetId%s>[targetName:%s] key=%s failed. err=%v",
				p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, cacheKey, err))
		} else {
			if pid.LastTime > pidStatus.LastTime {
				pidStatus.LastTime = pid.LastTime
				pidStatus.ErrSum = pid.ErrSum
				pidStatus.LastError = pid.LastError
				pidStatus.LastOutput = pid.LastOutput
			}
		}
	}
	return pidStatus
}

func (p *PIDController) writePIDStatus(itemOrExpId string) {
	cacheKey := p.cachePrefix + p.target.TrafficControlTargetId + "_" + itemOrExpId
	var pidStatus *PIDStatus
	if itemOrExpId == "" {
		pidStatus = p.status
	} else if status, ok := p.itemStatusMap.Load(itemOrExpId); ok {
		pidStatus = status.(*PIDStatus)
	}
	if pidStatus == nil {
		log.Error(fmt.Sprintf("module=PIDController\tno PID status to be written, <taksId:%s/targetId:%s>[targetName:%s] key=%s",
			p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, cacheKey))
		return
	}
	data, err := json.Marshal(*pidStatus)
	if err != nil {
		log.Error(fmt.Sprintf("module=PIDControl\tPID status convert to string failed. err=%v", err))
		return
	}
	err = p.cache.Put(cacheKey, data, p.cacheTime)
	if err != nil {
		log.Error(fmt.Sprintf("module=PIDController\twrite PID status <taskId:%s/targetId:%s>[targetName:%s] key=%s failed. err=%v",
			p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, cacheKey, err))
	} else {
		log.Info(fmt.Sprintf("module=PIDController\twrite PID status <taskId:%s/targetId:%s>[targetName:%s] key=%s, value=%s",
			p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, cacheKey, string(data)))
	}
}
*/

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
	if targetExpression != "" && taskExpression != "" {
		p.itemExpression = fmt.Sprintf("%s&&%s", taskExpression, targetExpression)
	} else if targetExpression != "" {
		p.itemExpression = targetExpression
	} else if taskExpression != "" {
		p.itemExpression = taskExpression
	}
}

func BetweenSlackInterval(trafficValue, setValue, delta float64) bool {
	var lower, upper float64
	if delta > 0 {
		lower = setValue
		upper = setValue + delta
	} else {
		lower = setValue + delta
		upper = setValue
	}
	if lower <= trafficValue && trafficValue <= upper {
		return true
	}
	return false
}

func (p *PIDController) IsControlledItem(item *module.Item) bool {
	if p.itemExpression == "" {
		return true
	}
	expression, err := govaluate.NewEvaluableExpression(p.itemExpression)
	if err != nil {
		log.Error(fmt.Sprintf("module=PIDController\tgenerate item expression field, itemId:%s,expression:%s, err:%v",
			item.Id, p.itemExpression, err))
		return false
	}
	properties := item.GetCloneFeatures()

	result, err := expression.Evaluate(properties)
	if err != nil {
		//log.Warning(fmt.Sprintf("module=PIDController\tcompute item expression field, itemId:%s, err:%v", item.Id, err))
		return false
	}

	return ToBool(result, false)
}

func (p *PIDController) IsControlledTraffic(ctx *context.RecommendContext) bool {
	if !p.online {
		return false
	}
	s := ctx.GetParameter("scene")
	if scene, good := s.(string); !good {
		return false
	} else if scene != p.task.SceneName {
		return false
	}
	pageNo := utils.ToInt(ctx.GetParameter("pageNum"), 1)
	if pageNo < p.startPageNum {
		return false
	}
	return true
}

func (p *PIDController) GetMinExpTraffic() float64 {
	return p.minExpTraffic
}

func (p *PIDController) SetMinExpTraffic(traffic float64) {
	p.minExpTraffic = traffic
}

func (p *PIDController) SetFreezeMinutes(minutes int) {
	if 0 < minutes && minutes < 1440 {
		p.freezeMinutes = minutes
	}
}

func (p *PIDController) SetRunWithZeroInput(run bool) {
	if run != p.runWithZeroInput {
		log.Info(fmt.Sprintf("module=PIDController\t<taskId:%s/targetId:%s>[targetName:%s] set runWithZeroInput=%v",
			p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, run))
		p.runWithZeroInput = run
	}
}

func (p *PIDController) SetOnline(online bool) {
	if p.online != online {
		p.online = online
		log.Info(fmt.Sprintf("module=PIDController\t<taskId:%s/targetId:%s>[targetName:%s] set online=%v",
			p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, p.online))
	}
}

func (p *PIDController) SetAllocateExpWise(wise bool) {
	p.allocateExpWise = wise
}

func (p *PIDController) IsAllocateExpWise() bool {
	return p.allocateExpWise
}

func (p *PIDController) SetParameters(kp, ki, kd float64) {
	changed := false
	if p.kp != kp {
		p.kp = kp
		changed = true
	}
	if p.ki != ki {
		p.ki = ki
		changed = true
	}
	if p.kd != kd {
		p.kd = kd
		changed = true
	}
	if changed {
		p.status.lastOutput = 0.0
		p.itemStatusMap.Range(func(key, value interface{}) bool {
			value.(*PIDStatus).lastOutput = 0.0
			return true
		})
		log.Info(fmt.Sprintf("module=PIDController\tThe parameters of PIDController <taskId:%s/targetId:%s>[targetName:%s] changed to: kp=%f, ki=%f, kd=%f",
			p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, p.kp, p.ki, p.kd))
	}
}

//func (p *PIDController) SetTimeWindow(timeWindow int) {
//	if timeWindow > 0 {
//		p.timeWindow = timeWindow
//	}
//}

func (p *PIDController) SetErrDiscount(decay float64) {
	if decay > 0 {
		p.errDiscount = decay
	}
}

func (p *PIDController) SetUserExpress(expression string) {
	p.userExpression = expression
	if expression != "" {
		log.Info(fmt.Sprintf("module=PIDController\t<taskId:%s/targetId:%s>[targetName:%s] set userConditions=%s",
			p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, expression))
	}
}

func (p *PIDController) SetStartPageNum(pageNum int) {
	p.startPageNum = pageNum
}

//func ToString(task interface{}, excludes ...string) string {
//	typeOfTask := reflect.TypeOf(task)
//	valueOfTask := reflect.ValueOf(task)
//	fields := make(map[string]interface{})
//	for i := 0; i < typeOfTask.NumField(); i++ {
//		fieldType := typeOfTask.Field(i)
//		fieldName := fieldType.Name
//		if utils.StringContains(excludes, []string{fieldName}) {
//			continue
//		}
//		fieldValue := valueOfTask.Field(i)
//		fields[fieldName] = fieldValue.Interface()
//	}
//	if jsonStr, err := json.Marshal(fields); err == nil {
//		return string(jsonStr)
//	}
//	return fmt.Sprintf("%v", fields)
//}

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
			}
			conditionExpr := fmt.Sprintf("%s%s%v", condition.Field, condition.Option, condition.Value)
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
