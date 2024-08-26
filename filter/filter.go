package filter

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

var filterMapping = make(map[string]IFilter)
var filterSigns = make(map[string]string)

var filterService *FilterService

func init() {
	filterService = &FilterService{}
	filterService.Filters = make(map[string][]IFilter)
}

type FilterData struct {
	Uid          module.UID
	User         *module.User
	Data         interface{}
	Context      *context.RecommendContext
	PipelineName string
}
type IFilter interface {
	Filter(filterData *FilterData) error
}

type FilterService struct {
	Filters map[string][]IFilter
}

func (fs *FilterService) AddFilter(scene string, filter IFilter) {
	// if exist in Filters, direct return
	filters, ok := fs.Filters[scene]
	if ok {
		for _, f := range filters {
			if f == filter {
				return
			}
		}

		filters = append(filters, filter)
		fs.Filters[scene] = filters
	} else {
		filters := []IFilter{filter}
		fs.Filters[scene] = filters
	}
}
func (fs *FilterService) AddFilters(scene string, filters []IFilter) {
	fs.Filters[scene] = filters
}

func (fs *FilterService) Filter(filterData *FilterData, tag string) {
	context := filterData.Context
	scene := context.GetParameter("scene").(string)

	var filters []IFilter
	found := false
	if context.ExperimentResult != nil {
		names := context.ExperimentResult.GetExperimentParams().Get("filterNames", nil)
		if names != nil {
			found = true
			if values, ok := names.([]interface{}); ok {
				for _, v := range values {
					if filterName, ok := v.(string); ok {
						if filter, exist := filterMapping[filterName]; exist {
							filters = append(filters, filter)
						}
					}
				}
			}
		}
	}

	if found && len(filters) == 0 {
		return
	}

	if len(filters) == 0 {
		if filterList, ok := fs.Filters[scene]; ok {
			filters = filterList
		} else {
			filters = fs.Filters["default"]

		}

		/*
			if len(filters) == 0 {
				log.Error(fmt.Sprintf("Filters:not find, scene:%s", scene))
				return
			}
		*/

	}

	// var err error
	for _, f := range filters {
		f.Filter(filterData)
	}
}

func RegisterFilterWithConfig(config *recconf.RecommendConfig) {
	for _, conf := range config.FilterConfs {
		if _, ok := filterMapping[conf.Name]; ok {
			sign, _ := json.Marshal(&conf)
			if utils.Md5(string(sign)) == filterSigns[conf.Name] {
				continue
			}
		}

		var f IFilter
		if conf.FilterType == "User2ItemExposureFilter" {
			f = NewUser2ItemExposureFilter(conf)
		} else if conf.FilterType == "User2ItemCustomFilter" {
			f = NewUser2ItemCustomFilter(conf)
		} else if conf.FilterType == "AdjustCountFilter" {
			f = NewAdjustCountFilter(conf)
		} else if conf.FilterType == "PriorityAdjustCountFilter" {
			f = NewPriorityAdjustCountFilter(conf)
		} else if conf.FilterType == "ItemStateFilter" {
			f = NewItemStateFilter(conf)
		} else if conf.FilterType == "ItemCustomFilter" {
			f = NewItemCustomFilter(conf)
		} else if conf.FilterType == "CompletelyFairFilter" {
			f = NewCompletelyFairCountFilter(conf)
		} else if conf.FilterType == "GroupWeightCountFilter" {
			f = NewGroupWeightCountFilter(conf)
		} else if conf.FilterType == "DimensionFieldUniqueFilter" {
			f = NewDimensionFieldUniqueFilter(conf)
		} else if conf.FilterType == "User2ItemExposureWithConditionFilter" {
			f = NewUser2ItemExposureWithConditionFilter(conf)
		} else if conf.FilterType == "ConditionFilter" {
			f = NewConditionFilter(conf)
		}

		if f == nil {
			panic("Filter is nil, name:" + conf.Name)
		}

		sign, _ := json.Marshal(&conf)
		registerFilterWithSign(conf.Name, f, utils.Md5(string(sign)))

	}
}

func Load(config *recconf.RecommendConfig) {
	for scene, filterList := range config.FilterNames {
		var filters []IFilter
		for _, name := range filterList {
			if filter, ok := filterMapping[name]; ok {
				filters = append(filters, filter)
			} else {
				log.Error(fmt.Sprintf("Filter:not find, name:%s", name))
			}
		}
		filterService.AddFilters(scene, filters)
	}
}

func Filter(filterData *FilterData, tag string) {
	filterService.Filter(filterData, tag)
}

func RegisterFilter(name string, filter IFilter) {
	if filter == nil {
		panic("Filter is nil, name:" + name)
	}
	if _, ok := filterMapping[name]; !ok {
		filterMapping[name] = filter
	}
}
func registerFilterWithSign(name string, filter IFilter, sign string) {
	filterMapping[name] = filter
	filterSigns[name] = sign
}

// GetFilter get filter by the name
func GetFilter(name string) (IFilter, error) {
	filter, ok := filterMapping[name]
	if !ok {
		return nil, fmt.Errorf("Filter not found, name:%s", name)
	}

	return filter, nil
}

func GetFiltersBySceneName(sceneName string) ([]IFilter, bool) {
	var ret []IFilter
	ret, ok := filterService.Filters[sceneName]
	if ok {
		return ret, true
	}

	ret, ok = filterService.Filters["default"]

	return ret, ok
}

func filterInfoLog(filterData *FilterData, module string, count int, start time.Time) {
	ctx := filterData.Context
	if filterData.PipelineName != "" {
		ctx.LogInfo(fmt.Sprintf("module=%s\tpipeline=%s\tcount=%d\tcost=%d", module, filterData.PipelineName, count, utils.CostTime(start)))
	} else {
		ctx.LogInfo(fmt.Sprintf("module=%s\tcount=%d\tcost=%d", module, count, utils.CostTime(start)))
	}
}
