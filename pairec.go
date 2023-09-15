package pairec

import (
	"flag"
	"io"
	"net/http"
	"os"

	"github.com/alibaba/pairec/datasource/graph"
	"github.com/alibaba/pairec/datasource/hbase_thrift"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/alibaba/pairec/abtest"
	"github.com/alibaba/pairec/algorithm"
	"github.com/alibaba/pairec/config"
	"github.com/alibaba/pairec/datasource"
	"github.com/alibaba/pairec/datasource/beengine"
	"github.com/alibaba/pairec/datasource/datahub"
	"github.com/alibaba/pairec/datasource/ha3engine"
	"github.com/alibaba/pairec/datasource/hbase"
	"github.com/alibaba/pairec/datasource/sls"
	"github.com/alibaba/pairec/filter"
	"github.com/alibaba/pairec/persist/clickhouse"
	"github.com/alibaba/pairec/persist/fs"
	"github.com/alibaba/pairec/persist/holo"
	"github.com/alibaba/pairec/persist/lindorm"
	"github.com/alibaba/pairec/persist/mysqldb"
	"github.com/alibaba/pairec/persist/redisdb"
	"github.com/alibaba/pairec/persist/tablestoredb"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/service"
	"github.com/alibaba/pairec/service/feature"
	"github.com/alibaba/pairec/service/general_rank"
	"github.com/alibaba/pairec/service/metrics"
	"github.com/alibaba/pairec/service/pipeline"
	"github.com/alibaba/pairec/service/rank"
	"github.com/alibaba/pairec/service/recall/berecall"
	"github.com/alibaba/pairec/sort"
	"github.com/alibaba/pairec/web"
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
	datasource.Load(recconf.Config)
	datahub.Load(recconf.Config)
	beengine.Load(recconf.Config)
	graph.Load(recconf.Config)
	ha3engine.Load(recconf.Config)
	hbase.Load(recconf.Config)
	hbase_thrift.Load(recconf.Config)
	holo.Load(recconf.Config)
	lindorm.Load(recconf.Config)
	fs.Load(recconf.Config)
	clickhouse.Load(recconf.Config)
	abtest.Load(recconf.Config)
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

	// register recommend Controller
	Route("/api/recommend", &web.RecommendController{})
	Route("/api/recall", &web.UserRecallController{})
	Route("/api/callback", &web.CallBackController{})
	Route("/api/feature_reply", &web.FeatureReplyController{})
	HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})
	HandleFunc("/custom_metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.HandlerFor(metrics.CustomRegister, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	})
}
