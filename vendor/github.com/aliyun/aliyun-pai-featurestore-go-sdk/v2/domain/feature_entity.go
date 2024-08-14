package domain

import (
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/api"
)

type FeatureEntity struct {
	*api.FeatureEntity
}

func NewFeatureEntity(entity *api.FeatureEntity) *FeatureEntity {
	featureEntity := &FeatureEntity{
		FeatureEntity: entity,
	}
	return featureEntity
}
