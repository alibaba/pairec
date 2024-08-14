package dao

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/api"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/constants"
)

type FeatureViewDao interface {
	GetFeatures(keys []interface{}, selectFields []string) ([]map[string]interface{}, error)
	GetUserSequenceFeature(keys []interface{}, userIdField string, sequenceConfig api.FeatureViewSeqConfig, onlineConfig []*api.SeqConfig) ([]map[string]interface{}, error)
}

func NewFeatureViewDao(config DaoConfig) FeatureViewDao {
	if config.DatasourceType == constants.Datasource_Type_Hologres {
		return NewFeatureViewHologresDao(config)
	} else if config.DatasourceType == constants.Datasource_Type_Mysql {
		return NewFeatureViewMysqlDao(config)
	} else if config.DatasourceType == constants.Datasource_Type_IGraph {
		return NewFeatureViewIGraphDao(config)
	} else if config.DatasourceType == constants.Datasource_Type_Redis {
		return NewFeatureViewRedisDao(config)
	} else if config.DatasourceType == constants.Datasource_Type_TableStore {
		return NewFeatureViewTableStoreDao(config)
	}

	panic("not found FeatureViewDao implement")
}

func makePlayTimeMap(sequenceConfig api.FeatureViewSeqConfig) map[string]float64 {
	sequencePlayTimeMap := make(map[string]float64)
	if sequenceConfig.PlayTimeFilter != "" {
		playTimes := strings.Split(sequenceConfig.PlayTimeFilter, ";")
		for _, eventTime := range playTimes {
			strs := strings.Split(eventTime, ":")
			if len(strs) == 2 {
				if t, err := strconv.ParseFloat(strs[1], 64); err == nil {
					sequencePlayTimeMap[strs[0]] = t
				}
			}
		}
	}

	return sequencePlayTimeMap
}

func makeSequenceFeatures(offlineSequences, onlineSequences []*sequenceInfo, seqConfig *api.SeqConfig, sequenceConfig api.FeatureViewSeqConfig, currTime int64) map[string]interface{} {
	//combine offlineSequences and onlineSequences
	if len(offlineSequences) > 0 {
		index := 0
		for index < len(onlineSequences) {
			if onlineSequences[index].timestamp < offlineSequences[0].timestamp {
				break
			}
			index++
		}

		onlineSequences = onlineSequences[:index]
		onlineSequences = append(onlineSequences, offlineSequences...)
		if len(onlineSequences) > seqConfig.SeqLen {
			onlineSequences = onlineSequences[:seqConfig.SeqLen]
		}
	}

	//produce seqeunce feature correspond to easyrec processor
	sequencesValueMap := make(map[string][]string)
	sequenceMap := make(map[string]bool, 0)

	for _, seq := range onlineSequences {
		key := fmt.Sprintf("%s#%s", seq.itemId, seq.event)
		if _, exist := sequenceMap[key]; !exist {
			sequenceMap[key] = true
			sequencesValueMap[sequenceConfig.ItemIdField] = append(sequencesValueMap[sequenceConfig.ItemIdField], seq.itemId)
			sequencesValueMap[sequenceConfig.TimestampField] = append(sequencesValueMap[sequenceConfig.TimestampField], fmt.Sprintf("%d", seq.timestamp))
			sequencesValueMap[sequenceConfig.EventField] = append(sequencesValueMap[sequenceConfig.EventField], seq.event)
			if sequenceConfig.PlayTimeField != "" {
				sequencesValueMap[sequenceConfig.PlayTimeField] = append(sequencesValueMap[sequenceConfig.PlayTimeField], fmt.Sprintf("%.2f", seq.playTime))
			}
			sequencesValueMap["ts"] = append(sequencesValueMap["ts"], fmt.Sprintf("%d", currTime-seq.timestamp))
		}
	}

	properties := make(map[string]interface{})
	for key, value := range sequencesValueMap {
		curSequenceSubName := (seqConfig.OnlineSeqName + "__" + key)
		properties[curSequenceSubName] = strings.Join(value, ";")
	}
	properties[seqConfig.OnlineSeqName] = strings.Join(sequencesValueMap[sequenceConfig.ItemIdField], ";")

	return properties

}
