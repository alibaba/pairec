package experiments

import (
	"fmt"
	"time"

	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/api"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/common"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
	"github.com/antihax/optional"
)

func (e *ExperimentClient) logError(err error) {
	if e.ErrorLogger != nil {
		e.ErrorLogger.Printf(err.Error())
		return
	}

	if e.Logger != nil {
		e.Logger.Printf(err.Error())
	}
}

// LoadExperimentData specifies a function to load data from A/B Test Server
func (e *ExperimentClient) LoadExperimentData() {
	sceneData := make(map[string]*model.Scene, 0)

	listScenesResponse, err := e.APIClient.SceneApi.ListAllScenes()
	if err != nil {
		e.logError(fmt.Errorf("list scenes error, err=%v", err))
		return
	}

	for _, scene := range listScenesResponse.Scenes {
		listExpRoomsResponse, err := e.APIClient.ExperimentRoomApi.ListExperimentRooms(e.Environment,
			&api.ExperimentRoomApiListExperimentRoomsOpts{SceneId: optional.NewInt64(scene.SceneId), Status: optional.NewUint32(common.ExpRoom_Status_Online)})

		if err != nil {
			e.logError(fmt.Errorf("list experiment rooms error, err=%v", err))
			return
		}
		for _, experimentRoom := range listExpRoomsResponse.ExperimentRooms {
			if experimentRoom.DebugCrowdId != 0 {
				listCrowdUsersResponse, err := e.APIClient.CrowdApi.GetCrowdUsersById(experimentRoom.DebugCrowdId)
				if err != nil {
					e.logError(fmt.Errorf("list crowd users error, err=%v", err))
					return
				}
				experimentRoom.DebugCrowdIdUsers = listCrowdUsersResponse.Users
			}
			// ExperimentRoom init
			if err := experimentRoom.Init(); err != nil {
				e.logError(fmt.Errorf("experiment room init error, err=%v", err))
				return
			}

			scene.AddExperimentRoom(experimentRoom)
			listLayersResponse, err := e.APIClient.LayerApi.ListLayers(experimentRoom.ExpRoomId)
			if err != nil {
				e.logError(fmt.Errorf("list layers error, err=%v", err))
				return
			}
			for _, layer := range listLayersResponse.Layers {
				experimentRoom.AddLayer(layer)

				listExperimentGroupResponse, err := e.APIClient.ExperimentGroupApi.ListExperimentGroups(layer.LayerId,
					&api.ExperimentGroupApiListExperimentGroupsOpts{Status: optional.NewUint32(common.ExpGroup_Status_Online)})
				if err != nil {
					e.logError(fmt.Errorf("list experiment groups error, err=%v", err))
					return
				}

				for _, experimentGroup := range listExperimentGroupResponse.ExperimentGroups {
					if experimentGroup.CrowdId != 0 {
						listCrowdUsersResponse, err := e.APIClient.CrowdApi.GetCrowdUsersById(experimentGroup.CrowdId)
						if err != nil {
							e.logError(fmt.Errorf("list crowd users error, err=%v", err))
							return
						}
						experimentGroup.CrowdUsers = listCrowdUsersResponse.Users
					}

					if experimentGroup.DebugCrowdId != 0 {
						listCrowdUsersResponse, err := e.APIClient.CrowdApi.GetCrowdUsersById(experimentGroup.DebugCrowdId)
						if err != nil {
							e.logError(fmt.Errorf("list crowd users error, err=%v", err))
							return
						}
						experimentGroup.DebugCrowdUsers = listCrowdUsersResponse.Users
					}

					// ExperimentGroup init
					if err := experimentGroup.Init(); err != nil {
						e.logError(fmt.Errorf("experiment group init error, err=%v", err))
						return
					}

					layer.AddExperimentGroup(experimentGroup)

					listExperimentsResponse, err := e.APIClient.ExperimentApi.ListExperiments(experimentGroup.ExpGroupId,
						&api.ExperimentApiListExperimentsOpts{Status: optional.NewUint32(common.Experiment_Status_Online)})
					if err != nil {
						e.logError(fmt.Errorf("list experiments  error, err=%v", err))
						return
					}

					for _, experiment := range listExperimentsResponse.Experiments {
						if experiment.DebugCrowdId != 0 {
							listCrowdUsersResponse, err := e.APIClient.CrowdApi.GetCrowdUsersById(experiment.DebugCrowdId)
							if err != nil {
								e.logError(fmt.Errorf("list crowd users error, err=%v", err))
								return
							}
							experiment.DebugCrowdUsers = listCrowdUsersResponse.Users
						}
						if err := experiment.Init(); err != nil {
							e.logError(fmt.Errorf("experiment init error, err=%v", err))
							return
						}
						experimentGroup.AddExperiment(experiment)
					}
				}
			}
		}
		sceneData[scene.SceneName] = scene
	}
	if len(sceneData) > 0 {
		e.sceneMap = sceneData
	}
}

// loopLoadExperimentData async loop invoke LoadExperimentData function
func (e *ExperimentClient) loopLoadExperimentData() {

	for {
		time.Sleep(time.Minute)
		e.LoadExperimentData()
	}
}
