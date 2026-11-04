package filter

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/goburrow/cache"
)

var filterMapping = make(map[string]IFilter)
var filterSigns = make(map[string]string)
var filterMappingMu sync.RWMutex
var filterLoadMu sync.Mutex

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

type RecallEngineFilter interface {
	GetExposureItemIds(user *module.User, context *context.RecommendContext) string
}

type ICloneFilter interface {
	CloneWithConfig(params map[string]interface{}) IFilter
	GetFilterName() string
}

type FilterService struct {
	mu      sync.RWMutex
	Filters map[string][]IFilter
}

func (fs *FilterService) AddFilter(scene string, filter IFilter) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
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
	fs.mu.Lock()
	defer fs.mu.Unlock()
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
						if filter, err := GetFilter(filterName); err == nil {
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
		fs.mu.RLock()
		if filterList, ok := fs.Filters[scene]; ok {
			filters = filterList
		} else {
			filters = fs.Filters["default"]

		}
		fs.mu.RUnlock()

		/*
			if len(filters) == 0 {
				log.Error(fmt.Sprintf("Filters:not find, scene:%s", scene))
				return
			}
		*/

	}

	for _, f := range filters {
		newFilter := f
		if cloneFilter, ok := f.(ICloneFilter); ok && context.ExperimentResult != nil {
			filterConfig := context.ExperimentResult.GetExperimentParams().Get("filter."+cloneFilter.GetFilterName(), nil)
			if filterConfig != nil {
				if params, ok := filterConfig.(map[string]interface{}); ok {
					if filterInstance := cloneFilter.CloneWithConfig(params); !utils.IsNil(filterInstance) {
						newFilter = filterInstance
					}
				}
			}
		}

		newFilter.Filter(filterData)
	}
}

func RegisterFilterWithConfig(config *recconf.RecommendConfig) {
	filterLoadMu.Lock()
	defer filterLoadMu.Unlock()
	for _, conf := range config.FilterConfs {
		sign, _ := json.Marshal(&conf)
		filterMappingMu.RLock()
		old, ok := filterMapping[conf.Name]
		unchanged := ok && utils.Md5(string(sign)) == filterSigns[conf.Name]
		filterMappingMu.RUnlock()
		if unchanged {
			continue
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
		} else if conf.FilterType == "PriorityAdjustCountFilterV2" {
			f = NewPriorityAdjustCountFilterV2(conf)
		} else if conf.FilterType == "ItemStateFilter" {
			inherited, cacheSign := itemStateCacheConfig(conf, old)
			f = newItemStateFilterWithCache(conf, inherited, cacheSign)
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
		} else if conf.FilterType == "DiversityAdjustCountFilter" {
			f = NewDiversityAdjustCountFilter(conf)
		} else if conf.FilterType == "SnakeFilter" {
			f = NewSnakeFilter(conf)
		} else if conf.FilterType == "RRFusionFilter" {
			f = NewRRFusionFilter(conf)
		}

		if f == nil {
			panic("Filter is nil, name:" + conf.Name)
		}

		registerFilterWithSign(conf.Name, f, utils.Md5(string(sign)))
		if previous, ok := old.(*ItemStateFilter); ok && previous != nil && previous.cacheSign != "" {
			var freshCache cache.Cache
			if current, ok := f.(*ItemStateFilter); ok {
				freshCache = current.itemStateCache()
			}
			invalidateItemStateCache(previous.itemStateCache(), freshCache)
		}

	}
}

func Load(config *recconf.RecommendConfig) {
	for scene, filterList := range config.FilterNames {
		var filters []IFilter
		for _, name := range filterList {
			if filter, err := GetFilter(name); err == nil {
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
	filterMappingMu.Lock()
	defer filterMappingMu.Unlock()
	if filter == nil {
		panic("Filter is nil, name:" + name)
	}
	if _, ok := filterMapping[name]; !ok {
		filterMapping[name] = filter
	}
}
func registerFilterWithSign(name string, filter IFilter, sign string) {
	filterMappingMu.Lock()
	defer filterMappingMu.Unlock()
	filterMapping[name] = filter
	filterSigns[name] = sign
}

// GetFilter get filter by the name
func GetFilter(name string) (IFilter, error) {
	filterMappingMu.RLock()
	defer filterMappingMu.RUnlock()
	filter, ok := filterMapping[name]
	if !ok {
		return nil, fmt.Errorf("Filter not found, name:%s", name)
	}

	return filter, nil
}

func GetFiltersBySceneName(sceneName string) ([]IFilter, bool) {
	filterService.mu.RLock()
	defer filterService.mu.RUnlock()
	var ret []IFilter
	ret, ok := filterService.Filters[sceneName]
	if ok {
		return ret, true
	}

	ret, ok = filterService.Filters["default"]

	return ret, ok
}

func filterInfoLog(filterData *FilterData, module string, filterName string, count int, start time.Time) {
	ctx := filterData.Context
	if filterData.PipelineName != "" {
		ctx.LogInfo(fmt.Sprintf("module=%s\tname=%s\tpipeline=%s\tcount=%d\tcost=%d", module, filterName, filterData.PipelineName, count, utils.CostTime(start)))
	} else {
		ctx.LogInfo(fmt.Sprintf("module=%s\tname=%s\tcount=%d\tcost=%d", module, filterName, count, utils.CostTime(start)))
	}
}
