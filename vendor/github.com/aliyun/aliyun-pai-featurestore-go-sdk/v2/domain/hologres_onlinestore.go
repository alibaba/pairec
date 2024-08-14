package domain

import (
	"fmt"
	"strings"

	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/api"
)

type HologresOnlineStore struct {
	*api.Datasource
}

func (s *HologresOnlineStore) GetTableName(featureView *BaseFeatureView) string {
	project := featureView.Project
	tableName := fmt.Sprintf("%s_%s_online", project.ProjectName, featureView.Name)
	return strings.ToLower(tableName)
}

func (s *HologresOnlineStore) GetDatasourceName() string {
	return s.Name
}

func (s *HologresOnlineStore) GetSeqOfflineTableName(seqFeatureView *SequenceFeatureView) string {
	project := seqFeatureView.Project
	tableName := fmt.Sprintf("%s_%s_seq_offline", project.ProjectName, seqFeatureView.Name)
	return strings.ToLower(tableName)
}

func (s *HologresOnlineStore) GetSeqOnlineTableName(sequenceFeatureView *SequenceFeatureView) string {
	project := sequenceFeatureView.Project
	tableName := fmt.Sprintf("%s_%s_seq", project.ProjectName, sequenceFeatureView.Name)
	return strings.ToLower(tableName)
}
