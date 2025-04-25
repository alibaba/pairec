package recconf

import (
	"fmt"
)

type Dependent interface {
	Requirements() Requirements
}

type Constraint func(module any) error

type Requirements map[ModuleIndex][]Constraint

func (r Requirements) Add(modType string, modName string, constraints ...Constraint) {
	index := ModuleIndex{
		Type: modType,
		Name: modName,
	}

	r[index] = append(r[index], constraints...)
}

func (r Requirements) Check(modules map[ModuleIndex]any) error {
	for requireModuleIndex, constraints := range r {
		jsonPath := modJsonPath[requireModuleIndex.Type]

		if requireModule, ok := modules[requireModuleIndex]; !ok {
			if jsonPath != "" {
				return fmt.Errorf("%s (%s name) is not defined in %s", requireModuleIndex.Name, requireModuleIndex.Type, jsonPath)
			}

			return fmt.Errorf("%s (%s name) is undefined", requireModuleIndex.Name, requireModuleIndex.Type)
		} else {
			for _, constraint := range constraints {
				if err := constraint(requireModule); err != nil {
					if jsonPath != "" {
						return fmt.Errorf("%s (%s name in %s) does not satisfy the requirement, details: %s",
							requireModuleIndex.Name, requireModuleIndex.Type, jsonPath, err.Error())
					}
					return fmt.Errorf("%s (%s name) does not satisfy the requirement, details: %s",
						requireModuleIndex.Name, requireModuleIndex.Type, err.Error())
				}
			}
		}
	}

	return nil
}

func newRequirements() Requirements {
	return map[ModuleIndex][]Constraint{}
}

func (conf RecallConfig) Requirements() Requirements {
	requirements := newRequirements()

	addDaoRequirements(conf.DaoConf, requirements)
	addDaoRequirements(conf.UserCollaborativeDaoConf.DaoConfig, requirements)
	addDaoRequirements(conf.RealTimeUser2ItemDaoConf.UserTriggerDaoConf.SqlDaoConfig.DaoConfig, requirements)
	addDaoRequirements(conf.VectorDaoConf.DaoConfig, requirements)
	for _, featureConf := range conf.UserFeatureConfs {
		addDaoRequirements(featureConf.FeatureDaoConf.DaoConfig, requirements)
	}
	if conf.GraphConf.GraphName != "" {
		requirements.Add(GraphConfig{}.ModuleType(), conf.GraphConf.GraphName)
	}
	addDaoRequirements(conf.ColdStartDaoConf.DaoConfig, requirements)
	addDaoRequirements(conf.ItemCollaborativeDaoConf.DaoConfig, requirements)
	if conf.OpenSearchConf.OpenSearchName != "" {
		requirements.Add(OpenSearchConfig{}.ModuleType(), conf.OpenSearchConf.OpenSearchName)
	}

	return requirements
}

func (conf FilterConfig) Requirements() Requirements {
	requirements := newRequirements()

	addDaoRequirements(conf.DaoConf, requirements)

	return requirements
}

func (conf SortConfig) Requirements() Requirements {
	requirements := newRequirements()

	addDaoRequirements(conf.DPPConf.DaoConf, requirements)
	addDaoRequirements(conf.SSDConf.DaoConf, requirements)
	addDaoRequirements(conf.BoostScoreByWeightDao.DaoConfig, requirements)

	return requirements
}

var builtInRecalls = map[string]bool{
	"ContextItemRecall": true,
}

var builtInFilters = map[string]bool{
	"UniqueFilter": true,
}

var builtInSort = map[string]bool{
	"ItemRankScore": true,
}

func (conf SceneRecallConfig) Requirements() Requirements {
	requirements := newRequirements()

	for _, config := range conf {
		for _, name := range config.RecallNames {
			if builtInRecalls[name] {
				continue
			}

			requirements.Add(RecallConfig{}.ModuleType(), name)
		}
	}

	return requirements
}

func (conf SceneFilterConfig) Requirements() Requirements {
	requirements := newRequirements()

	for _, name := range conf {
		if builtInFilters[name] {
			continue
		}

		requirements.Add(FilterConfig{}.ModuleType(), name)
	}

	return requirements
}

func (conf GeneralRankConfig) Requirements() Requirements {
	requirements := newRequirements()

	for _, loadConf := range conf.FeatureLoadConfs {
		addDaoRequirements(loadConf.FeatureDaoConf.DaoConfig, requirements)
	}

	for _, algoName := range conf.RankConf.RankAlgoList {
		requirements.Add(AlgoConfig{}.ModuleType(), algoName)
	}

	return requirements
}

func (conf SceneFeatureConfig) Requirements() Requirements {
	requirements := newRequirements()

	for _, loadConf := range conf.FeatureLoadConfs {
		addDaoRequirements(loadConf.FeatureDaoConf.DaoConfig, requirements)
	}

	return requirements
}

func (conf RankConfig) Requirements() Requirements {
	requirements := newRequirements()

	for _, algoName := range conf.RankAlgoList {
		requirements.Add(AlgoConfig{}.ModuleType(), algoName)
	}

	return requirements
}

func (conf SceneSortConfig) Requirements() Requirements {
	requirements := newRequirements()

	for _, name := range conf {
		if builtInSort[name] {
			continue
		}

		requirements.Add(FilterConfig{}.ModuleType(), name)
	}

	return requirements
}

func addDaoRequirements(conf DaoConfig, requirements Requirements) {
	switch conf.AdapterType {
	case DaoConf_Adapter_Hologres:
		requirements.Add(HologresConfig{}.ModuleType(), conf.HologresName)
	case DaoConf_Adapter_TableStore:
		requirements.Add(TableStoreConfig{}.ModuleType(), conf.TableStoreName)
	case DaoConf_Adapter_Redis:
		requirements.Add(RedisConfig{}.ModuleType(), conf.RedisName)
	case DaoConf_Adapter_Mysql:
		requirements.Add(MysqlConfig{}.ModuleType(), conf.MysqlName)
	case DaoConf_Adapter_HBase:
		requirements.Add(HBaseConfig{}.ModuleType(), conf.HBaseName)
	case DataSource_Type_FeatureStore:
		requirements.Add(FeatureStoreConfig{}.ModuleType(), conf.FeatureStoreName)
	case DataSource_Type_BE:
		requirements.Add(BEConfig{}.ModuleType(), conf.BeName)
	case DataSource_Type_ClickHouse:
		requirements.Add(ClickHouseConfig{}.ModuleType(), conf.ClickHouseName)
	case DataSource_Type_Lindorm:
		requirements.Add(LindormConfig{}.ModuleType(), conf.LindormName)
	case Datasource_Type_Graph:
		requirements.Add(GraphConfig{}.ModuleType(), conf.GraphName)
	case DataSource_Type_HBase_Thrift:
		requirements.Add(HBaseThriftConfig{}.ModuleType(), conf.HBaseName)
	}
}
