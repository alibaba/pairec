package module

import (
	gocontext "context"
	"database/sql"
	"fmt"
	"math/cmplx"
	gosort "sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/huandu/go-sqlbuilder"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

var (
	govaluateFunctions = map[string]govaluate.ExpressionFunction{
		"exp": func(args ...interface{}) (interface{}, error) {
			v := utils.ToFloat(args[0], 0)
			return real(cmplx.Exp(complex(v, 0))), nil
		},
	}
)
var (
	weight_mode_sum = "sum"
	weight_mode_max = "max"
)

type RealtimeUser2ItemHologresDao struct {
	*RealtimeUser2ItemBaseDao
	hasPlayTimeField          bool
	itemCount                 int
	db                        *sql.DB
	userTriggerTable          string
	whereClause               string
	itemTable                 string
	weightEvaluableExpression *govaluate.EvaluableExpression
	weightMode                string
	mu                        sync.RWMutex
	userStmt                  *sql.Stmt
	itemStmtMap               map[int]*sql.Stmt
}

func NewRealtimeUser2ItemHologresDao(config recconf.RecallConfig) *RealtimeUser2ItemHologresDao {
	dao := &RealtimeUser2ItemHologresDao{
		itemCount:                config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.ItemCount,
		userTriggerTable:         config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.HologresTableName,
		whereClause:              config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.WhereClause,
		hasPlayTimeField:         true,
		itemTable:                config.RealTimeUser2ItemDaoConf.Item2ItemTable,
		itemStmtMap:              make(map[int]*sql.Stmt, 0),
		weightMode:               config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.WeightMode,
		RealtimeUser2ItemBaseDao: NewRealtimeUser2ItemBaseDao(&config),
	}
	if config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.NoUsePlayTimeField {
		dao.hasPlayTimeField = false
	}

	hologres, err := holo.GetPostgres(config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.HologresName)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}
	dao.db = hologres.DB

	expression, err := govaluate.NewEvaluableExpressionWithFunctions(config.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.WeightExpression,
		govaluateFunctions)
	if err != nil {
		log.Error(fmt.Sprintf("error=%v", err))
		return nil
	}

	dao.weightEvaluableExpression = expression

	if dao.weightMode == "" {
		dao.weightMode = weight_mode_sum
	}

	return dao
}

func (d *RealtimeUser2ItemHologresDao) getItemStmt(key int) *sql.Stmt {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.itemStmtMap[key]
}
func (d *RealtimeUser2ItemHologresDao) ListItemsByUser(user *User, context *context.RecommendContext) (ret []*Item) {

	itemTriggers := d.GetTriggers(user, context)
	if len(itemTriggers) == 0 {
		return
	}

	var itemIds []interface{}
	for id := range itemTriggers {
		itemIds = append(itemIds, id)
	}

	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("item_id", "similar_item_ids").
		From(d.itemTable).
		Where(
			sb.In("item_id", itemIds...),
		)
	sql, args := sb.Build()

	stmtkey := len(itemIds)
	stmt := d.getItemStmt(stmtkey)
	if stmt == nil {
		d.mu.Lock()
		stmt = d.itemStmtMap[stmtkey]
		if stmt == nil {
			stmt2, err := d.db.Prepare(sql)
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=RealtimeUser2ItemHologresDao\terror=hologres error(%v)", context.RecommendId, err))
				return
			}
			d.itemStmtMap[stmtkey] = stmt2
			stmt = stmt2
			d.mu.Unlock()
		} else {
			d.mu.Unlock()
		}
	}
	ctx, cancel := gocontext.WithTimeout(gocontext.Background(), 100*time.Millisecond)
	defer cancel()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=RealtimeUser2ItemHologresDao\tsql=%s\terror=%v", context.RecommendId, sql, err))
		return
	}

	defer rows.Close()
	for rows.Next() {
		var triggerId, ids string
		if err := rows.Scan(&triggerId, &ids); err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=RealtimeUser2ItemHologresDao\terror=%v", context.RecommendId, err))
			continue
		}

		preferScore := itemTriggers[triggerId]

		list := strings.Split(ids, ",")
		for _, str := range list {
			strs := strings.Split(str, ":")
			if len(strs) == 2 && len(strs[0]) > 0 && strs[0] != "null" {
				item := NewItem(strs[0])
				item.RetrieveId = d.recallName
				if tmpScore, err := strconv.ParseFloat(strings.TrimSpace(strs[1]), 64); err == nil {
					item.Score = tmpScore * preferScore
				} else {
					item.Score = preferScore
				}

				ret = append(ret, item)
			}
		}
	}

	gosort.Sort(gosort.Reverse(ItemScoreSlice(ret)))
	ret = uniqItems(ret)

	if len(ret) > d.recallCount {
		ret = ret[:d.recallCount]
	}

	return
}

func (d *RealtimeUser2ItemHologresDao) GetTriggerInfos(user *User, context *context.RecommendContext) (triggerInfos []*TriggerInfo) {
	itemTriggerMap := make(map[string]*TriggerInfo, d.limit)
	var selectFields []string
	if d.hasPlayTimeField {
		selectFields = []string{"item_id", "event", "play_time", "timestamp"}
	} else {
		selectFields = []string{"item_id", "event", "timestamp"}
	}
	if len(d.propertyFields) > 0 {
		selectFields = append(selectFields, d.propertyFields...)
	}
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select(selectFields...)
	builder.From(d.userTriggerTable)
	where := []string{builder.Equal("user_id", string(user.Id))}
	if d.whereClause != "" {
		where = append(where, d.whereClause)
	}
	builder.Where(where...).Limit(d.limit)
	builder.OrderBy("timestamp").Desc()

	sqlquery, args := builder.Build()
	if d.userStmt == nil {
		d.mu.Lock()
		if d.userStmt == nil {
			stmt, err := d.db.Prepare(sqlquery)
			if err != nil {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=RealtimeUser2ItemHologresDao\terror=hologres error(%v)", context.RecommendId, err))
				d.mu.Unlock()
				return
			}
			d.userStmt = stmt
			d.mu.Unlock()
		} else {
			d.mu.Unlock()
		}
	}
	ctx, cancel := gocontext.WithTimeout(gocontext.Background(), 100*time.Millisecond)
	defer cancel()
	rows, err := d.userStmt.QueryContext(ctx, args...)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=RealtimeUser2ItemHologresDao\terror=hologres error(%v)", context.RecommendId, err))
		return
	}

	defer rows.Close()
	currentTime := time.Now()
	for rows.Next() {
		trigger := new(TriggerInfo)
		var dst []interface{}
		if d.hasPlayTimeField {
			dst = []interface{}{&trigger.ItemId, &trigger.event, &trigger.playTime, &trigger.timestamp}
		} else {
			dst = []interface{}{&trigger.ItemId, &trigger.event, &trigger.timestamp}
		}
		if len(d.propertyFields) > 0 {
			trigger.propertyFieldValues = make([]sql.NullString, len(d.propertyFields))
			for i := range trigger.propertyFieldValues {
				dst = append(dst, &trigger.propertyFieldValues[i])
			}
		}
		if err := rows.Scan(dst...); err == nil {
			if t, exist := d.eventPlayTimeMap[trigger.event]; exist {
				if trigger.playTime <= t {
					continue
				}
			}

			weightScore := float64(1)
			if score, ok := d.eventWeightMap[trigger.event]; ok {
				weightScore = score
			}

			eventScore := float64(0)
			properties := map[string]interface{}{
				"currentTime": float64(currentTime.Unix()),
				"eventTime":   float64(trigger.timestamp),
			}

			if result, err := d.weightEvaluableExpression.Evaluate(properties); err == nil {
				if value, ok := result.(float64); ok {
					eventScore = value
				}
			}

			weight := weightScore * eventScore

			if info, exist := itemTriggerMap[trigger.ItemId]; exist {
				switch d.weightMode {
				case weight_mode_max:
					if weight > info.Weight {
						info.Weight = weight
					}
				default:
					info.Weight += weight
				}
			} else {
				trigger.Weight = weight
				itemTriggerMap[trigger.ItemId] = trigger
			}

			//fmt.Println(trigger.itemId, itemTriggerMap[trigger.itemId])
			//itemTriggers[trigger.itemId] += weight
		} else {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=RealtimeUser2ItemHologresDao\terror=hologres error(%v)", context.RecommendId, err))
		}
	}

	for _, triggerInfo := range itemTriggerMap {
		triggerInfos = append(triggerInfos, triggerInfo)
	}
	gosort.Sort(gosort.Reverse(TriggerInfoSlice(triggerInfos)))

	triggerInfos = d.DiversityTriggers(triggerInfos)

	if len(triggerInfos) > d.triggerCount {
		triggerInfos = triggerInfos[:d.triggerCount]
	}

	return
}
func (d *RealtimeUser2ItemHologresDao) GetTriggers(user *User, context *context.RecommendContext) (itemTriggers map[string]float64) {
	triggerInfos := d.GetTriggerInfos(user, context)
	itemTriggers = make(map[string]float64, len(triggerInfos))

	for _, trigger := range triggerInfos {
		itemTriggers[trigger.ItemId] = trigger.Weight
	}

	return
}
