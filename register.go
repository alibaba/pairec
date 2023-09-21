package pairec

import (
	"github.com/alibaba/pairec/filter"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/service/metrics"
	"github.com/alibaba/pairec/service/recall"
	"github.com/alibaba/pairec/sort"
)

func register(conf *recconf.RecommendConfig) {
	registerRecall(conf)
	registerFilter(conf)
	registerSort(conf)
	registerMetrics(conf)
}

func registerFilter(conf *recconf.RecommendConfig) {
	filter.RegisterFilterWithConfig(conf)
	filter.RegisterFilter("UniqueFilter", filter.NewUniqueFilter())
}

func registerSort(conf *recconf.RecommendConfig) {
	for _, conf_ := range conf.DPPConf {
		sort.RegisterSort(conf_.Name, sort.NewDPPSort(conf_))
	}
	sort.RegisterSortWithConfig(conf)
}

func registerRecall(conf *recconf.RecommendConfig) {
	recall.RegisterRecall("ContextItemRecall", recall.NewContextItemRecall(recconf.RecallConfig{Name: "ContextItemRecall"}))
	recall.Load(conf)
}

func registerMetrics(conf *recconf.RecommendConfig) {
	metrics.Load(conf)
}
