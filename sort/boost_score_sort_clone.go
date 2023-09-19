package sort

import (
	"encoding/json"
	"fmt"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

func (s *BoostScoreSort) CloneWithConfig(params map[string]interface{}) ISort {
	j, err := json.Marshal(params)
	if err != nil {
		log.Error(fmt.Sprintf("event=CloneWithConfig\terror=%v", err))
		return s
	}

	config := recconf.SortConfig{}
	if err := json.Unmarshal(j, &config); err != nil {
		log.Error(fmt.Sprintf("event=CloneWithConfig\terror=%v", err))
		return s
	}

	d, _ := json.Marshal(config)
	md5 := utils.Md5(string(d))
	if sort, ok := s.cloneInstances[md5]; ok {
		return sort
	}

	config.Name = s.name
	sort := NewBoostScoreSort(config)
	if sort != nil {
		s.cloneInstances[md5] = sort
	}
	return sort
}

func (s *BoostScoreSort) GetSortName() string {
	return s.name
}
