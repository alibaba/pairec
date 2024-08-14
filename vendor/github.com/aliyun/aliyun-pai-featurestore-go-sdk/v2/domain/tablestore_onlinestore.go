package domain

import (
	"fmt"

	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/api"
)

type TableStoreOnlineStore struct {
	*api.Datasource
}

func (s *TableStoreOnlineStore) GetTableName(featureView *BaseFeatureView) string {
	project := featureView.Project
	return fmt.Sprintf("%s_%s_online", project.ProjectName, featureView.Name)
}

func (s *TableStoreOnlineStore) GetDatasourceName() string {
	return s.Name
}

func (s *TableStoreOnlineStore) GetSeqOfflineTableName(sequenceFeatureView *SequenceFeatureView) string {
	project := sequenceFeatureView.Project
	return fmt.Sprintf("%s_%s_seq_offline", project.ProjectName, sequenceFeatureView.Name)
}

func (s *TableStoreOnlineStore) GetSeqOnlineTableName(sequenceFeatureView *SequenceFeatureView) string {
	project := sequenceFeatureView.Project
	return fmt.Sprintf("%s_%s_seq", project.ProjectName, sequenceFeatureView.Name)
}
