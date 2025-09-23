package pipeline

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/general_rank"
	"github.com/alibaba/pairec/v2/utils"
)

type GeneralRank struct {
	*general_rank.BaseGeneralRank
}

func NewGeneralRankWithConfig(scene string, config recconf.GeneralRankConfig) (*GeneralRank, error) {

	preRank := GeneralRank{
		BaseGeneralRank: general_rank.NewBaseGeneralRankWithConfig(scene, config),
	}

	return &preRank, nil
}

func (r *GeneralRank) Rank(user *module.User, items []*module.Item, context *context.RecommendContext, pipeline string) (ret []*module.Item) {

	ret = r.DoRank(user, items, context, pipeline)
	return
}

type GeneralRankService struct {
	pipelineName    string
	generalRankConf recconf.GeneralRankConfig
	mu              sync.RWMutex
	generalRanks    map[string]*GeneralRank
}

func NewGeneralRankService(config *recconf.PipelineConfig) *GeneralRankService {
	rank := GeneralRankService{
		generalRanks:    make(map[string]*GeneralRank, 0),
		pipelineName:    config.Name,
		generalRankConf: config.GeneralRankConf,
	}

	return &rank
}

func (r *GeneralRankService) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.generalRanks = make(map[string]*GeneralRank, 0)
}

func (r *GeneralRankService) AddGeneralRank(rankId string, rank *GeneralRank) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.generalRanks[rankId] = rank
}

func (r *GeneralRankService) GetGeneralRank(rankId string) (*GeneralRank, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rank, ok := r.generalRanks[rankId]
	return rank, ok
}

// GetGeneralRank return the GeneralRank by the context
// If not found, return nil
func (r *GeneralRankService) GetGeneralRankByContext(context *context.RecommendContext) *GeneralRank {
	scene := context.GetParameter("scene").(string)
	// find rank config
	var rankConfig recconf.GeneralRankConfig
	found := false
	if context.ExperimentResult != nil {
		rankconf := context.ExperimentResult.GetExperimentParams().Get("pipelines."+r.pipelineName+".GeneralRankConf", "")
		if rankconf != "" {
			d, _ := json.Marshal(rankconf)
			if err := json.Unmarshal(d, &rankConfig); err == nil {
				found = true
			}
		}
	}

	if !found {
		rankConfig = r.generalRankConf
		found = true
	}

	// not find any generalRankconf, return  nil
	if !found {
		return nil
	}
	d, _ := json.Marshal(rankConfig)
	id := scene + "#" + utils.Md5(string(d))
	r.mu.RLock()
	generalRank, ok := r.generalRanks[id]
	r.mu.RUnlock()
	if ok {
		return generalRank
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if generalRank, ok := r.generalRanks[id]; ok {
		return generalRank
	}

	generalRank, err := NewGeneralRankWithConfig(scene, rankConfig)
	if err != nil {
		log.Error(fmt.Sprintf("create generalRank error, err:%v", err))
		return nil
	}

	r.generalRanks[id] = generalRank

	return generalRank
}

func (r *GeneralRankService) Rank(user *module.User, items []*module.Item, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()

	generalRank := r.GetGeneralRankByContext(context)
	if generalRank == nil {
		return items
	}

	var (
		rankItems []*module.Item

		generalRankResult []*module.Item
	)

	rankItems = items

	generalRankResult = generalRank.Rank(user, rankItems, context, r.pipelineName)

	ret = append(ret, generalRankResult...)

	log.Info(fmt.Sprintf("requestId=%s\tmodule=GeneralRank\tpipeline=%s\tcount=%d\tcost=%d", context.RecommendId, r.pipelineName, len(ret), utils.CostTime(start)))
	return
}
