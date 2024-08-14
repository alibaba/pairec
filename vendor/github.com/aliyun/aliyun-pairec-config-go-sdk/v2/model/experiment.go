package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/common"
)

type ExperimentContext struct {
	RequestId string

	// Uid
	Uid string

	// FilterParams is map of params, use for filter condition
	FilterParams map[string]interface{}

	experimentHashStr string
}

func (c *ExperimentContext) SetExperimentHashString(str string) {
	c.experimentHashStr = str
}
func (c *ExperimentContext) ExperimentHashString() string {
	return c.experimentHashStr
}

// ExperimentResult is the result match by ExperimentContext
type ExperimentResult struct {
	ExperimentContext *ExperimentContext

	// ExperimentRoom is the result of the match ExperimentRoom
	ExperimentRoom *ExperimentRoom

	// Layers is a list of the layer, the match ExperimentRoom has layers list here
	Layers []*Layer

	// ExpId is path of match experiment ,  from experiment room to experiment
	// Example: ER2_L1#EG1#E2_L2#EG1#E3_L3#EG1#E6
	ExpId string

	SceneName string

	// layer2ExperimentGroup is a map of layerName as key
	layer2ExperimentGroup map[string]*ExperimentGroup

	layer2Experiment map[string]*Experiment

	layerParamsMap map[string]LayerParams

	mergedLayerParams LayerParams
}

func NewExperimentResult(sceneName string, experimentContext *ExperimentContext) *ExperimentResult {
	result := ExperimentResult{
		SceneName:             sceneName,
		ExperimentContext:     experimentContext,
		layer2ExperimentGroup: make(map[string]*ExperimentGroup, 0),
		layer2Experiment:      make(map[string]*Experiment, 0),
		layerParamsMap:        make(map[string]LayerParams, 0),
	}

	return &result
}

func (r *ExperimentResult) LayerSize() int {
	return len(r.Layers)
}
func (r *ExperimentResult) ContainsLayer(layerName string) bool {
	for _, layer := range r.Layers {
		if layer.LayerName == layerName {
			return true
		}
	}
	return false
}

func (r *ExperimentResult) GetExpId() string {
	return r.ExpId
}

func (r *ExperimentResult) AddMatchExperimentGroup(layerName string, experimentGroup *ExperimentGroup) {

	r.layer2ExperimentGroup[layerName] = experimentGroup
}

func (r *ExperimentResult) AddMatchExperiment(layerName string, experiment *Experiment) {

	r.layer2Experiment[layerName] = experiment
}

func (r *ExperimentResult) Init() {
	buf := bytes.NewBuffer(nil)

	if r.ExperimentRoom != nil {
		buf.WriteString("ER")
		buf.WriteString(strconv.Itoa(int(r.ExperimentRoom.ExpRoomId)))
	}
	for _, layer := range r.Layers {
		buf.WriteString("_L")
		buf.WriteString(strconv.Itoa(int(layer.LayerId)))
		if experimentGroup, found := r.layer2ExperimentGroup[layer.LayerName]; found {
			buf.WriteString("#")
			buf.WriteString("EG")
			buf.WriteString(strconv.Itoa(int(experimentGroup.ExpGroupId)))
			layerParams := NewLayerParams()
			params := make(map[string]interface{}, 0)
			if experimentGroup.ExpGroupConfig != "" {
				if err := json.Unmarshal([]byte(experimentGroup.ExpGroupConfig), &params); err == nil {
					layerParams.AddParams(params)
				}
			}
			if experiment, found := r.layer2Experiment[layer.LayerName]; found {
				if experiment.Type != common.Experiment_Type_Default {
					buf.WriteString("#")
					buf.WriteString("E")
					buf.WriteString(strconv.Itoa(int(experiment.ExperimentId)))
				}
				//buf.WriteString("#")
				if experiment.ExperimentConfig != "" {
					if err := json.Unmarshal([]byte(experiment.ExperimentConfig), &params); err == nil {
						layerParams.AddParams(params)
					}
				}
			}
			r.layerParamsMap[layer.LayerName] = layerParams
		}
	}

	id := buf.String()
	if len(id) > 0 {
		if id[len(id)-1] == '#' || id[len(id)-1] == '_' {
			id = id[0 : len(id)-1]
		}
	}

	r.ExpId = id
}

func (r *ExperimentResult) GetLayerParams(layerName string) LayerParams {
	if r.ExperimentRoom == nil || r.LayerSize() == 0 {
		return NewEmptyLayerParams()
	}

	// omit layer name
	if r.LayerSize() == 1 {
		if layerParams, found := r.layerParamsMap[r.Layers[0].LayerName]; found {
			return layerParams
		}
	}

	layerParams, found := r.layerParamsMap[layerName]
	if !found {
		return NewEmptyLayerParams()
	}

	return layerParams
}
func (r *ExperimentResult) Info() string {
	var info []string

	if r.ExperimentContext != nil {
		info = append(info, fmt.Sprintf("requestId=%s", r.ExperimentContext.RequestId))
		info = append(info, fmt.Sprintf("uid=%s", r.ExperimentContext.Uid))
	}
	info = append(info, fmt.Sprintf("scene_name=%s", r.SceneName))
	if r.ExperimentRoom != nil {
		info = append(info, fmt.Sprintf("exp_room_id=%d", r.ExperimentRoom.ExpRoomId))
	}
	info = append(info, fmt.Sprintf("exp_id=%s", r.ExpId))

	return strings.Join(info, ",")
}

func (r *ExperimentResult) GetExperimentParams() LayerParams {
	if r.mergedLayerParams == nil {
		r.mergedLayerParams = MergeLayerParams(r.layerParamsMap)
	}
	return r.mergedLayerParams
}
