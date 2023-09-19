package sort

import (
	"errors"
	"fmt"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

type BoostScoreCondition struct {
	filterParam         *module.FilterParam
	evaluableExpression *govaluate.EvaluableExpression
}

func NewBoostScoreCondition(config *recconf.BoostScoreCondition) (*BoostScoreCondition, error) {
	condition := &BoostScoreCondition{}
	if config.Expression != "" {
		expression, err := govaluate.NewEvaluableExpression(config.Expression)
		if err != nil {
			return nil, err
		}

		condition.evaluableExpression = expression
	}

	filterParam := module.NewFilterParamWithConfig(config.Conditions)

	condition.filterParam = filterParam

	return condition, nil
}

type BoostScoreSort struct {
	debug          bool
	filterAll      bool
	name           string
	conditions     []*BoostScoreCondition
	cloneInstances map[string]*BoostScoreSort
}

func NewBoostScoreSort(config recconf.SortConfig) *BoostScoreSort {
	sort := BoostScoreSort{
		debug:          config.Debug,
		filterAll:      config.BoostScoreConditionsFilterAll,
		name:           config.Name,
		cloneInstances: make(map[string]*BoostScoreSort),
	}

	for _, boostScoreConditionConfig := range config.BoostScoreConditions {
		condition, err := NewBoostScoreCondition(&boostScoreConditionConfig)
		if err != nil {
			log.Error(fmt.Sprintf("BoostScoreCondition error:%v", err))
		} else {
			sort.conditions = append(sort.conditions, condition)
		}
	}

	return &sort
}
func (s *BoostScoreSort) Sort(sortData *SortData) error {
	if _, ok := sortData.Data.([]*module.Item); !ok {
		return errors.New("sort data type error")
	}

	return s.doSort(sortData)
}

func (s *BoostScoreSort) doSort(sortData *SortData) error {
	start := time.Now()
	items := sortData.Data.([]*module.Item)
	userProperties := sortData.User.MakeUserFeatures2()
	for _, item := range items {
		properties := item.GetProperties()
		for _, condition := range s.conditions {
			if flag, err := condition.filterParam.EvaluateByDomain(userProperties, properties); err == nil && flag {
				properties["score"] = item.Score
				result, err := condition.evaluableExpression.Evaluate(properties)
				if err != nil {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=BoostScoreSort\titemId=%s\terror=%v", sortData.Context.RecommendId, item.Id, err))
				} else {
					if value, ok := result.(float64); ok {
						if s.debug {
							item.AddProperty("org_score", item.Score)
							log.Info(fmt.Sprintf("requestId=%s\tmodule=BoostScoreSort\tname=%s\titemId=%s\torg_score=%f\tscore=%f",
								sortData.Context.RecommendId, s.name, item.Id, item.Score, value))
						}
						item.Score = value
					}
				}
				if !s.filterAll {
					break
				}
			}
		}
	}
	sortData.Data = items
	sortInfoLogWithName(sortData, "BoostScoreSort", s.name, len(items), start)
	return nil
}
