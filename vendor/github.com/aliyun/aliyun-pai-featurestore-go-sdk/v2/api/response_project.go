package api

type ListProjectsResponse struct {
	Projects []*Project `json:"data,omitempty"`
}
