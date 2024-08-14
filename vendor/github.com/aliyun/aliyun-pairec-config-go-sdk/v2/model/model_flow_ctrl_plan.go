package model

import "time"

type FlowCtrlPlan struct {
	PlanId                        int                   `json:"plan_id"`
	SceneId                       int                   `json:"scene_id"`
	SceneName                     string                `json:"scene_name"`
	PlanName                      string                `json:"plan_name"`
	PlanDesc                      string                `json:"plan_desc"`
	OnlineDatasourceType          string                `json:"online_datasource_type"`
	OnlineDatasourceId            int                   `json:"online_datasource_id"`
	OnlineTableName               string                `json:"online_table_name"`
	OnlineTableItemIdField        string                `json:"online_table_item_id_field"`
	PlanScopeFilter               string                `json:"plan_scope_filter"`
	PlanScopeFilterJson           string                `json:"plan_scope_filter_json"`
	TargetValueInPercentageFormat bool                  `json:"target_value_in_percentage_format"`
	PlanType                      string                `json:"plan_type"`
	Granularity                   string                `json:"granularity"`
	RealtimeLogType               string                `json:"realtime_log_type"`
	RealtimeLogTableMetaId        int                   `json:"realtime_log_table_meta_id"`
	RealtimeLogFilter             string                `json:"RealtimeLogFilter"`
	RealtimeLogFilterJson         string                `json:"realtime_log_filter_json"`
	FlowScopeFilterJson           string                `json:"flow_scope_filter_json"`
	LoadTrafficByPlan             bool                  `json:"load_traffic_by_plan"`
	StartTime                     time.Time             `json:"start_time"`
	EndTime                       time.Time             `json:"end_time"`
	Status                        string                `json:"status"`
	CreateTime                    time.Time             `json:"create_time"`
	Targets                       []FlowCtrlPlanTargets `json:"targets"`
}

type FlowCtrlPlanTargets struct {
	TargetId              int                `json:"target_id"`
	PlanId                int                `json:"planId"`
	TargetName            string             `json:"target_name"`
	TargetType            int                `json:"target_type"`
	TargetScopeFilter     string             `json:"target_scope_filter"`
	TargetScopeFilterJson string             `json:"target_scope_filter_json"`
	ItemScopeFilterJson   string             `json:"item_scope_filter_json"`
	TimeUint              string             `json:"time_uint"`
	SetPoint              float64            `json:"set_point"`
	SetPointRange         float64            `json:"set_point_range"`
	DoRecall              bool               `json:"do_recall"`
	Status                string             `json:"status"`
	StartTime             time.Time          `json:"start_time"`
	EndTime               time.Time          `json:"end_time"`
	CreateTime            time.Time          `json:"create_time"`
	TargetTraffics        map[string]float64 `json:"target_traffics"`
	PlanTraffic           map[string]float64 `json:"plan_traffics"`
	TimePoints            []int              `json:"time_points"`
	SetPoints             []float64          `json:"set_points"`
}
