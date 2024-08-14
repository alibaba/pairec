package domain

import (
	"encoding/json"
	"fmt"

	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/api"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/constants"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/dao"
)

type SequenceFeatureView struct {
	*api.FeatureView
	Project                  *Project
	FeatureEntity            *FeatureEntity
	userIdField              string
	sequenceConfig           api.FeatureViewSeqConfig
	featureViewDao           dao.FeatureViewDao
	offline_2_online_seq_map map[string]string
}

func NewSequenceFeatureView(view *api.FeatureView, p *Project, entity *FeatureEntity) *SequenceFeatureView {
	sequenceFeatureView := &SequenceFeatureView{
		FeatureView:   view,
		Project:       p,
		FeatureEntity: entity,
	}
	for _, field := range view.Fields {
		if field.IsPrimaryKey {
			sequenceFeatureView.userIdField = field.Name
			break
		}
	}

	err := json.Unmarshal([]byte(view.Config), &sequenceFeatureView.sequenceConfig)
	if err != nil {
		panic("sequence featureview config unmarshal failed")
	}

	sequenceFeatureView.offline_2_online_seq_map = make(map[string]string, len(sequenceFeatureView.sequenceConfig.SeqConfig))

	for _, seqConfig := range sequenceFeatureView.sequenceConfig.SeqConfig {
		sequenceFeatureView.offline_2_online_seq_map[seqConfig.OfflineSeqName] = seqConfig.OnlineSeqName
	}

	requiredElements1 := []string{"user_id", "item_id", "event"}
	requiredElements2 := []string{"user_id", "item_id", "event", "timestamp"}
	if len(sequenceFeatureView.sequenceConfig.DeduplicationMethod) == len(requiredElements1) {
		for i, v := range sequenceFeatureView.sequenceConfig.DeduplicationMethod {
			if v != requiredElements1[i] {
				panic("deduplication_method invalid")
			}
		}
		sequenceFeatureView.sequenceConfig.DeduplicationMethodNum = 1
	} else if len(sequenceFeatureView.sequenceConfig.DeduplicationMethod) == len(requiredElements2) {
		for i, v := range sequenceFeatureView.sequenceConfig.DeduplicationMethod {
			if v != requiredElements2[i] {
				panic("deduplication_method invalid")
			}
		}
		sequenceFeatureView.sequenceConfig.DeduplicationMethodNum = 2
	} else {
		panic("deduplication_method invalid")
	}

	daoConfig := dao.DaoConfig{
		DatasourceType:  p.OnlineDatasourceType,
		PrimaryKeyField: sequenceFeatureView.userIdField,
	}

	switch p.OnlineDatasourceType {
	case constants.Datasource_Type_Hologres:
		daoConfig.HologresName = p.OnlineStore.GetDatasourceName()
		daoConfig.HologresOfflineTableName = p.OnlineStore.GetSeqOfflineTableName(sequenceFeatureView)
		daoConfig.HologresOnlineTableName = p.OnlineStore.GetSeqOnlineTableName(sequenceFeatureView)
	case constants.Datasource_Type_TableStore:
		daoConfig.TableStoreName = p.OnlineStore.GetDatasourceName()
		daoConfig.TableStoreOfflineTableName = p.OnlineStore.GetSeqOfflineTableName(sequenceFeatureView)
		daoConfig.TableStoreOnlineTableName = p.OnlineStore.GetSeqOnlineTableName(sequenceFeatureView)

	case constants.Datasource_Type_IGraph:
		daoConfig.SaveOriginalField = true
		daoConfig.IGraphName = p.OnlineStore.GetDatasourceName()
		daoConfig.GroupName = p.ProjectName
		daoConfig.IgraphEdgeName = p.OnlineStore.GetSeqOnlineTableName(sequenceFeatureView)

	default:

	}

	featureViewDao := dao.NewFeatureViewDao(daoConfig)
	sequenceFeatureView.featureViewDao = featureViewDao

	return sequenceFeatureView
}

func (f *SequenceFeatureView) GetOnlineFeatures(joinIds []interface{}, features []string, alias map[string]string) ([]map[string]interface{}, error) {
	sequenceConfig := f.sequenceConfig
	onlineConfig := []*api.SeqConfig{}

	for _, feature := range features {
		if feature == "*" {
			onlineConfig = sequenceConfig.SeqConfig
			break
		} else {
			found := false
			for _, seqConfig := range sequenceConfig.SeqConfig {
				if seqConfig.OnlineSeqName == feature {
					found = true
					onlineConfig = append(onlineConfig, seqConfig)
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("sequence feature name :%s not found in feature view config", feature)
			}
		}
	}

	sequenceFeatureResults, err := f.featureViewDao.GetUserSequenceFeature(joinIds, f.userIdField, sequenceConfig, onlineConfig)

	if f.userIdField != f.FeatureEntity.FeatureEntityJoinid {
		for _, sequencefeatureMap := range sequenceFeatureResults {
			sequencefeatureMap[f.FeatureEntity.FeatureEntityJoinid] = sequencefeatureMap[f.userIdField]
			delete(sequencefeatureMap, f.userIdField)
		}
	}

	return sequenceFeatureResults, err
}

func (f *SequenceFeatureView) GetName() string {
	return f.Name
}

func (f *SequenceFeatureView) GetFeatureEntityName() string {
	return f.FeatureEntityName
}

func (f *SequenceFeatureView) GetType() string {
	return f.Type
}

func (f *SequenceFeatureView) Offline2Online(input string) string {
	return f.offline_2_online_seq_map[input]
}
