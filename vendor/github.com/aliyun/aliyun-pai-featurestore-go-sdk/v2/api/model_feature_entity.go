package api

type FeatureEntity struct {
	FeatureEntityId     int    `json:"feature_entity_id,omitempty"`
	ProjectId           int    `json:"project_id"`
	ProjectName         string `json:"project_name,omitempty"`
	FeatureEntityName   string `json:"feature_entity_name"`
	FeatureEntityJoinid string `json:"feature_entity_joinid"`
}
