package model

type FeatureConsistencyReplyData struct {
	FeatureConsistencyCheckJobConfigId string `json:"FeatureConsistencyCheckJobConfigId,omitempty"`
	InstanceId                         string `json:"InstanceId,omitempty"`
	LogUserId                          string `json:"LogUserId,omitempty"`
	LogItemId                          string `json:"LogItemId,omitempty"`
	LogRequestId                       string `json:"LogRequestId,omitempty"`
	SceneName                          string `json:"SceneName,omitempty"`
	GeneratedFeatures                  string `json:"GeneratedFeatures,omitempty"`
	ContextFeatures                    string `json:"ContextFeatures,omitempty"`
	RawFeatures                        string `json:"RawFeatures,omitempty"`
	LogRequestTime                     int64  `json:"LogRequestTime,omitempty"`
}
