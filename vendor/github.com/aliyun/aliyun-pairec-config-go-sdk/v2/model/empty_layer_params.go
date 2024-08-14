package model

type EmptyLayerParams struct {
}

func NewEmptyLayerParams() *EmptyLayerParams {
	return &EmptyLayerParams{}
}

func (r *EmptyLayerParams) AddParam(key string, value interface{}) {}

func (r *EmptyLayerParams) AddParams(params map[string]interface{}) {}

func (r *EmptyLayerParams) Get(key string, defaultValue interface{}) interface{} {
	return defaultValue
}

func (r *EmptyLayerParams) GetString(key, defaultValue string) string {
	return defaultValue
}

func (r *EmptyLayerParams) GetInt(key string, defaultValue int) int {
	return defaultValue
}
func (r *EmptyLayerParams) GetFloat(key string, defaultValue float64) float64 {
	return defaultValue
}
func (r *EmptyLayerParams) GetInt64(key string, defaultValue int64) int64 {
	return defaultValue
}
