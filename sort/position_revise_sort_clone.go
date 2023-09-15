package sort

import (
	"encoding/json"
	"fmt"

	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/utils"
)

func (p *PositionReviseSort) CloneWithConfig(params map[string]interface{}) ISort {
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

	sort := NewPositionReviseSort(config)
	if sort != nil {
		p.cloneInstances[md5] = sort
	}
	return sort
}

func (p *PositionReviseSort) GetSortName() string {
	return p.name
}
