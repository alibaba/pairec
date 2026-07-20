package sort

import (
	"encoding/json"
	"fmt"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

func (s *DiversityRuleSortV2) CloneWithConfig(params map[string]interface{}) ISort {
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
	s.cloneMutex.RLock()
	if sort, ok := s.cloneInstances[md5]; ok {
		s.cloneMutex.RUnlock()
		return sort
	}
	s.cloneMutex.RUnlock()

	sort := NewDiversityRuleSortV2(config)
	if sort != nil {
		s.cloneMutex.Lock()
		if existing, ok := s.cloneInstances[md5]; ok {
			s.cloneMutex.Unlock()
			return existing
		}
		s.cloneInstances[md5] = sort
		s.cloneMutex.Unlock()
	}
	return sort
}

func (s *DiversityRuleSortV2) GetSortName() string {
	return s.name
}
