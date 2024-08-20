package pairec

import (
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"os"

	"github.com/alibaba/pairec/v2/datasource/graph"
	"github.com/alibaba/pairec/v2/datasource/hbase_thrift"
	"github.com/alibaba/pairec/v2/datasource/kafka"
	"github.com/alibaba/pairec/v2/datasource/opensearch"

	"github.com/alibaba/pairec/v2/abtest"
	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/config"
	"github.com/alibaba/pairec/v2/datasource/beengine"
	"github.com/alibaba/pairec/v2/datasource/datahub"
	"github.com/alibaba/pairec/v2/datasource/ha3engine"
	"github.com/alibaba/pairec/v2/datasource/hbase"
	"github.com/alibaba/pairec/v2/datasource/sls"
	"github.com/alibaba/pairec/v2/filter"
	"github.com/alibaba/pairec/v2/persist/clickhouse"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/persist/holo"
	"github.com/alibaba/pairec/v2/persist/lindorm"
	"github.com/alibaba/pairec/v2/persist/mysqldb"
	"github.com/alibaba/pairec/v2/persist/redisdb"
	"github.com/alibaba/pairec/v2/persist/tablestoredb"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service"
	"github.com/alibaba/pairec/v2/service/feature"
	"github.com/alibaba/pairec/v2/service/general_rank"
	"github.com/alibaba/pairec/v2/service/metrics"
	"github.com/alibaba/pairec/v2/service/pipeline"
	"github.com/alibaba/pairec/v2/service/rank"
	"github.com/alibaba/pairec/v2/service/recall/berecall"
	"github.com/alibaba/pairec/v2/sort"
	"github.com/alibaba/pairec/v2/web"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "go.uber.org/automaxprocs"
)

type hookfunc func() error

var (
	hooks = make([]hookfunc, 0)
)
var configFile string

func init() {
	flag.StringVar(&configFile, "config", "", "config file path")
	flag.BoolVar(&config.AppConfig.WarmUpData, "warm-up-data", false, "create eas warm up data flag")
}
func AddStartHook(hf ...hookfunc) {
	hooks = append(hooks, hf...)
}
func Run() {
	mode := os.Getenv("RUN_MODE")
	if mode != "COMMAND" {
		flag.Parse()
	}

	configName := os.Getenv("CONFIG_NAME")
	// if CONFIG_NAME is set, so load the pairec config from abtest server
	// first create the abtest client connect to the server use the env params
	if configName != "" {
		abtest.LoadFromEnvironment()
		ListenConfig(configName)
	} else {
		// load config from local file
		err := recconf.LoadConfig(configFile)
		if err != nil {
			panic(err)
		}
	}

	registerRouteInfo()

	runStartHook()

	PairecApp.Run()
}

func runBeforeStart() {
	mysqldb.Load(recconf.Config)
	redisdb.Load(recconf.Config)
	tablestoredb.Load(recconf.Config)
	sls.Load(recconf.Config)
	kafka.Load(recconf.Config)
	datahub.Load(recconf.Config)
	beengine.Load(recconf.Config)
	graph.Load(recconf.Config)
	ha3engine.Load(recconf.Config)
	opensearch.Load(recconf.Config)
	hbase.Load(recconf.Config)
	hbase_thrift.Load(recconf.Config)
	holo.Load(recconf.Config)
	lindorm.Load(recconf.Config)
	fs.Load(recconf.Config)
	clickhouse.Load(recconf.Config)
	//abtest.Load(recconf.Config)
	algorithm.Load(recconf.Config) // holo must be loaded before loading some algorithm
	register(recconf.Config)
}
func runStartHook() {
	runBeforeStart()

	// first register hook
	AddStartHook(func() error {
		filter.Load(recconf.Config)
		berecall.RegisterFilterWithConfig(recconf.Config)
		sort.Load(recconf.Config)
		// gdb.Load(recconf.Config)
		service.Load(recconf.Config)
		feature.UserLoadFeatureConfig(recconf.Config)
		feature.LoadFeatureConfig(recconf.Config)
		general_rank.LoadGeneralRankWithConfig(recconf.Config)
		rank.LoadColdStartRankConfig(recconf.Config)
		pipeline.LoadPipelineConfigs(recconf.Config)
		// clean log dir
		ClearDir(recconf.Config.LogConf)

		return nil
	})

	for _, hf := range hooks {
		if err := hf(); err != nil {
			panic(err)
		}
	}
}
func registerRouteInfo() {
	// use for listen http server state
	HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "success")
	})
	HandleFunc("/route_paths", func(w http.ResponseWriter, r *http.Request) {
		paths := PairecApp.Handlers.GetRoutePath()

		var result []string
		for _, p := range paths {
			if p == "/ping" ||
				p == "/route_paths" ||
				p == "/api/recommend" ||
				p == "/api/recall" ||
				p == "/api/feature_reply" ||
				p == "/metrics" ||
				p == "/custom_metrics" {
				continue
			}

			result = append(result, p)
		}

		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		d, _ := json.Marshal(result)
		io.WriteString(w, string(d))
	})

	// register recommend Controller
	Route("/api/recommend", &web.RecommendController{})
	Route("/api/recall", &web.UserRecallController{})
	Route("/api/callback", &web.CallBackController{})
	Route("/api/feature_reply", &web.FeatureReplyController{})
	Route("/api/embedding", &web.EmbeddingController{})
	HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})
	HandleFunc("/custom_metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.HandlerFor(metrics.CustomRegister, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	})
}
