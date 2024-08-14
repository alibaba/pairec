package metrics

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
)

var (
	SizeNotEnoughTotal    *prometheus.CounterVec
	RecTotal              *prometheus.CounterVec
	RecallItemsPercentage *prometheus.GaugeVec
	RecallDurSecs         *prometheus.HistogramVec
	FilterDurSecs         *prometheus.HistogramVec
	GeneralRankDurSecs    *prometheus.HistogramVec
	LoadFeatureDurSecs    *prometheus.HistogramVec
	RankDurSecs           *prometheus.HistogramVec
	SortDurSecs           *prometheus.HistogramVec
	RecDurSecs            *prometheus.HistogramVec

	enabled = false
	once    sync.Once
)

var CustomRegister = prometheus.NewRegistry()

func Enabled() bool {
	return enabled
}

func Load(conf *recconf.RecommendConfig) {
	once.Do(func() {
		initMetrics(conf)

		register(RecTotal,
			SizeNotEnoughTotal,
			RecallItemsPercentage,
			RecDurSecs,
			RecallDurSecs,
			FilterDurSecs,
			GeneralRankDurSecs,
			LoadFeatureDurSecs,
			RankDurSecs, SortDurSecs)
	})

	enabled = conf.PrometheusConfig.Enable
}

func register(cs ...prometheus.Collector) {
	for _, c := range cs {
		if c != nil {
			err := prometheus.Register(c)
			if err != nil {
				log.Error(fmt.Sprintf("Module=PrometheusMetric\terr=%s", err))
			}
		}
	}
}

func initMetrics(conf *recconf.RecommendConfig) {
	subsystem := conf.PrometheusConfig.Subsystem
	if subsystem == "" {
		subsystem = "pairec"
	}
	commonLabels := []string{
		"scene", "exp_id",
	}
	buckets := []float64{
		.005, .01, .02, .025, .03, .04, .05, .06, .07, .08, .09, .1, .15, .2, .25, .3, .35, .4, .45, .5, .55, .6, .7, .8,
	}

	RecTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: subsystem,
		Name:      "rec_total",
		Help:      "How many times of recommend.",
	}, commonLabels)

	SizeNotEnoughTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: subsystem,
		Name:      "size_not_enough_total",
		Help:      "How many times of recommend with not enough items.",
	}, commonLabels)

	RecallItemsPercentage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: subsystem,
		Name:      "recall_items_percentage",
		Help:      "The recall items count percentage.",
	}, []string{
		"recall_name",
	})

	RecDurSecs = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "rec_duration_seconds",
		Buckets:   buckets,
		Help:      "The total recommend cost in seconds.",
	}, commonLabels)

	RecallDurSecs = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "recall_duration_seconds",
		Buckets:   buckets,
		Help:      "The recall cost in seconds.",
	}, commonLabels)

	FilterDurSecs = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "filter_duration_seconds",
		Buckets:   buckets,
		Help:      "The filter cost in seconds.",
	}, commonLabels)

	GeneralRankDurSecs = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "general_rank_duration_seconds",
		Buckets:   buckets,
		Help:      "The general rank cost in seconds.",
	}, commonLabels)

	LoadFeatureDurSecs = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "load_feature_duration_seconds",
		Buckets:   buckets,
		Help:      "The load feature cost in seconds.",
	}, []string{
		"scene", "exp_id", "stage",
	})

	RankDurSecs = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "rank_duration_seconds",
		Buckets:   buckets,
		Help:      "The rank cost in seconds.",
	}, commonLabels)

	SortDurSecs = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "sort_duration_seconds",
		Buckets:   buckets,
		Help:      "The sort cost in seconds.",
	}, commonLabels)
}
