package recconf

var modJsonPath = map[string]string{
	HologresConfig{}.ModuleType():     "HologresConfs",
	TableStoreConfig{}.ModuleType():   "TableStoreConfs",
	RedisConfig{}.ModuleType():        "RedisConfs",
	MysqlConfig{}.ModuleType():        "MysqlConfs",
	HBaseConfig{}.ModuleType():        "HBaseConfs",
	FeatureStoreConfig{}.ModuleType(): "FeatureStoreConfs",
	BEConfig{}.ModuleType():           "BEConfs",
	ClickHouseConfig{}.ModuleType():   "ClickHouseConfs",
	LindormConfig{}.ModuleType():      "LindormConfs",
	GraphConfig{}.ModuleType():        "GraphConfs",
	HBaseThriftConfig{}.ModuleType():  "HBaseThriftConfs",
	OpenSearchConfig{}.ModuleType():   "OpenSearchConfs",
	RecallConfig{}.ModuleType():       "RecallConfs",
	FilterConfig{}.ModuleType():       "FilterConfs",
	AlgoConfig{}.ModuleType():         "AlgoConfs",
	SortConfig{}.ModuleType():         "SortConfs",
	SceneRecallConfig{}.ModuleType():  "SceneConfs",
	SceneFilterConfig{}.ModuleType():  "FilterNames",
	GeneralRankConfig{}.ModuleType():  "GeneralRankConfs",
	SceneFeatureConfig{}.ModuleType(): "FeatureConfs",
	RankConfig{}.ModuleType():         "RankConf",
	SceneSortConfig{}.ModuleType():    "SortNames",
}

type ModuleIndex struct {
	Type string
	Name string
}

func (conf RecommendConfig) GetModules() map[ModuleIndex]any {
	modules := make(map[ModuleIndex]any)

	for name, config := range conf.HologresConfs {
		modules[ModuleIndex{Type: config.ModuleType(), Name: name}] = config
	}

	for name, config := range conf.TableStoreConfs {
		modules[ModuleIndex{Type: config.ModuleType(), Name: name}] = config
	}

	for name, config := range conf.RedisConfs {
		modules[ModuleIndex{Type: config.ModuleType(), Name: name}] = config
	}

	for name, config := range conf.MysqlConfs {
		modules[ModuleIndex{Type: config.ModuleType(), Name: name}] = config
	}

	for name, config := range conf.HBaseConfs {
		modules[ModuleIndex{Type: config.ModuleType(), Name: name}] = config
	}

	for name, config := range conf.FeatureStoreConfs {
		modules[ModuleIndex{Type: config.ModuleType(), Name: name}] = config
	}

	for name, config := range conf.BEConfs {
		modules[ModuleIndex{Type: config.ModuleType(), Name: name}] = config
	}

	for name, config := range conf.ClickHouseConfs {
		modules[ModuleIndex{Type: config.ModuleType(), Name: name}] = config
	}

	for name, config := range conf.LindormConfs {
		modules[ModuleIndex{Type: config.ModuleType(), Name: name}] = config
	}

	for name, config := range conf.GraphConfs {
		modules[ModuleIndex{Type: config.ModuleType(), Name: name}] = config
	}

	for name, config := range conf.HBaseThriftConfs {
		modules[ModuleIndex{Type: config.ModuleType(), Name: name}] = config
	}

	for _, config := range conf.RecallConfs {
		modules[ModuleIndex{Type: config.ModuleType(), Name: config.Name}] = config
	}

	for _, config := range conf.FilterConfs {
		modules[ModuleIndex{Type: config.ModuleType(), Name: config.Name}] = config
	}

	for _, config := range conf.AlgoConfs {
		modules[ModuleIndex{Type: config.ModuleType(), Name: config.Name}] = config
	}

	for _, config := range conf.SortConfs {
		modules[ModuleIndex{Type: config.ModuleType(), Name: config.Name}] = config
	}

	for name, config := range conf.SceneConfs {
		modules[ModuleIndex{
			Type: SceneRecallConfig(config).ModuleType(),
			Name: name,
		}] = SceneRecallConfig(config)
	}

	for name, config := range conf.FilterNames {
		modules[ModuleIndex{
			Type: SceneFilterConfig(config).ModuleType(),
			Name: name,
		}] = SceneFilterConfig(config)
	}

	for name, config := range conf.GeneralRankConfs {
		modules[ModuleIndex{Type: config.ModuleType(), Name: name}] = config
	}

	for name, config := range conf.FeatureConfs {
		modules[ModuleIndex{Type: config.ModuleType(), Name: name}] = config
	}

	for name, config := range conf.RankConf {
		modules[ModuleIndex{Type: config.ModuleType(), Name: name}] = config
	}

	for name, config := range conf.SortNames {
		modules[ModuleIndex{
			Type: SceneSortConfig(config).ModuleType(),
			Name: name,
		}] = SceneSortConfig(config)
	}

	return modules
}

func (conf HologresConfig) ModuleType() string {
	return "HologresConf"
}

func (conf TableStoreConfig) ModuleType() string {
	return "TableStoreConf"
}

func (conf RedisConfig) ModuleType() string {
	return "RedisConf"
}

func (conf MysqlConfig) ModuleType() string {
	return "MysqlConf"
}

func (conf HBaseConfig) ModuleType() string {
	return "HBaseConf"
}

func (conf HBaseThriftConfig) ModuleType() string {
	return "HBaseThriftConf"
}

func (conf FeatureStoreConfig) ModuleType() string {
	return "FeatureStoreConf"
}

func (conf BEConfig) ModuleType() string {
	return "BEConf"
}

func (conf ClickHouseConfig) ModuleType() string {
	return "ClickHouseConf"
}

func (conf LindormConfig) ModuleType() string {
	return "LindormConf"
}

func (conf GraphConfig) ModuleType() string {
	return "GraphConf"
}

func (conf OpenSearchConfig) ModuleType() string {
	return "OpenSearchConf"
}

func (conf RecallConfig) ModuleType() string {
	return "RecallConf"
}

func (conf FilterConfig) ModuleType() string {
	return "FilterConf"
}

func (conf AlgoConfig) ModuleType() string {
	return "AlgoConf"
}

func (conf SortConfig) ModuleType() string {
	return "SortConf"
}

type SceneRecallConfig map[string]CategoryConfig

func (conf SceneRecallConfig) ModuleType() string {
	return "SceneRecallConf"
}

type SceneFilterConfig []string

func (conf SceneFilterConfig) ModuleType() string {
	return "SceneFilterConf"
}

func (conf GeneralRankConfig) ModuleType() string {
	return "SceneGeneralRankConf"
}

func (conf SceneFeatureConfig) ModuleType() string {
	return "SceneFeatureConf"
}

func (conf RankConfig) ModuleType() string {
	return "SceneRankConf"
}

type SceneSortConfig []string

func (conf SceneSortConfig) ModuleType() string {
	return "SceneSortConf"
}
