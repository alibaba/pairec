package experiments

import (
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/api"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
)

func (e *ExperimentClient) BackflowFeatureConsistencyCheckJobData(backflowData *model.FeatureConsistencyBackflowData) (api.FeatureConsistencyBackflowResponse, error) {
	return e.APIClient.FeatureConsistencyCheckApi.BackflowFeatureConsistencyCheckJobData(backflowData)
}
func (e *ExperimentClient) SyncFeatureConsistencyCheckJobReplayLog(replyData *model.FeatureConsistencyReplyData) (api.FeatureConsistencyReplyResponse, error) {
	return e.APIClient.FeatureConsistencyCheckApi.SyncFeatureConsistencyCheckJobReplayLog(replyData)
}
