package filter

import (
	"encoding/json"
	"fmt"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

func (f *SnakeFilter) CloneWithConfig(params map[string]interface{}) IFilter {
	j, err := json.Marshal(params)
	if err != nil {
		log.Error(fmt.Sprintf("event=CloneWithConfig\terror=%v", err))
		return f
	}

	config := recconf.FilterConfig{}
	if err := json.Unmarshal(j, &config); err != nil {
		log.Error(fmt.Sprintf("event=CloneWithConfig\terror=%v", err))
		return f
	}

	d, _ := json.Marshal(config)
	md5 := utils.Md5(string(d))
	if cloneFilter, ok := f.cloneInstances.Load(md5); ok {
		return cloneFilter.(*SnakeFilter)
	}

	config.Name = f.name
	cloneFilter := NewSnakeFilter(config)
	if cloneFilter != nil {
		f.cloneInstances.Store(md5, cloneFilter)
	}
	return cloneFilter
}

func (f *SnakeFilter) GetFilterName() string {
	return f.name
}
