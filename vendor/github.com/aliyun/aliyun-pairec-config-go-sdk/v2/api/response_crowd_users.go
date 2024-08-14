package api

type ListCrowdUsersResponse struct {
	Users      []string `json:"Users,omitempty"`
	CrowdUsers []string `json:"CrowdUsers,omitempty"`
}
