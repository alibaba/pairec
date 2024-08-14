package pipeline

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/rank"
	"github.com/alibaba/pairec/v2/utils"
)

type ColdStartRankService struct {
	pipelineName      string
	coldStartRankConf *recconf.ColdStartRankConfig
	coldStartRanks    map[string]*rank.ColdStartRank
}

func NewColdStartRankService(config *recconf.PipelineConfig) *ColdStartRankService {
	service := ColdStartRankService{
		pipelineName:      config.Name,
		coldStartRankConf: &config.ColdStartRankConf,
		coldStartRanks:    make(map[string]*rank.ColdStartRank),
	}

	return &service
}

func (s *ColdStartRankService) GetColdStartRankByContext(context *context.RecommendContext) *rank.ColdStartRank {
	scene := context.GetParameter("scene").(string)
	// find rank config
	var rankConfig recconf.ColdStartRankConfig
	found := false
	if context.ExperimentResult != nil {
		rankconf := context.ExperimentResult.GetExperimentParams().Get("pipelines."+s.pipelineName+".ColdStartRankConf", "")
		if rankconf != "" {
			d, _ := json.Marshal(rankconf)
			if err := json.Unmarshal(d, &rankConfig); err == nil {
				found = true
			}
		}
	}

	if !found {
		if s.coldStartRankConf != nil {
			rankConfig = *s.coldStartRankConf
			found = true
		}
	}

	// not find any generalRankconf, return  nil
	if !found {
		return nil
	}

	if rankConfig.AlgoName == "" {
		return nil
	}

	d, _ := json.Marshal(rankConfig)
	id := scene + "#" + utils.Md5(string(d))
	if coldStartRank, ok := s.coldStartRanks[id]; ok {
		return coldStartRank
	}

	coldStartRank := rank.NewColdStartRank(&rankConfig)

	s.coldStartRanks[id] = coldStartRank

	return coldStartRank
}

func (s *ColdStartRankService) Rank(user *module.User, items []*module.Item, context *context.RecommendContext) {
	start := time.Now()
	coldStartRank := s.GetColdStartRankByContext(context)
	if coldStartRank == nil {
		return
	}
	var requestData []map[string]interface{}
	for _, item := range items {
		features := item.GetFeatures()
		requestData = append(requestData, features)

	}

	coldStartRank.Rank(user, items, requestData, context)

	log.Info(fmt.Sprintf("requestId=%s\tmodule=ColdStartRank\tpipeline=%s\tcost=%d", context.RecommendId, s.pipelineName, utils.CostTime(start)))
}
