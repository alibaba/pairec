package api

type ListModelsResponse struct {
	TotalCount int      `json:"total_count,omitempty"`
	Models     []*Model `json:"models,omitempty"`
}

type GetModelResponse struct {
	Model *Model `json:"model,omitempty"`
}
