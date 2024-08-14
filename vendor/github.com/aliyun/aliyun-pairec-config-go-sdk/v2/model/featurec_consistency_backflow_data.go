package model

type FeatureConsistencyBackflowData struct {
	FeatureConsistencyCheckJobConfigId string `json:"FeatureConsistencyCheckJobConfigId,omitempty"`
	InstanceId                         string `json:"InstanceId,omitempty"`
	LogUserId                          string `json:"LogUserId,omitempty"`
	LogItemId                          string `json:"LogItemId,omitempty"`
	LogRequestId                       string `json:"LogRequestId,omitempty"`
	SceneName                          string `json:"SceneName,omitempty"`
	Scores                             string `json:"Scores,omitempty"`
	UserFeatures                       string `json:"UserFeatures,omitempty"`
	ItemFeatures                       string `json:"ItemFeatures,omitempty"`
	LogRequestTime                     int64  `json:"LogRequestTime,omitempty"`
}
