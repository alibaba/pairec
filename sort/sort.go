package sort

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

var sortMapping = make(map[string]ISort)
var sortSigns = make(map[string]string)

var sortService *SortService

// var sortMapping map[string]ISort

func init() {
	sortService = &SortService{}
	sortService.SortStrategies = make(map[string][]ISort, 0)
}

type SortData struct {
	Data         interface{}
	Context      *context.RecommendContext
	User         *module.User
	PipelineName string
}
type ISort interface {
	Sort(sortData *SortData) error
}
type ICloneSort interface {
	CloneWithConfig(params map[string]interface{}) ISort
	GetSortName() string
}

type SortService struct {
	SortStrategies map[string][]ISort
}

func (ss *SortService) AddSort(scene string, s ISort) {
	sorts, ok := ss.SortStrategies[scene]
	if ok {
		for _, sort := range sorts {
			if sort == s {
				return
			}
		}

		sorts = append(sorts, s)
		ss.SortStrategies[scene] = sorts
	} else {
		sorts := []ISort{s}
		ss.SortStrategies[scene] = sorts
	}
}
func (ss *SortService) AddSorts(scene string, sorts []ISort) {
	ss.SortStrategies[scene] = sorts
}

func (ss *SortService) Sort(data *SortData, tag string) {
	var scene string
	ctx := data.Context
	s := ctx.GetParameter("scene")
	if _, ok := s.(string); ok {
		scene = s.(string)
	}
	var categoryName string
	c := ctx.GetParameter("category")
	if _, ok := c.(string); ok {
		categoryName = c.(string)
	} else {
		categoryName = "default"
	}

	sorts := make([]ISort, 0)
	if ctx.ExperimentResult != nil {
		names := ctx.ExperimentResult.GetExperimentParams().Get(categoryName+".SortNames", nil)
		if names != nil {
			if values, ok := names.([]interface{}); ok {
				for _, v := range values {
					if name, okay := v.(string); okay {
						if sort, found := sortMapping[name]; found {
							sorts = append(sorts, sort)
						}
					}
				}
			}
		}
	}

	if len(sorts) == 0 {
		scene = scene + tag
		var ok bool
		sorts, ok = ss.SortStrategies[scene]
		if !ok {
			sorts, ok = ss.SortStrategies[categoryName]
		}
		if !ok || sorts == nil || len(sorts) == 0 {
			sorts = make([]ISort, 1)
			sorts[0] = NewItemRankScoreSort()
			ctx.LogInfo(fmt.Sprintf("defaultSort=ItemRankScore\tscene=%s", scene))
		}
	}

	for _, sort := range sorts {
		newSort := sort
		if cloneSort, ok := sort.(ICloneSort); ok && ctx.ExperimentResult != nil {
			sortConfig := ctx.ExperimentResult.GetExperimentParams().Get("sort."+cloneSort.GetSortName(), nil)
			if sortConfig != nil {
				if params, ok := sortConfig.(map[string]interface{}); ok {
					if sortInstance := cloneSort.CloneWithConfig(params); !utils.IsNil(sortInstance) {
						newSort = sortInstance
					}
				}
			}
		}

		newSort.Sort(data)
	}
}

func Load(config *recconf.RecommendConfig) {
	for scene, names := range config.SortNames {
		var sorts []ISort
		for _, name := range names {
			if sort, ok := sortMapping[name]; ok {
				sorts = append(sorts, sort)
			}
		}
		sortService.AddSorts(scene, sorts)
	}
}

func Sort(sortData *SortData, tag string) {
	sortService.Sort(sortData, tag)
}

func RegisterSort(name string, s ISort) {
	if s == nil {
		panic("Sort is nil, name:" + name)
	}
	if _, ok := sortMapping[name]; !ok {
		sortMapping[name] = s
	}
}

// GetSort return sort by name
func GetSort(name string) (ISort, error) {
	sort, ok := sortMapping[name]
	if !ok {
		return nil, fmt.Errorf("ISort not found, name:%s", name)
	}

	return sort, nil
}

func RegisterSortWithConfig(config *recconf.RecommendConfig) {
	for _, conf := range config.SortConfs {
		if _, ok := sortMapping[conf.Name]; ok {
			sign, _ := json.Marshal(&conf)
			if utils.Md5(string(sign)) == sortSigns[conf.Name] {
				continue
			}
		}

		var s ISort
		if conf.SortType == "DPPSort" {
			s = NewDPPSort(conf.DPPConf)
		} else if conf.SortType == "MultiRecallMixSort" {
			s = NewMultiRecallMixSort(conf)
		} else if conf.SortType == "BoostScoreSort" {
			s = NewBoostScoreSort(conf)
		} else if conf.SortType == "DiversityRuleSort" {
			s = NewDiversityRuleSort(conf)
		} else if conf.SortType == "AlgoScoreSort" {
			s = NewAlgoScoreSort(conf)
		} else if conf.SortType == "TrafficControlSort" {
			s = NewTrafficControlSort(conf)
		} else if conf.SortType == "BoostScoreByWeight" {
			s = NewBoostScoreByWeight(conf)
		} else if conf.SortType == "DistinctIdSort" {
			s = NewDistinctIdSort(conf)
		}

		if s == nil {
			panic("Sort is nil, name:" + conf.Name)
		}

		sign, _ := json.Marshal(&conf)
		registerSortWithSign(conf.Name, s, utils.Md5(string(sign)))
	}
}
func registerSortWithSign(name string, sort ISort, sign string) {
	sortMapping[name] = sort
	sortSigns[name] = sign
}

func sortInfoLog(sortData *SortData, module string, count int, start time.Time) {
	if sortData.PipelineName != "" {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=%s\tpipeline=%s\tcount=%d\tcost=%d", sortData.Context.RecommendId, module, sortData.PipelineName, count, utils.CostTime(start)))
	} else {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=%s\tcount=%d\tcost=%d", sortData.Context.RecommendId, module, count, utils.CostTime(start)))
	}
}

func sortInfoLogWithName(sortData *SortData, module, name string, count int, start time.Time) {
	if sortData.PipelineName != "" {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=%s\tname=%s\tpipeline=%s\tcount=%d\tcost=%d", sortData.Context.RecommendId, module, name, sortData.PipelineName, count, utils.CostTime(start)))
	} else {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=%s\tname=%s\tcount=%d\tcost=%d", sortData.Context.RecommendId, module, name, count, utils.CostTime(start)))
	}
}
