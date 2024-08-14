package api

type ListFeatureEntitiesResponse struct {
	FeatureEntities []*FeatureEntity
}

type GetFeatureEntityResponse struct {
	FeatureEntity *FeatureEntity
}
