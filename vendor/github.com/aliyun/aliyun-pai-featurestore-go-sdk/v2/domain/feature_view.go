package domain

import (
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/api"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/constants"
)

type FeatureView interface {
	GetOnlineFeatures(joinIds []interface{}, features []string, alias map[string]string) ([]map[string]interface{}, error)
	GetName() string
	GetFeatureEntityName() string
	GetType() string
	Offline2Online(input string) string
}

func NewFeatureView(view *api.FeatureView, p *Project, entity *FeatureEntity) FeatureView {
	if view.Type == constants.Feature_View_Type_Sequence {
		return NewSequenceFeatureView(view, p, entity)
	} else {
		return NewBaseFeatureView(view, p, entity)
	}
}
