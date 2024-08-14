package model

type Layer struct {
	LayerId   int64  `json:"layer_id,omitempty"`
	ExpRoomId int64  `json:"exp_room_id"`
	SceneId   int64  `json:"scene_id"`
	LayerName string `json:"layer_name"`
	LayerInfo string `json:"layer_info"`

	ExperimentGroups []*ExperimentGroup `json:"experiment_groups"`
}

func (l *Layer) AddExperimentGroup(g *ExperimentGroup) {
	l.ExperimentGroups = append(l.ExperimentGroups, g)
}
