package berecall

import (
	"encoding/json"
	"fmt"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

var filterMapping = make(map[string]IBeFilter)
var filterSigns = make(map[string]string)

var filterService *FilterService

func init() {
	filterService = &FilterService{}
	filterService.Filters = make(map[string][]IBeFilter)
}

type IBeFilter interface {
	BuildQueryParams(user *module.User, context *context.RecommendContext) map[string]string
}

type FilterService struct {
	Filters map[string][]IBeFilter
}

func (fs *FilterService) AddFilter(scene string, filter IBeFilter) {
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
		filters := []IBeFilter{filter}
		fs.Filters[scene] = filters
	}
}
func (fs *FilterService) AddFilters(scene string, filters []IBeFilter) {
	fs.Filters[scene] = filters
}

func RegisterFilterWithConfig(config *recconf.RecommendConfig) {
	for _, conf := range config.BeFilterConfs {
		if _, ok := filterMapping[conf.Name]; ok {
			sign, _ := json.Marshal(&conf)
			if utils.Md5(string(sign)) == filterSigns[conf.Name] {
				continue
			}
		}

		var f IBeFilter
		if conf.FilterType == "User2ItemExposureFilter" {
			f = NewUser2ItemExposureFilter(conf)
		}

		if f == nil {
			panic("Filter is nil, name:" + conf.Name)
		}

		sign, _ := json.Marshal(&conf)
		registerFilterWithSign(conf.Name, f, utils.Md5(string(sign)))

	}
}

func RegisterFilter(name string, filter IBeFilter) {
	if filter == nil {
		panic("Filter is nil, name:" + name)
	}
	if _, ok := filterMapping[name]; !ok {
		filterMapping[name] = filter
	}
}
func registerFilterWithSign(name string, filter IBeFilter, sign string) {
	filterMapping[name] = filter
	filterSigns[name] = sign
}

// GetFilter get filter by the name
func GetFilter(name string) (IBeFilter, error) {
	filter, ok := filterMapping[name]
	if !ok {
		return nil, fmt.Errorf("befilter not found, name:%s", name)
	}

	return filter, nil
}
