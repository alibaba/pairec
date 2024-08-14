package model

type Param struct {
	ParamId     int64  `json:"param_id,omitempty"`
	SceneId     int64  `json:"scene_id"`
	ParamName   string `json:"param_name"`
	ParamValue  string `json:"param_value"`
	Environment int32  `json:"environment"`
	Scene       *Scene `json:"scene"`
}
