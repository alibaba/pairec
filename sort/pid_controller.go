package sort

import (
	"encoding/json"
	"fmt"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/model"
	"github.com/alibaba/pairec/abtest"
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/persist/cache"
	"github.com/alibaba/pairec/persist/redisdb"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/utils"
	"math"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"
)

type PIDController struct {
	plan             model.FlowCtrlPlan        // the meta info of current plan
	target           model.FlowCtrlPlanTargets // the meta info of current target
	kp               float32                   // The value for the proportional gain
	ki               float32                   // The value for the integral gain
	kd               float32                   // The value for the derivative gain
	sampleTime       float32                   // The time in seconds which the controller should wait before generating a new output value
	errDiscount      float64                   // the discount of err sum
	status           *PIDStatus
	itemStatus       sync.Map // for single granularity or experiment wisely task
	cache            cache.Cache
	cachePrefix      string
	cacheTime        time.Duration
	testTimestamp    int64        // set timestamp to this value to get plan meta info of that time
	allocateExpWise  bool         // whether to allocate traffic experiment wisely
	minExpTraffic    float64      // the minimum traffic to activate the experimental control
	conditions       []Expression // the conditions of traffic to be allocated
	matchConditions  []Expression // the match conditions of context and item pair
	scopeConditions  []Expression // the conditions of candidate items
	startPageNum     int          // turn off pid controller when current pageNum < startPageNum
	online           bool
	syncStatus       bool // whether to sync pid status between instances
	freezeMinutes    int  // use last output when time less than this at everyday morning
	runWithZeroInput bool
}

type PIDStatus struct {
	LastTime   int64
	LastOutput float32
	LastError  float32
	ErrSum     float32
}

type Expression struct {
	Field  string      `json:"field"`
	Option string      `json:"option"`
	Value  interface{} `json:"value"`
}

var pidStatusCache cache.Cache
var pidInitOnce sync.Once
var pidTargets map[int]model.FlowCtrlPlanTargets
var serviceStartTimeStamp int64

func NewPIDController(plan model.FlowCtrlPlan, target model.FlowCtrlPlanTargets,
	conf *recconf.PIDControllerConfig, expId string) *PIDController {
	pidInitOnce.Do(func() {
		serviceStartTimeStamp = time.Now().Unix()
		if conf.SyncPIDStatus {
			redisConf, err := redisdb.GetRedisConf(conf.RedisName)
			if err != nil {
				log.Error(fmt.Sprintf("get redis `%s` failed. err=%v", conf.RedisName, err))
				return
			}
			b, err := json.Marshal(redisConf)
			if err != nil {
				log.Error(fmt.Sprintf("Marshal redis conf failed, err=%v", err))
				return
			}
			pidStatusCache, err = cache.NewCache("redis", string(b))
			if err != nil {
				log.Error(fmt.Sprintf("new redis cache failed. error=%v", err))
				return
			}
		}
		loadFlowCtrlTargetSetPoint(plan.SceneName, conf.TestTimestamp) // 第一次要执行完才能继续执行创建任务
	})

	if conf.SyncPIDStatus && pidStatusCache == nil {
		log.Error("create pid controller failed because of init redis cache failed")
		return nil
	}

	sampleTime := float32(30) // 30秒内控制信号保持不变
	if conf.SampleTime > 0.0 {
		sampleTime = conf.SampleTime
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
	pid := PIDController{
		plan:             plan,
		target:           target,
		kp:               conf.DefaultKp,
		ki:               conf.DefaultKi,
		kd:               conf.DefaultKd,
		sampleTime:       sampleTime,
		errDiscount:      1.0,
		status:           &status,
		cache:            pidStatusCache,
		cachePrefix:      cachePrefix,
		cacheTime:        time.Hour,
		testTimestamp:    conf.TestTimestamp,
		allocateExpWise:  conf.AllocateExperimentWise,
		online:           true,
		runWithZeroInput: true,
		syncStatus:       conf.SyncPIDStatus,
	}
	pid.GenScopeConditions()
	log.Info(fmt.Sprintf("NewPIDController:\texp=%s\t%s, target:%s", expId,
		ToString(pid.plan, "Targets"), ToString(pid.target, "TargetTraffics", "PlanTraffic")))
	return &pid
}

func loadFlowCtrlTargetSetPoint(sceneName string, timePoint int64) {
	client := abtest.GetExperimentClient()
	if client == nil {
		log.Error("module=loadFlowCtrlTargetSetPoint\tGetExperimentClient failed.")
		return
	}

	runEnv := os.Getenv("PAIREC_ENVIRONMENT")
	pidTargets = client.GetFlowCtrlPlanTargetList(runEnv, sceneName, timePoint)
}

func (p *PIDController) SetOnline(online bool) {
	if p.online != online {
		p.online = online
		log.Info(fmt.Sprintf("PIDController <%d/%d>[%s] set online=%v",
			p.plan.PlanId, p.target.TargetId, p.target.TargetName, p.online))
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
		p.itemStatus.Range(func(key, value interface{}) bool {
			value.(*PIDStatus).LastOutput = 0.0
			return true
		})
		log.Info(fmt.Sprintf("The parameters of PIDController <%d/%d>[%s] changed to: kp=%f, ki=%f, kd=%f",
			p.plan.PlanId, p.target.TargetId, p.target.TargetName, p.kp, p.ki, p.kd))
	}
}

func (p *PIDController) SetSampleTime(sampleTime float32) {
	if sampleTime > 0 {
		p.sampleTime = sampleTime
	}
}
func (p *PIDController) SetErrDiscount(discount float64) {
	if discount > 0 {
		p.errDiscount = discount
	}
}
func (p *PIDController) SetConditions(conditions []Expression) {
	p.conditions = conditions
	if len(conditions) > 0 {
		log.Info(fmt.Sprintf("PIDController <%d/%d>[%s] set conditions=%v",
			p.plan.PlanId, p.target.TargetId, p.target.TargetName, p.conditions))
	}
}
func (p *PIDController) SetStartPageNum(pageNum int) {
	p.startPageNum = pageNum
}

func BetweenSlackInterval(input, setPoint, delta float64) bool {
	var lower, upper float64
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

func (p *PIDController) DoWithId(input float64, itemOrExpId string) (float64, float64) {
	if !p.online {
		return 0, 0
	}
	if p.plan.TargetValueInPercentageFormat && input > 1.0 {
		log.Error(fmt.Sprintf("invalid traffic percentage <%d/%d>[%s] value=%f",
			p.plan.PlanId, p.target.TargetId, p.target.TargetName, input))
		return 0, 0
	}
	if input == 0 && !p.runWithZeroInput {
		return 0, 0
	}
	setPoint, enabled := p.getSetPoint()
	if !enabled {
		return 0, setPoint
	}
	if p.plan.PlanType == "guaranteed" && input >= setPoint {
		// 调控类型为"保量"，并且当前时刻目标已达成的情况下，直接返回0
		return 0, setPoint
	}
	if BetweenSlackInterval(input, setPoint, p.target.SetPointRange) {
		// when current input is between `setPoint` and `setPoint+SetPointRange`, turn off controller
		return 0, setPoint
	}

	now := time.Now()
	curTime := now.Unix()
	var status = p.readStatus(itemOrExpId, curTime)
	dt := float32(curTime - status.LastTime)
	if dt < p.sampleTime && status.LastOutput != 0 {
		return float64(status.LastOutput), setPoint
	}
	if p.freezeMinutes > 0 {
		// 流量占比型任务凌晨刚开始的时候流量占比统计值不置信，直接输出前一天最后一次的调控信号
		curHour := now.Hour()
		curMinute := now.Minute()
		elapseMinutes := curHour*60 + curMinute
		if elapseMinutes < p.freezeMinutes {
			return float64(status.LastOutput), setPoint
		}
	}

	if dt < 1 {
		dt = 1
	}
	if p.errDiscount != 1.0 {
		status.ErrSum *= float32(math.Pow(p.errDiscount, float64(dt/p.sampleTime)))
	}

	var err float32
	if p.plan.TargetValueInPercentageFormat {
		err = float32(setPoint - input)
	} else {
		err = float32(1.0 - input/setPoint)
	}
	status.ErrSum += err * dt
	dErr := (err - status.LastError) / dt

	// Compute final output
	output := p.kp*err + p.ki*status.ErrSum + p.kd*dErr

	// Keep track of state
	status.LastOutput = output
	status.LastError = err
	status.LastTime = curTime
	if p.syncStatus {
		go p.writeStatus(itemOrExpId) // 通过外部存储同步中间状态
	}
	return float64(output), setPoint
}

func (p *PIDController) getSetPoint() (float64, bool) {
	if !p.target.EndTime.After(p.target.StartTime) {
		return 0, false
	}
	now := time.Now()
	if !now.After(p.target.StartTime) {
		return 0, false
	}
	loadFlowCtrlTargetSetPoint(p.plan.SceneName, p.testTimestamp)
	if target, ok := pidTargets[p.target.TargetId]; ok {
		p.target = target
	} else {
		return 0, false
	}
	n := len(p.target.SetPoints)
	if n == 0 {
		log.Error(fmt.Sprintf("the size of target SetPoints array is 0: %v", p.target))
		return p.target.SetPoint, false
	}
	if len(p.target.TimePoints) != n {
		log.Error(fmt.Sprintf("the size of target TimePoints array is not equal to the size of target SetPoints array: %v", p.target))
		return p.target.SetPoint, false
	}
	if p.plan.TargetValueInPercentageFormat {
		return p.target.SetPoints[n-1], true // p.target.SetPoint 不会被动态更新
	}

	// 获取当前时间点的setPoint, 调控目标拆解到分钟级
	curHour := now.Hour()
	curMinute := now.Minute()
	elapseMinutes := curHour*60 + curMinute + 1
	morning := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	if p.target.TimeUint == "daily" {
		tomorrow := morning.AddDate(0, 0, 1)
		start, end := morning, tomorrow
		if p.target.StartTime.After(morning) {
			start = p.target.StartTime
		}
		if p.target.EndTime.Before(tomorrow) {
			end = p.target.EndTime
		}
		duration := end.Sub(start)
		if duration < time.Hour*24 {
			// part day
			d := time.Since(start)
			progress := d.Seconds() / duration.Seconds()
			return p.target.SetPoints[n-1] * progress, true
		}
		// whole day
		startTimePoint := 0
		timeSpan := p.target.TimePoints[0]
		for i, t := range p.target.TimePoints {
			if t >= elapseMinutes {
				base := 0.0
				if i > 0 {
					startTimePoint = p.target.TimePoints[i-1]
					timeSpan = t - startTimePoint
					base = p.target.SetPoints[i-1]
				}
				timeProgress := float64(elapseMinutes-startTimePoint) / float64(timeSpan)
				//  每一个时间段内，目标随时间线性增长
				next := p.target.SetPoints[i]
				return base + (next-base)*timeProgress, true
			}
		}
		return p.target.SetPoints[n-1], true
	} else if p.target.TimeUint == "hourly" {
		if p.target.StartTime.After(morning) {
			beginHour := p.target.StartTime.Hour()
			beginMinute := p.target.StartTime.Minute()
			elapseMinutes -= beginHour*60 + beginMinute
		}
		return p.target.SetPoint * float64(elapseMinutes) / 60.0, true
	} else {
		return p.target.SetPoints[n-1], true // p.target.SetPoint 不会被动态更新
	}
}

func (p *PIDController) readStatus(itemOrExpId string, now int64) *PIDStatus {
	var pidStatus *PIDStatus
	if itemOrExpId == "" {
		pidStatus = p.status
	} else if status, ok := p.itemStatus.Load(itemOrExpId); ok {
		pidStatus = status.(*PIDStatus)
	} else {
		pidStatus = &PIDStatus{
			LastTime:   time.Now().Unix(),
			LastOutput: 0.0,
			LastError:  0.0,
			ErrSum:     0.0,
		}
		p.itemStatus.Store(itemOrExpId, pidStatus)
		return pidStatus
	}

	if !p.syncStatus {
		return pidStatus
	}

	dt := float32(now - pidStatus.LastTime)
	if dt < p.sampleTime {
		return pidStatus
	}

	if now-serviceStartTimeStamp < 600 {
		return pidStatus // 服务启动的10分钟内不读状态, 重启任务需清空状态
	}

	cacheKey := p.cachePrefix + strconv.Itoa(p.target.TargetId) + "_" + itemOrExpId
	value := p.cache.Get(cacheKey)
	if value != nil {
		status := value.([]byte)
		pid := PIDStatus{}
		if err := json.Unmarshal(status, &pid); err != nil {
			log.Error(fmt.Sprintf("read PID status <%d/%d>[%s] key=%s failed. err=%v",
				p.plan.PlanId, p.target.TargetId, p.target.TargetName, cacheKey, err))
		} else {
			if pid.LastTime > pidStatus.LastTime {
				pidStatus.LastTime = pid.LastTime
				pidStatus.ErrSum = pid.ErrSum
				pidStatus.LastError = pid.LastError
				pidStatus.LastOutput = pid.LastOutput
				//if itemOrExpId == "" || (len(itemOrExpId) > 2 && itemOrExpId[:2] == "ER") {
				//	log.Info(fmt.Sprintf("read PID status <%d/%d>[%s] key=%s, value=%s",
				//		p.plan.PlanId, p.target.TargetId, p.target.TargetName, cacheKey, string(status)))
				//}
			}
		}
	}
	return pidStatus
}

func (p *PIDController) writeStatus(itemOrExpId string) {
	cacheKey := p.cachePrefix + strconv.Itoa(p.target.TargetId) + "_" + itemOrExpId
	var pidStatus *PIDStatus
	if itemOrExpId == "" {
		pidStatus = p.status
	} else if status, ok := p.itemStatus.Load(itemOrExpId); ok {
		pidStatus = status.(*PIDStatus)
	}
	if pidStatus == nil {
		log.Error(fmt.Sprintf("no PID status to be written, <%d/%d>[%s] key=%s",
			p.plan.PlanId, p.target.TargetId, p.target.TargetName, cacheKey))
		return
	}
	data, err := json.Marshal(*pidStatus)
	if err != nil {
		log.Error(fmt.Sprintf("PID status convert to string failed. err=%v", err))
		return
	}
	err = p.cache.Put(cacheKey, data, p.cacheTime)
	if err != nil {
		log.Error(fmt.Sprintf("write PID status <%d/%d>[%s] key=%s failed. err=%v",
			p.plan.PlanId, p.target.TargetId, p.target.TargetName, cacheKey, err))
	} else {
		log.Info(fmt.Sprintf("write PID status <%d/%d>[%s] key=%s, value=%s",
			p.plan.PlanId, p.target.TargetId, p.target.TargetName, cacheKey, string(data)))
	}
}

func (p *PIDController) GenScopeConditions() {
	var expressions []Expression
	err := json.Unmarshal([]byte(p.target.TargetScopeFilterJson), &expressions)
	if err != nil {
		log.Error(fmt.Sprintf("module=PIDController\tUnmarshal TargetScopeFilterJson '%s' failed. err=%v",
			p.target.TargetScopeFilterJson, err))
		return
	}
	var planExpressions []Expression
	err = json.Unmarshal([]byte(p.plan.PlanScopeFilterJson), &planExpressions)
	if err != nil {
		log.Error(fmt.Sprintf("module=PIDController\tUnmarshal PlanScopeFilterJson '%s' failed. err=%v",
			p.plan.PlanScopeFilterJson, err))
		return
	}
	expressions = append(expressions, planExpressions...)
	p.scopeConditions = expressions
}

func (p *PIDController) IsControlledItem(ctx *context.RecommendContext, item *module.Item) bool {
	for _, expr := range p.scopeConditions {
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
	} else if scene != p.plan.SceneName {
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
		log.Info(fmt.Sprintf("PIDController <%d/%d>[%s] set runWithZeroInput=%v",
			p.plan.PlanId, p.target.TargetId, p.target.TargetName, run))
		p.runWithZeroInput = run
	}
}

func (p *PIDController) SetMatchConditions(conditions []Expression) {
	p.matchConditions = conditions
	if len(conditions) > 0 {
		log.Info(fmt.Sprintf("PIDController <%d/%d>[%s] set match conditions=%v",
			p.plan.PlanId, p.target.TargetId, p.target.TargetName, p.matchConditions))
	}
}

func ToString(plan interface{}, excludes ...string) string {
	typeOfPlan := reflect.TypeOf(plan)
	valueOfPlan := reflect.ValueOf(plan)
	fields := make(map[string]interface{})
	for i := 0; i < typeOfPlan.NumField(); i++ {
		fieldType := typeOfPlan.Field(i)
		fieldName := fieldType.Name
		if utils.StringContains(excludes, []string{fieldName}) {
			continue
		}
		fieldValue := valueOfPlan.Field(i)
		fields[fieldName] = fieldValue.Interface()
	}
	if jsonStr, err := json.Marshal(fields); err == nil {
		return string(jsonStr)
	}
	return fmt.Sprintf("%v", fields)
}
