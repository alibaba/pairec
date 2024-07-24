package sort

import (
	"encoding/json"
	"fmt"
	"github.com/alibaba/pairec/v2/abtest"
	"github.com/alibaba/pairec/v2/constants"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/experiments"
	"math"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/persist/cache"
	"github.com/alibaba/pairec/v2/persist/redisdb"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
)

var itemStatusMap sync.Map          // for single granularity or experiment wisely task // key 是 itemId or exp id, value: pidStatus
var pidStatusCacheRedis cache.Cache // If the engine has multiple instances, it is used to synchronize values between multiple instances
var once sync.Once
var targetMap map[string]model.TrafficControlTarget // key: targetId, value: target
var serviceStartTimeStamp int64                     // engine start time stamp
var ExperimentClient *experiments.ExperimentClient

type PIDController struct {
	task             *model.TrafficControlTask   // the meta info of current task
	target           *model.TrafficControlTarget // the meta info of current target
	kp               float32                     // The value for the proportional gain
	ki               float32                     // The value for the integral gain
	kd               float32                     // The value for the derivative gain
	timeWindow       int                         // The time in seconds which the controller should wait before generating a new output value
	errDiscount      float64                     // the discount of err sum. 误差衰减系数
	status           *PIDStatus
	cache            cache.Cache // 下列的 redis 存储
	cachePrefix      string
	cacheTime        time.Duration
	testTimestamp    int64         // set timestamp to this value to get task meta info of that time
	allocateExpWise  bool          // whether to allocate traffic experiment wisely。如果为是，每个实验都达成目标，如果为否，整个链路达成目标
	minExpTraffic    float64       // the minimum traffic to activate the experimental control
	conditions       []*Expression // the conditions of traffic to be allocated
	matchConditions  []*Expression // the match conditions of context and item pair
	itemConditions   []*Expression // the conditions of candidate items
	startPageNum     int           // turn off pid controller when current pageNum < startPageNum
	online           bool
	syncStatus       bool // whether to sync pid status between instances
	freezeMinutes    int  // use last output when time less than this at everyday morning
	runWithZeroInput bool
}

type PIDStatus struct {
	LastTime   int64
	LastOutput float32 // the @ value in the formula
	LastError  float32
	ErrSum     float32
}

type Expression struct {
	Field  string      `json:"field"`
	Option string      `json:"option"`
	Value  interface{} `json:"value"`
}

func NewPIDController(task *model.TrafficControlTask, target *model.TrafficControlTarget, conf *recconf.PIDControllerConfig, expId string) *PIDController {
	once.Do(func() {
		serviceStartTimeStamp = time.Now().Unix()
		if conf.SyncPIDStatus {
			redisConf, err := redisdb.GetRedisConf(conf.RedisName)
			if err != nil {
				log.Error(fmt.Sprintf("module=PIDController\terror=%v", err))
				return
			}
			b, err := json.Marshal(redisConf)
			if err != nil {
				log.Error(fmt.Sprintf("module=PIDController\tMarshal redis conf failed, err=%v", err))
				return
			}
			pidStatusCacheRedis, err = cache.NewCache("redis", string(b))
			if err != nil {
				log.Error(fmt.Sprintf("module=PIDController\tnew redis cache failed. error=%v", err))
				return
			}
		}
		ExperimentClient = abtest.GetExperimentClient()
		if ExperimentClient == nil {
			log.Error("module=PIDControl\tGetExperimentClient failed.")
			return
		}
		loadTrafficControlTargetData(task.SceneName, conf.TestTimestamp) // 第一次要执行完才能继续执行创建任务
	})

	if conf.SyncPIDStatus && pidStatusCacheRedis == nil {
		log.Error("module=PIDController\tcreate pid controller failed because of init redis cache failed")
		return nil
	}

	timeWindow := 30 // 时间窗口内控制信号保持不变
	if conf.TimeWindow > 0 {
		timeWindow = conf.TimeWindow
	}
	cachePrefix := "_PID_"
	if conf.RedisKeyPrefix != "" {
		cachePrefix = conf.RedisKeyPrefix
	}
	runEnv := os.Getenv("PAIREC_ENVIRONMENT")
	if runEnv != "" {
		cachePrefix += runEnv + "_"
	}
	if expId != "" {
		cachePrefix += expId + "_"
	}
	status := PIDStatus{
		LastTime:   time.Now().Unix(),
		LastOutput: 0.0,
		LastError:  0.0,
		ErrSum:     0.0,
	}
	controller := PIDController{
		task:             task,
		target:           target,
		kp:               conf.DefaultKp,
		ki:               conf.DefaultKi,
		kd:               conf.DefaultKd,
		timeWindow:       timeWindow,
		errDiscount:      1.0,
		status:           &status,
		cache:            pidStatusCacheRedis,
		cachePrefix:      cachePrefix,
		cacheTime:        time.Hour,
		testTimestamp:    conf.TestTimestamp,
		allocateExpWise:  conf.AllocateExperimentWise,
		online:           true,
		runWithZeroInput: true,
		syncStatus:       conf.SyncPIDStatus,
	}
	controller.GenerateItemConditions()
	log.Info(fmt.Sprintf("NewPIDController:\texp=%s\t%s, target:%s", expId, ToString(controller.task, "targets"), ToString(controller.target, "TargetTraffics", "PlanTraffic")))
	return &controller
}

func loadTrafficControlTargetData(sceneName string, timePoint int64) {
	runEnv := os.Getenv("PAIREC_ENVIRONMENT")
	targetMap = ExperimentClient.GetTrafficControlTargetData(runEnv, sceneName, timePoint)
}

func (p *PIDController) SetOnline(online bool) {
	if p.online != online {
		p.online = online
		log.Info(fmt.Sprintf("module=PIDController\tPIDController <%s/%s>[%s] set online=%v", p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, p.online))
	}
}

func (p *PIDController) SetAllocateExpWise(wise bool) {
	p.allocateExpWise = wise
}

func (p *PIDController) IsAllocateExpWise() bool {
	return p.allocateExpWise
}

func (p *PIDController) SetParameters(kp, ki, kd float32) {
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
		p.status.LastOutput = 0.0
		itemStatusMap.Range(func(key, value interface{}) bool {
			value.(*PIDStatus).LastOutput = 0.0
			return true
		})
		log.Info(fmt.Sprintf("module=PIDController\tThe parameters of PIDController <%s/%s>[%s] changed to: kp=%f, ki=%f, kd=%f", p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, p.kp, p.ki, p.kd))
	}
}

func (p *PIDController) SetTimeWindow(timeWindow int) {
	if timeWindow > 0 {
		p.timeWindow = timeWindow
	}
}

func (p *PIDController) SetErrDiscount(decay float64) {
	if decay > 0 {
		p.errDiscount = decay
	}
}

func (p *PIDController) SetConditions(conditions []*Expression) {
	p.conditions = conditions
	if len(conditions) > 0 {
		log.Info(fmt.Sprintf("module=PIDController\tPIDController <%s/%s>[%s] set conditions=%v", p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, p.conditions))
	}
}

func (p *PIDController) SetStartPageNum(pageNum int) {
	p.startPageNum = pageNum
}

func BetweenSlackInterval(input, setPoint, delta int64) bool {
	var lower, upper int64
	if delta > 0 {
		lower = setPoint
		upper = setPoint + delta
	} else {
		lower = setPoint + delta
		upper = setPoint
	}
	if lower <= input && input <= upper {
		return true
	}
	return false
}

func (p *PIDController) Do(input float64) (float64, float64) {
	return p.DoWithId(input, "")
}

func (p *PIDController) DoWithId(targetValue float64, itemOrExpId string) (float64, float64) {
	if !p.online {
		return 0, 0
	}
	if p.task.ControlType == constants.TrafficControlTaskControlTypePercent && targetValue > 1.0 {
		log.Error(fmt.Sprintf("module=PIDController\tinvalid traffic percentage <%s/%s>[%s] value=%f", p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, targetValue))
		return 0, 0
	}
	if targetValue == 0 && !p.runWithZeroInput {
		return 0, 0
	}
	setValue, enabled := p.getSetValue()
	if !enabled {
		return 0, setValue
	}
	if p.task.ControlLogic == constants.TrafficControlTaskControlLogicGuaranteed && targetValue >= setValue {
		// 调控类型为"保量"，并且当前时刻目标已达成的情况下，直接返回0
		return 0, setValue
	}
	if BetweenSlackInterval(int64(targetValue), int64(setValue), p.target.ToleranceValue) {
		// when current input is between `setValue` and `setValue+SetPointRange`, turn off controller
		return 0, setValue
	}

	now := time.Now()
	curTime := now.Unix()
	var status = p.readStatus(itemOrExpId, curTime)
	timeDiff := curTime - status.LastTime
	// 时间差还在一个时间窗口内
	if timeDiff < int64(p.timeWindow) && status.LastOutput != 0 {
		return float64(status.LastOutput), setValue
	}
	if p.freezeMinutes > 0 {
		// 流量占比型任务凌晨刚开始的时候流量占比统计值不置信，直接输出前一天最后一次的调控信号
		curHour := now.Hour()
		curMinute := now.Minute()
		elapseMinutes := curHour*60 + curMinute
		if elapseMinutes < p.freezeMinutes {
			return float64(status.LastOutput), setValue
		}
	}

	if timeDiff < 1 {
		timeDiff = 1
	}

	// 时间过了几个时间窗口，要加上误差衰减系数
	if p.errDiscount != 1.0 {
		status.ErrSum *= float32(math.Pow(p.errDiscount, float64(timeDiff/int64(p.timeWindow))))
	}

	var err float32
	if p.task.ControlType == constants.TrafficControlTaskControlTypePercent {
		err = float32(setValue - targetValue)
	} else {
		err = float32(1.0 - targetValue/setValue)
	}
	status.ErrSum += err * float32(timeDiff)
	dErr := (err - status.LastError) / float32(timeDiff)

	// Compute final output
	output := p.kp*err + p.ki*status.ErrSum + p.kd*dErr

	// Keep track of state
	status.LastOutput = output
	status.LastError = err
	status.LastTime = curTime
	if p.syncStatus {
		go p.writeStatus(itemOrExpId) // 通过外部存储同步中间状态
	}
	return float64(output), setValue
}

// 获取被拆解的目标值
func (p *PIDController) getSetValue() (float64, bool) {
	endTime, _ := time.Parse("2006-01-02T15:04:05.00+08:00", p.target.EndTime)
	startTime, _ := time.Parse("2006-01-02T15:04:05.00+08:00", p.target.StartTime)
	if !endTime.After(startTime) {
		return 0, false
	}
	now := time.Now()
	if !now.After(startTime) {
		return 0, false
	}
	loadTrafficControlTargetData(p.task.SceneName, p.testTimestamp)
	if target, ok := targetMap[p.target.TrafficControlTargetId]; ok {
		p.target = &target
	} else {
		return 0, false
	}
	n := len(p.target.SplitParts.SetValues)
	if n == 0 {
		log.Error(fmt.Sprintf("module=PIDController\tthe size of target set values array is 0: %v", p.target))
		return p.target.Value, false
	}
	if len(p.target.SplitParts.TimePoints) != n {
		log.Error(fmt.Sprintf("module=PIDController\tthe size of target time points array is not equal to the size of target set values array: %v", p.target))
		return p.target.Value, false
	}
	if p.task.ControlType == "Percent" {
		return float64(p.target.SplitParts.SetValues[n-1]), true // p.target.SetValues 不会被动态更新
	}

	// 获取当前时间点的setPoint, 调控目标拆解到分钟级
	curHour := now.Hour()
	curMinute := now.Minute()
	elapseMinutes := curHour*60 + curMinute + 1
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
				}
				timeProgress := float64(elapseMinutes-startTimePoint) / float64(timeSpan)
				//  每一个时间段内，目标随时间线性增长
				next := p.target.SplitParts.SetValues[i]
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

func (p *PIDController) readStatus(itemOrExpId string, timestamp int64) *PIDStatus {
	var pidStatus *PIDStatus
	if itemOrExpId == "" {
		pidStatus = p.status
	} else if status, ok := itemStatusMap.Load(itemOrExpId); ok {
		pidStatus = status.(*PIDStatus)
	} else {
		pidStatus = &PIDStatus{
			LastTime:   time.Now().Unix(),
			LastOutput: 0.0,
			LastError:  0.0,
			ErrSum:     0.0,
		}
		itemStatusMap.Store(itemOrExpId, pidStatus)
		return pidStatus
	}

	if !p.syncStatus {
		return pidStatus
	}

	timeInterval := int(timestamp - pidStatus.LastTime)

	if timeInterval < p.timeWindow {
		return pidStatus
	}

	if timestamp-serviceStartTimeStamp < 600 {
		return pidStatus // 服务启动的10分钟内不读状态, 重启任务需清空状态
	}

	cacheKey := p.cachePrefix + p.target.TrafficControlTargetId + "_" + itemOrExpId
	value := p.cache.Get(cacheKey)
	if value != nil {
		status := value.([]byte)
		pid := PIDStatus{}
		if err := json.Unmarshal(status, &pid); err != nil {
			log.Error(fmt.Sprintf("read PID status <%s/%s>[%s] key=%s failed. err=%v", p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, cacheKey, err))
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

func (p *PIDController) writeStatus(itemOrExpId string) {
	cacheKey := p.cachePrefix + p.target.TrafficControlTargetId + "_" + itemOrExpId
	var pidStatus *PIDStatus
	if itemOrExpId == "" {
		pidStatus = p.status
	} else if status, ok := itemStatusMap.Load(itemOrExpId); ok {
		pidStatus = status.(*PIDStatus)
	}
	if pidStatus == nil {
		log.Error(fmt.Sprintf("no PID status to be written, <%s/%s>[%s] key=%s", p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, cacheKey))
		return
	}
	data, err := json.Marshal(*pidStatus)
	if err != nil {
		log.Error(fmt.Sprintf("module=PIDControl\tMsg=PID status convert to string failed. err=%v", err))
		return
	}
	err = p.cache.Put(cacheKey, data, p.cacheTime)
	if err != nil {
		log.Error(fmt.Sprintf("write PID status <%s/%s>[%s] key=%s failed. err=%v", p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, cacheKey, err))
	} else {
		log.Info(fmt.Sprintf("write PID status <%s/%s>[%s] key=%s, value=%s", p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, cacheKey, string(data)))
	}
}

func (p *PIDController) GenerateItemConditions() {
	var targetExpressions []*Expression
	err := json.Unmarshal([]byte(p.target.ItemConditionArray), &targetExpressions)
	if err != nil {
		log.Error(fmt.Sprintf("module=PIDController\tUnmarshal ItemConditionArray '%s' failed. err=%v", p.target.ItemConditionArray, err))
		return
	}
	var taskExpressions []*Expression
	err = json.Unmarshal([]byte(p.task.ItemConditionArray), &taskExpressions)
	if err != nil {
		log.Error(fmt.Sprintf("module=PIDController\tUnmarshal ItemConditionArray '%s' failed. err=%v", p.task.UserConditionArray, err))
		return
	}
	targetExpressions = append(targetExpressions, taskExpressions...)
	p.itemConditions = targetExpressions
}

func (p *PIDController) IsControlledItem(ctx *context.RecommendContext, item *module.Item) bool {
	for _, expr := range p.itemConditions {
		field := expr.Field
		op := expr.Option
		value := expr.Value
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
	return p.IsContextMatch(ctx, item)
}

func (p *PIDController) IsContextMatch(ctx *context.RecommendContext, item *module.Item) bool {
	if p.matchConditions == nil || len(p.matchConditions) == 0 {
		return true
	}
	for _, expr := range p.matchConditions {
		value := ctx.GetParameterByPath(expr.Field)
		if value == nil {
			return true
		}
		itemField := expr.Value.(string)
		v := item.GetProperty(itemField)
		if v == nil {
			return true
		}
		switch expr.Option {
		case "=":
			if utils.NotEqual(value, v) {
				return false
			}
		case "!=":
			if utils.Equal(value, v) {
				return false
			}
		case ">":
			if utils.LessEqual(value, v) {
				return false
			}
		case ">=":
			if utils.Less(value, v) {
				return false
			}
		case "<":
			if utils.GreaterEqual(value, v) {
				return false
			}
		case "<=":
			if utils.Greater(value, v) {
				return false
			}
		case "in":
			if !utils.In(value, v) {
				return false
			}
		}
	}
	return true
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
	if p.conditions == nil || len(p.conditions) == 0 {
		return true
	}

	for _, expr := range p.conditions {
		value := ctx.GetParameterByPath(expr.Field)
		if value == nil {
			if expr.Option == "!=" {
				return true
			}
			return false
		}
		targetValue := utils.ToString(expr.Value, "")
		if targetValue[0] == uint8('$') && utils.IsDateExpression(targetValue) {
			if date, ok := utils.EvalDate(targetValue); ok {
				targetValue = date
			} else {
				ctx.LogError("parse date failed: " + targetValue)
			}
		}
		switch expr.Option {
		case "=":
			if utils.NotEqual(value, targetValue) {
				return false
			}
		case "!=":
			if utils.Equal(value, targetValue) {
				return false
			}
		case ">":
			if utils.LessEqual(value, targetValue) {
				return false
			}
		case ">=":
			if utils.Less(value, targetValue) {
				return false
			}
		case "<":
			if utils.GreaterEqual(value, targetValue) {
				return false
			}
		case "<=":
			if utils.Greater(value, targetValue) {
				return false
			}
		case "in":
			if !utils.In(value, expr.Value) {
				return false
			}
		}
	}
	return true
}

func (p *PIDController) SetMinExpTraffic(traffic float64) {
	p.minExpTraffic = traffic
}

func (p *PIDController) GetMinExpTraffic() float64 {
	return p.minExpTraffic
}

func (p *PIDController) SetFreezeMinutes(minutes int) {
	if 0 < minutes && minutes < 1440 {
		p.freezeMinutes = minutes
	}
}

func (p *PIDController) SetRunWithZeroInput(run bool) {
	if run != p.runWithZeroInput {
		log.Info(fmt.Sprintf("PIDController <%s/%s>[%s] set runWithZeroInput=%v",
			p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, run))
		p.runWithZeroInput = run
	}
}

func (p *PIDController) SetMatchConditions(conditions []*Expression) {
	p.matchConditions = conditions
	if len(conditions) > 0 {
		log.Info(fmt.Sprintf("PIDController <%s/%s>[%s] set match conditions=%v", p.task.TrafficControlTaskId, p.target.TrafficControlTargetId, p.target.Name, p.matchConditions))
	}
}

func ToString(task interface{}, excludes ...string) string {
	typeOfTask := reflect.TypeOf(task)
	valueOfTask := reflect.ValueOf(task)
	fields := make(map[string]interface{})
	for i := 0; i < typeOfTask.NumField(); i++ {
		fieldType := typeOfTask.Field(i)
		fieldName := fieldType.Name
		if utils.StringContains(excludes, []string{fieldName}) {
			continue
		}
		fieldValue := valueOfTask.Field(i)
		fields[fieldName] = fieldValue.Interface()
	}
	if jsonStr, err := json.Marshal(fields); err == nil {
		return string(jsonStr)
	}
	return fmt.Sprintf("%v", fields)
}
