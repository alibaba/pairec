package pipeline

import (
	"github.com/alibaba/pairec/v2/filter"
	"github.com/alibaba/pairec/v2/recconf"
)

type FilterService struct {
	pipelineName string
	filterNames  []string
}

func NewFilterService(config *recconf.PipelineConfig) *FilterService {
	service := FilterService{
		pipelineName: config.Name,
		filterNames:  config.FilterNames,
	}

	return &service
}

func (fs *FilterService) Filter(filterData *filter.FilterData) {
	context := filterData.Context

	var filters []filter.IFilter
	var filterNames []string
	if context.ExperimentResult != nil {
		names := context.ExperimentResult.GetExperimentParams().Get("pipelines."+fs.pipelineName+".FilterNames", nil)
		if names != nil {
			if values, ok := names.([]interface{}); ok {
				for _, v := range values {
					if filterName, ok := v.(string); ok {
						filterNames = append(filterNames, filterName)
					}
				}
			}
		}
	}

	if len(filterNames) == 0 {
		filterNames = fs.filterNames
	}

	for _, filterName := range filterNames {
		if filter, err := filter.GetFilter(filterName); err == nil {
			filters = append(filters, filter)
		}
	}

	// var err error
	for _, f := range filters {
		f.Filter(filterData)
	}
}
