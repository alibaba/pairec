package pipeline

import (
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/sort"
)

type SortService struct {
	pipelineName string
	sortNames    []string
}

func NewSortService(config *recconf.PipelineConfig) *SortService {
	service := SortService{
		pipelineName: config.Name,
		sortNames:    config.SortNames,
	}

	return &service
}

func (ss *SortService) Sort(sortData *sort.SortData) {
	context := sortData.Context

	var sorts []sort.ISort
	var sortNames []string
	if context.ExperimentResult != nil {
		names := context.ExperimentResult.GetExperimentParams().Get("pipelines."+ss.pipelineName+".SortNames", nil)
		if names != nil {
			if values, ok := names.([]interface{}); ok {
				for _, v := range values {
					if sortName, ok := v.(string); ok {
						sortNames = append(sortNames, sortName)
					}
				}
			}
		}
	}

	if len(sortNames) == 0 {
		sortNames = ss.sortNames
	}

	for _, sortName := range sortNames {
		if sort, err := sort.GetSort(sortName); err == nil {
			sorts = append(sorts, sort)
		}
	}

	// var err error
	for _, s := range sorts {
		s.Sort(sortData)

	}
}
