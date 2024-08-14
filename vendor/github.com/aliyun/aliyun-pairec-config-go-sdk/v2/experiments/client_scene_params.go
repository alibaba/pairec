package experiments

import (
	"fmt"
	"time"

	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/api"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
	"github.com/antihax/optional"
)

// LoadSceneParamsData specifies a function to load param data from A/B Test Server
func (e *ExperimentClient) LoadSceneParamsData() {
	sceneParamData := make(map[string]model.SceneParams, 0)

	listScenesResponse, err := e.APIClient.SceneApi.ListAllScenes()
	if err != nil {
		e.logError(fmt.Errorf("list scenes error, err=%v", err))
		return
	}

	for _, scene := range listScenesResponse.Scenes {
		sceneParams := model.NewSceneParams()
		listParamsResponse, err := e.APIClient.ParamApi.GetParam(scene.SceneId,
			&api.ParamApiGetParamOpts{Environment: optional.NewString(e.Environment)})

		if err != nil {
			e.logError(fmt.Errorf("list params error, err=%v", err))
			continue
		}
		for _, param := range listParamsResponse.Params {
			sceneParams.AddParam(param.ParamName, param.ParamValue)
		}
		sceneParamData[scene.SceneName] = sceneParams
	}
	if len(sceneParamData) > 0 {
		e.sceneParamData = sceneParamData
	}
}

// loopLoadExperimentData async loop invoke LoadExperimentData function
func (e *ExperimentClient) loopLoadSceneParamsData() {

	for {
		time.Sleep(time.Minute)
		e.LoadSceneParamsData()
	}
}

func (e *ExperimentClient) GetSceneParams(sceneName string) model.SceneParams {
	sceneParams, ok := e.sceneParamData[sceneName]
	if !ok {
		return model.NewEmptySceneParams()
	}

	return sceneParams
}
