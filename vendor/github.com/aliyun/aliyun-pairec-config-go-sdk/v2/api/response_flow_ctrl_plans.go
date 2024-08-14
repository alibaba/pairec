package api

import "github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"

type ListFlowCtrlPlansResponse struct {
	BaseResponse
	Data struct {
		Plans []model.FlowCtrlPlan `json:"plans"`
	} `json:"data,omitempty"`
}
