package domain

type OnlineStore interface {
	GetTableName(featureView *BaseFeatureView) string
	GetDatasourceName() string
	GetSeqOfflineTableName(seqFeatureView *SequenceFeatureView) string
	GetSeqOnlineTableName(seqFeatureView *SequenceFeatureView) string
}
