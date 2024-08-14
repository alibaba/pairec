package experiments

import (
	"time"

	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
)

/**
// LoadSceneFlowCtrlPlansData specifies a function to load flow ctrl plan data from A/B Test Server
func (e *ExperimentClient) LoadSceneFlowCtrlPlansData() {
	sceneFlowCtrlPlanData := make(map[string][]model.FlowCtrlPlan, 0)

	opt := &api.FlowCtrlApiListFlowCtrlPlansOpts{}
	opt.Env = optional.NewString("product")

	listPlansResponse, err := e.APIClient.FlowCtrlApi.ListFlowCtrlPlans(context.Background(), opt)
	if err != nil {
		e.logError(fmt.Errorf("list flow plans error, err=%v", err))
		return
	}

	if listPlansResponse.Code != common.CODE_OK {
		e.logError(fmt.Errorf("list flow plans error, requestid=%s,code=%s, msg=%s", listPlansResponse.RequestId, listPlansResponse.Code, listPlansResponse.Message))
		return
	}

	for _, plan := range listPlansResponse.Data.Plans {
		sceneFlowCtrlPlanData[plan.SceneName] = append(sceneFlowCtrlPlanData[plan.SceneName], plan)
	}
	if len(sceneFlowCtrlPlanData) > 0 {
		e.sceneFlowCtrlPlanData = sceneFlowCtrlPlanData
	}

	prepubSceneFlowCtrlPlanData := make(map[string][]model.FlowCtrlPlan, 0)
	opt.Env = optional.NewString("prepub")

	listPlansResponse, err = e.APIClient.FlowCtrlApi.ListFlowCtrlPlans(context.Background(), opt)
	if err != nil {
		e.logError(fmt.Errorf("list flow plans error, err=%v", err))
		return
	}

	if listPlansResponse.Code != common.CODE_OK {
		e.logError(fmt.Errorf("list flow plans error, requestid=%s,code=%s, msg=%s", listPlansResponse.RequestId, listPlansResponse.Code, listPlansResponse.Message))
		return
	}

	for _, plan := range listPlansResponse.Data.Plans {
		prepubSceneFlowCtrlPlanData[plan.SceneName] = append(prepubSceneFlowCtrlPlanData[plan.SceneName], plan)
	}
	if len(prepubSceneFlowCtrlPlanData) > 0 {
		e.prepubSceneFlowCtrlPlanData = prepubSceneFlowCtrlPlanData
	}
}

// loopLoadSceneFlowCtrlPlansData async loop invoke LoadSceneFlowCtrlPlansData function
func (e *ExperimentClient) loopLoadSceneFlowCtrlPlansData() {

	for {
		time.Sleep(time.Second * 30)
		e.LoadSceneFlowCtrlPlansData()
	}
}
**/

func (e *ExperimentClient) GetFlowCtrlPlanTargetList(env, sceneName string, currentTimestamp int64) map[int]model.FlowCtrlPlanTargets {
	if currentTimestamp == 0 {
		currentTimestamp = time.Now().Unix()
	}

	targetsMap := make(map[int]model.FlowCtrlPlanTargets)

	data := e.sceneFlowCtrlPlanData
	if env == "prepub" {
		data = e.prepubSceneFlowCtrlPlanData
	}

	for scene, scenePlans := range data {
		if sceneName != "" && sceneName != scene {
			continue
		}

		for _, plan := range scenePlans {
			for i, target := range plan.Targets {
				if target.Status == "enable" && target.StartTime.Unix() < currentTimestamp && currentTimestamp <= target.EndTime.Unix() {
					targetsMap[target.TargetId] = plan.Targets[i]
				}
			}
		}
	}

	return targetsMap
}

func (e *ExperimentClient) GetFlowCtrlPlanMetaList(env string, currentTimestamp int64) []model.FlowCtrlPlan {
	if currentTimestamp == 0 {
		currentTimestamp = time.Now().Unix()
	}

	plans := make([]model.FlowCtrlPlan, 0)

	data := e.sceneFlowCtrlPlanData
	if env == "prepub" {
		data = e.prepubSceneFlowCtrlPlanData
	}

	for _, scenePlans := range data {
		for i, plan := range scenePlans {
			if plan.Status == "enable" && plan.StartTime.Unix() <= currentTimestamp && currentTimestamp < plan.EndTime.Unix() {
				plans = append(plans, scenePlans[i])
			}
		}
	}
	return plans
}

func (e *ExperimentClient) CheckIfFlowCtrlPlanTargetIsEnabled(env string, targetId int, currentTimestamp int64) bool {
	if currentTimestamp == 0 {
		currentTimestamp = time.Now().Unix()
	}

	data := e.sceneFlowCtrlPlanData
	if env == "prepub" {
		data = e.prepubSceneFlowCtrlPlanData
	}

	for _, scenePlans := range data {
		for _, plan := range scenePlans {
			for _, target := range plan.Targets {
				if target.TargetId == targetId {
					if target.Status == "enable" && target.StartTime.Unix() < currentTimestamp && currentTimestamp < target.EndTime.Unix() {
						return true
					}
				}
			}
		}
	}
	return false
}

type FlowCtrlPlanTargetTraffic struct {
	ItemOrExpId   string  `json:"item_or_exp_id"`
	PlanId        int     `json:"plan_id"`
	TargetId      int     `json:"target_id"`
	TargetTraffic float64 `json:"target_traffic"`
	PlanTraffic   float64 `json:"plan_traffic"`
}

func (e *ExperimentClient) GetFlowCtrlPlanTargetTraffic(env, sceneName string, idList ...string) []FlowCtrlPlanTargetTraffic {
	targets := e.GetFlowCtrlPlanTargetList(env, sceneName, 0)

	var traffics []FlowCtrlPlanTargetTraffic

	idMap := make(map[string]bool)
	for _, id := range idList {
		idMap[id] = true
	}

	for _, planTarget := range targets {
		for id, value := range planTarget.TargetTraffics {
			if len(idList) == 0 || idMap[id] {
				traffics = append(traffics, FlowCtrlPlanTargetTraffic{
					ItemOrExpId:   id,
					PlanId:        planTarget.PlanId,
					TargetId:      planTarget.TargetId,
					TargetTraffic: value,
					PlanTraffic:   planTarget.PlanTraffic[id],
				})
			}
		}
	}

	return traffics
}
