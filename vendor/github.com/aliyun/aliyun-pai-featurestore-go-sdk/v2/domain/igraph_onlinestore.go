package domain

import (
	"fmt"

	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/api"
)

type IGraphOnlineStore struct {
	*api.Datasource
}

func (s *IGraphOnlineStore) GetTableName(featureView *BaseFeatureView) string {
	return fmt.Sprintf("%s_fv%d", featureView.FeatureEntityName, featureView.FeatureViewId)
}

func (s *IGraphOnlineStore) GetDatasourceName() string {
	return s.Name
}

func (s *IGraphOnlineStore) getSequenceTableName(sequenceFeatureView *SequenceFeatureView) string {
	return fmt.Sprintf("%s_fv%d_seq", sequenceFeatureView.FeatureEntityName, sequenceFeatureView.FeatureViewId)
}

func (s *IGraphOnlineStore) GetSeqOfflineTableName(sequenceFeatureView *SequenceFeatureView) string {
	return s.getSequenceTableName(sequenceFeatureView)
}

func (s *IGraphOnlineStore) GetSeqOnlineTableName(sequenceFeatureView *SequenceFeatureView) string {
	return s.getSequenceTableName(sequenceFeatureView)
}
