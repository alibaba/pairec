package model

type Scene struct {
	SceneId int64 `json:"scene_id,omitempty"`
	SceneName string `json:"scene_name"`
	SceneInfo string `json:"scene_info"`
	ExperimentRooms []*ExperimentRoom `json:"experiment_rooms"`
}

func (s *Scene) AddExperimentRoom(room *ExperimentRoom) {
	s.ExperimentRooms = append(s.ExperimentRooms, room)
}
