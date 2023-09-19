package algorithm

import (
	"github.com/alibaba/pairec/v2/algorithm/response"
	"github.com/alibaba/pairec/v2/recconf"
)

type LookupPolicy struct {
	conf *recconf.LookupConfig
}

type LookupResponse struct {
	score float64
}

func (r *LookupResponse) GetScore() float64 {
	return r.score
}

func (r *LookupResponse) GetScoreMap() map[string]float64 {
	return nil
}

func (r *LookupResponse) GetModuleType() bool {
	return false
}

func NewLookupPolicy() *LookupPolicy {
	return &LookupPolicy{}
}

func (m *LookupPolicy) Init(conf *recconf.AlgoConfig) error {
	m.conf = &conf.LookupConf
	return nil
}

func (m *LookupPolicy) Run(algoData interface{}) (interface{}, error) {
	featureList := algoData.([]map[string]interface{})
	if (len(featureList)) == 0 {
		return nil, nil
	}
	result := make([]response.AlgoResponse, len(featureList))
	for i, f := range featureList {
		if score, ok := f[m.conf.FieldName]; ok {
			result[i] = &LookupResponse{score: score.(float64)}
		} else {
			result[i] = &LookupResponse{score: 0.5}
		}
	}
	return result, nil
}
