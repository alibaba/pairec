package api

type ListFeatureViewsResponse struct {
	TotalCount   int            `json:"total_count"`
	FeatureViews []*FeatureView `json:"feature_views"`
}

type GetFeatureViewResponse struct {
	FeatureView *FeatureView
}
