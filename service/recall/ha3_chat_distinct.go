package recall

import (
	"fmt"

	"github.com/alibaba/pairec/v2/recconf"
)

func normalizeHa3ChatDistinctConfig(conf *recconf.Ha3ChatDistinctConfig) *recconf.Ha3ChatDistinctConfig {
	if conf == nil {
		return nil
	}
	if conf.Default == nil {
		panic("Ha3ChatRecallConf.DistinctConf.Default is required")
	}

	rule := *conf.Default
	if !validHa3FieldName(rule.DistKey) {
		panic(fmt.Sprintf("Ha3ChatRecallConf.DistinctConf.Default.DistKey has invalid field name %q", rule.DistKey))
	}
	if rule.DistCount < 0 {
		panic("Ha3ChatRecallConf.DistinctConf.Default.DistCount must not be negative")
	}
	if rule.DistCount == 0 {
		rule.DistCount = 1
	}
	if rule.DistTimes < 0 {
		panic("Ha3ChatRecallConf.DistinctConf.Default.DistTimes must not be negative")
	}
	if rule.DistTimes == 0 {
		rule.DistTimes = 1
	}
	if rule.MaxItemCount < 0 {
		panic("Ha3ChatRecallConf.DistinctConf.Default.MaxItemCount must not be negative")
	}
	reserved := true
	if rule.Reserved != nil {
		reserved = *rule.Reserved
	}
	rule.Reserved = &reserved

	return &recconf.Ha3ChatDistinctConfig{Default: &rule}
}

func buildHa3ChatDistinctClause(conf *recconf.Ha3ChatDistinctConfig) map[string]interface{} {
	rule := conf.Default
	defaultRule := map[string]interface{}{
		"dist_key":   rule.DistKey,
		"dist_count": rule.DistCount,
		"dist_times": rule.DistTimes,
		"reserved":   *rule.Reserved,
	}
	if rule.MaxItemCount > 0 {
		defaultRule["max_item_count"] = rule.MaxItemCount
	}
	return map[string]interface{}{"default": defaultRule}
}
