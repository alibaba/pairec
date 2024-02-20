package pairec

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/datasource/graph"
	"github.com/alibaba/pairec/v2/datasource/hbase_thrift"
	"github.com/alibaba/pairec/v2/datasource/opensearch"

	"github.com/alibaba/pairec/v2/abtest"
	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/config"
	"github.com/alibaba/pairec/v2/config/pairec_config"
	"github.com/alibaba/pairec/v2/datasource"
	"github.com/alibaba/pairec/v2/datasource/beengine"
	"github.com/alibaba/pairec/v2/datasource/datahub"
	"github.com/alibaba/pairec/v2/datasource/ha3engine"
	"github.com/alibaba/pairec/v2/datasource/hbase"
	"github.com/alibaba/pairec/v2/datasource/sls"
	"github.com/alibaba/pairec/v2/filter"
	"github.com/alibaba/pairec/v2/log"
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
	"github.com/alibaba/pairec/v2/service/pipeline"
	"github.com/alibaba/pairec/v2/service/recall/berecall"
	"github.com/alibaba/pairec/v2/sort"
)

var (
	loader                  *ConfigLoader
	pairecConfigAdapterName = "pairec_config"
)

type ConfigLoader struct {
	configName         string
	configVersion      string
	configVersionValue string
}

func NewConfigLoader(configName string) *ConfigLoader {
	l := &ConfigLoader{configName: configName, configVersion: configName + "_version"}

	return l
}

func (l *ConfigLoader) loopLoadConfig() {
	for {
		time.Sleep(time.Second * 10)
		version := l.loadConfigVersion()
		// version changed
		if version != "" && version != l.configVersionValue {
			config, err := l.loadConfigFromConfigServer()
			if err != nil {
				fmt.Println(err)
				continue
			}

			l.configVersionValue = version
			log.Info(fmt.Sprintf("config version changed, version:%s, reload config", l.configVersionValue))
			func() {
				defer func() {
					if err := recover(); err != nil {
						log.Error(fmt.Sprintf("reload config error, error:%v", err))
					}
				}()

				l.reloadConfig(config)
			}()
			recconf.UpdateConf(config)
		}
	}
}

func (l *ConfigLoader) reloadConfig(config *recconf.RecommendConfig) {

	mysqldb.Load(config)
	redisdb.Load(config)
	tablestoredb.Load(config)
	sls.Load(config)
	datasource.Load(config)
	datahub.Load(config)
	beengine.Load(config)
	graph.Load(config)
	ha3engine.Load(recconf.Config)
	opensearch.Load(recconf.Config)
	hbase.Load(config)
	holo.Load(config)
	lindorm.Load(config)
	hbase_thrift.Load(config)
	fs.Load(config)
	clickhouse.Load(config)
	algorithm.Load(config) // holo must be loaded before loading some algorithm
	register(config)

	filter.RegisterFilterWithConfig(config)
	filter.Load(config)
	berecall.RegisterFilterWithConfig(config)
	sort.Load(config)
	service.Load(config)
	feature.UserLoadFeatureConfig(config)
	feature.LoadFeatureConfig(config)
	general_rank.LoadGeneralRankWithConfig(config)
	pipeline.LoadPipelineConfigs(config)
}

func (l *ConfigLoader) loadConfigFromConfigServer() (*recconf.RecommendConfig, error) {
	configer, err := config.NewConfig(pairecConfigAdapterName, l.configName)
	if err != nil {
		return nil, err
	}
	rawdata := configer.RawData()
	accessId := os.Getenv("AccessKey")
	accessSecret := os.Getenv("AccessSecret")
	data := string(rawdata)
	data = strings.ReplaceAll(data, "${AccessKey}", accessId)
	data = strings.ReplaceAll(data, "${AccessSecret}", accessSecret)

	configD := &recconf.RecommendConfig{}
	err = json.Unmarshal([]byte(data), configD)
	if err != nil {
		return nil, err
	}

	runMode := os.Getenv("PAIREC_ENVIRONMENT")
	if runMode == "" {
		configD.RunMode = "product"
	} else {
		configD.RunMode = runMode
	}

	return configD, nil
}

func (l *ConfigLoader) loadConfigVersion() string {
	return abtest.GetParams(pairec_config.Pairec_Config_Scene_Name).GetString(l.configVersion, "")
}

// ListenConfig init a instace of ConfigLoader
// ConfigLoader will loop load paire config from server when the config version change
func ListenConfig(configName string) {
	loader = NewConfigLoader(configName)
	config, err := loader.loadConfigFromConfigServer()
	if err != nil {
		panic(err)
	}

	recconf.UpdateConf(config)

	version := loader.loadConfigVersion()

	loader.configVersionValue = version

	go loader.loopLoadConfig()
}
