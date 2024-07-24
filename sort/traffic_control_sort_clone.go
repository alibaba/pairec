package sort

import (
	"encoding/json"
	"fmt"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

func (p *TrafficControlSort) CloneWithConfig(params map[string]interface{}) ISort {
	j, err := json.Marshal(params)
	if err != nil {
		log.Error(fmt.Sprintf("event=CloneWithConfig\terror=%v", err))
		return p
	}

	config := recconf.SortConfig{}
	if err := json.Unmarshal(j, &config); err != nil {
		log.Error(fmt.Sprintf("event=CloneWithConfig\terror=%v", err))
		return p
	}

	d, _ := json.Marshal(config)
	md5 := utils.Md5(string(d))
	if sort, ok := p.cloneInstances[md5]; ok {
		return sort
	}

	sort := NewTrafficControlSort(config)
	if sort != nil {
		p.cloneInstances[md5] = sort
	}
	return sort
}

func (p *TrafficControlSort) GetSortName() string {
	return p.name
}
