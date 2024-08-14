package general_rank

import (
	"fmt"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type GeneralRank struct {
	*BaseGeneralRank
}

func NewGeneralRankWithConfig(scene string, config recconf.GeneralRankConfig) (*GeneralRank, error) {

	preRank := GeneralRank{
		BaseGeneralRank: NewBaseGeneralRankWithConfig(scene, config),
	}

	return &preRank, nil
}

func (r *GeneralRank) Rank(user *module.User, items []*module.Item, context *context.RecommendContext) (ret []*module.Item) {

	ret = r.DoRank(user, items, context, "")
	log.Info(fmt.Sprintf("requestId=%s\tmodule=BaseGeneralRank\tcount=%d", context.RecommendId, len(ret)))
	return
}

// LoadGeneralRankWithConfig load general rank config by the recconf config
// When the function invoked, shows a new config reload, avoid using old filter or sort in the actions, we will clear the rank map,
// because when request come, GetColdStartGeneralRankByContext or GetGeneralRankByContext will get the new config
// and regiter the rank in the map.
func LoadGeneralRankWithConfig(config *recconf.RecommendConfig) {
	DefaultGeneralRankService().Clear()
	/**
	for sceneName, rankConfig := range config.GeneralRankConfs {
		d, _ := json.Marshal(rankConfig)
		id := sceneName + "#" + utils.Md5(string(d))
		_, ok := DefaultGeneralRankService().GetGeneralRank(id)
		if ok {
			continue
		}

		if generalRank, err := NewGeneralRankWithConfig(sceneName, rankConfig); err == nil {
			RegisterGeneralRank(id, generalRank)
		} else {
			log.Error(fmt.Sprintf("load general rank error:%v", err))
		}
	}

	for sceneName, rankConfig := range config.ColdStartGeneralRankConfs {
		d, _ := json.Marshal(rankConfig)
		id := sceneName + "#" + utils.Md5(string(d))
		_, ok := DefaultGeneralRankService().GetColdStartGeneralRank(id)
		if ok {
			continue
		}

		if coldStartGeneralRank, err := NewColdStartGeneralRankWithConfig(sceneName, rankConfig); err == nil {
			RegisterColdStartGeneralRank(id, coldStartGeneralRank)
		} else {
			log.Error(fmt.Sprintf("load coldstart general rank error:%v", err))
		}
	}
	**/
}

var generalRankService *GeneralRankService

func init() {
	generalRankService = NewGeneralRankService()
}

type GeneralRankService struct {
	generalRanks          map[string]*GeneralRank
	coldStartGeneralRanks map[string]*ColdStartGeneralRank
}

func NewGeneralRankService() *GeneralRankService {
	rank := GeneralRankService{
		generalRanks:          make(map[string]*GeneralRank, 0),
		coldStartGeneralRanks: make(map[string]*ColdStartGeneralRank, 0),
	}

	return &rank
}

func DefaultGeneralRankService() *GeneralRankService {
	return generalRankService
}

func RegisterGeneralRank(rankId string, rank *GeneralRank) {
	DefaultGeneralRankService().AddGeneralRank(rankId, rank)
}

func RegisterColdStartGeneralRank(rankId string, rank *ColdStartGeneralRank) {
	DefaultGeneralRankService().AddColdStartGeneralRank(rankId, rank)
}

func (r *GeneralRankService) Clear() {
	r.generalRanks = make(map[string]*GeneralRank, 0)
	r.coldStartGeneralRanks = make(map[string]*ColdStartGeneralRank, 0)
}

func (r *GeneralRankService) AddGeneralRank(rankId string, rank *GeneralRank) {
	r.generalRanks[rankId] = rank
}

func (r *GeneralRankService) AddColdStartGeneralRank(rankId string, rank *ColdStartGeneralRank) {
	r.coldStartGeneralRanks[rankId] = rank
}

func (r *GeneralRankService) GetGeneralRank(rankId string) (*GeneralRank, bool) {
	rank, ok := r.generalRanks[rankId]
	return rank, ok
}

func (r *GeneralRankService) GetColdStartGeneralRank(rankId string) (*ColdStartGeneralRank, bool) {
	rank, ok := r.coldStartGeneralRanks[rankId]
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
		rankconf := context.ExperimentResult.GetExperimentParams().Get("generalRankConf", "")
		if rankconf != "" {
			d, _ := json.Marshal(rankconf)
			if err := json.Unmarshal(d, &rankConfig); err == nil {
				found = true
			}
		}
	}

	if !found {
		if rankConfigs, ok := recconf.Config.GeneralRankConfs[scene]; ok {
			rankConfig = rankConfigs
			found = true
		}
	}

	// not find any generalRankconf, return  nil
	if !found {
		return nil
	}
	d, _ := json.Marshal(rankConfig)
	id := scene + "#" + utils.Md5(string(d))
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

func (r *GeneralRankService) GetColdStartGeneralRankByContext(context *context.RecommendContext) *ColdStartGeneralRank {
	scene := context.GetParameter("scene").(string)
	// find rank config
	var rankConfig recconf.ColdStartGeneralRankConfig
	found := false
	if context.ExperimentResult != nil {
		rankconf := context.ExperimentResult.GetExperimentParams().Get("coldStartGeneralRankConf", "")
		if rankconf != "" {
			d, _ := json.Marshal(rankconf)
			if err := json.Unmarshal(d, &rankConfig); err == nil {
				found = true
			}
		}
	}

	if !found {
		if rankConfigs, ok := recconf.Config.ColdStartGeneralRankConfs[scene]; ok {
			rankConfig = rankConfigs
			found = true
		}
	}

	// not find any generalRankconf, return  nil
	if !found {
		return nil
	}
	d, _ := json.Marshal(rankConfig)
	id := scene + "#" + utils.Md5(string(d))
	if coldStartGeneralRank, ok := r.coldStartGeneralRanks[id]; ok {
		return coldStartGeneralRank
	}

	coldStartGeneralRank, err := NewColdStartGeneralRankWithConfig(scene, rankConfig)
	if err != nil {
		log.Error(fmt.Sprintf("create ColdStartGeneralRank error, err:%v", err))
		return nil
	}

	r.coldStartGeneralRanks[id] = coldStartGeneralRank

	return coldStartGeneralRank
}
func (r *GeneralRankService) Rank(user *module.User, items []*module.Item, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()

	generalRank := r.GetGeneralRankByContext(context)
	if generalRank == nil {
		return items
	}

	var (
		coldStartItems []*module.Item
		rankItems      []*module.Item

		coldStartRankResult []*module.Item
		generalRankResult   []*module.Item
	)

	coldGeneralRank := r.GetColdStartGeneralRankByContext(context)
	if coldGeneralRank == nil {
		rankItems = items
	} else {
		for _, item := range items {
			if coldGeneralRank.Filter(user, item, context) {
				coldStartItems = append(coldStartItems, item)
			} else {
				rankItems = append(rankItems, item)
			}
		}
	}

	var wg sync.WaitGroup
	if coldGeneralRank != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			coldStartRankResult = coldGeneralRank.Rank(user, coldStartItems, context)
		}()
	}

	generalRankResult = generalRank.Rank(user, rankItems, context)

	wg.Wait()

	ret = append(ret, generalRankResult...)
	ret = append(ret, coldStartRankResult...)

	log.Info(fmt.Sprintf("requestId=%s\tmodule=GeneralRank\tcount=%d\tcost=%d", context.RecommendId, len(ret), utils.CostTime(start)))
	return
}
