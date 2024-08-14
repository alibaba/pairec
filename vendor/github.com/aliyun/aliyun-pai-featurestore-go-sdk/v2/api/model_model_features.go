package api

type ModelFeatures struct {
	FeatureViewId   int    `json:"feature_view_id,omitempty"`
	FeatureViewName string `json:"feature_view_name,omitempty"`
	Name            string `json:"name,omitempty"`
	AliasName       string `json:"alias_name,omitempty"`
	Type            int32  `json:"type,omitempty"`
	TypeStr         string `json:"type_str,omitempty"`
}
