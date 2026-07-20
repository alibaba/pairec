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

	// create under the write lock so that concurrent misses on the same md5
	// never build duplicated instances (the dao cache holds goroutine resource)
	s.cloneMutex.Lock()
	defer s.cloneMutex.Unlock()
	if sort, ok := s.cloneInstances[md5]; ok {
		return sort
	}
	sort := NewDiversityRuleSortV2(config)
	s.cloneInstances[md5] = sort
	return sort
}

func (s *DiversityRuleSortV2) GetSortName() string {
	return s.name
}
