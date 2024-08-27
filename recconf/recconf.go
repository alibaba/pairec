package recconf

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"

	"github.com/alibaba/pairec/v2/config"
)

var (
	Config      *RecommendConfig
	adapterName = "json"
	//pairecConfigAdapterName = "pairec_config"
)

const (
	DaoConf_Adapter_Mysql        = "mysql"
	DaoConf_Adapter_Redis        = "redis"
	DaoConf_Adapter_TableStore   = "tablestore"
	DaoConf_Adapter_HBase        = "hbase"
	DaoConf_Adapter_Hologres     = "hologres"
	DataSource_Type_Kafka        = "kafka"
	DataSource_Type_Datahub      = "datahub"
	DataSource_Type_ClickHouse   = "clickhouse"
	DataSource_Type_BE           = "be"
	DataSource_Type_Lindorm      = "lindorm"
	DataSource_Type_HBase_Thrift = "hbase_thrift"
	DataSource_Type_FeatureStore = "featurestore"
	Datasource_Type_Graph        = "graph"

	BE_RecallType_X2I        = "x2i_recall"
	BE_RecallType_Vector     = "vector_recall"
	BE_RecallType_MultiMerge = "multi_merge_recall"
)

func init() {
	Config = newRecommendConfig()
}

type RecommendConfig struct {
	RunMode                   string // Run Mode: daily | product
	Region                    string
	ListenConf                ListenConfig
	UserFeatureConfs          map[string]SceneFeatureConfig
	FeatureConfs              map[string]SceneFeatureConfig
	SortNames                 map[string][]string
	FilterNames               map[string][]string
	AlgoConfs                 []AlgoConfig
	RecallConfs               []RecallConfig
	FilterConfs               []FilterConfig
	BeFilterConfs             []BeFilterConfig
	SortConfs                 []SortConfig
	RedisConfs                map[string]RedisConfig
	MysqlConfs                map[string]MysqlConfig
	ClickHouseConfs           map[string]ClickHouseConfig
	HologresConfs             map[string]HologresConfig
	LindormConfs              map[string]LindormConfig
	GraphConfs                map[string]GraphConfig
	FeatureStoreConfs         map[string]FeatureStoreConfig
	KafkaConfs                map[string]KafkaConfig
	SlsConfs                  map[string]SlsConfig
	DatahubConfs              map[string]DatahubConfig
	BEConfs                   map[string]BEConfig
	Ha3EngineConfs            map[string]Ha3EngineConfig
	OpenSearchConfs           map[string]OpenSearchConfig
	HBaseConfs                map[string]HBaseConfig
	HBaseThriftConfs          map[string]HBaseThriftConfig
	TableStoreConfs           map[string]TableStoreConfig
	SceneConfs                map[string]map[string]CategoryConfig
	RankConf                  map[string]RankConfig
	LogConf                   LogConfig
	ABTestConf                ABTestConfig
	CallBackConfs             map[string]CallBackConfig
	EmbeddingConfs            map[string]EmbeddingConfig
	GeneralRankConfs          map[string]GeneralRankConfig
	ColdStartGeneralRankConfs map[string]ColdStartGeneralRankConfig
	ColdStartRankConfs        map[string]ColdStartRankConfig
	DPPConf                   []DPPSortConfig
	DebugConfs                map[string]DebugConfig
	PipelineConfs             map[string][]PipelineConfig
	PrometheusConfig          PrometheusConfig
	UserDefineConfs           json.RawMessage
}
type ListenConfig struct {
	HttpAddr string
	HttpPort int
}
type PrometheusConfig struct {
	Enable           bool
	Subsystem        string
	PushGatewayURL   string
	PushIntervalSecs int
	Job              string
	ReqDurBuckets    []float64
	ReqSizeBuckets   []float64
	RespSizeBuckets  []float64
}
type DaoConfig struct {
	Adapter             string
	AdapterType         string
	RedisName           string
	RedisPrefix         string
	RedisDataType       string
	RedisFieldType      string
	RedisDefaultKey     string
	RedisValueDelimeter string
	MysqlName           string
	MysqlTable          string
	Config              string
	TableStoreName      string
	TableStoreTableName string
	HBasePrefix         string
	HBaseName           string
	HBaseTable          string
	ColumnFamily        string
	Qualifier           string

	ItemIdField    string
	ItemScoreField string

	// hologres
	HologresName      string
	HologresTableName string

	// clickhouse
	ClickHouseName      string
	ClickHouseTableName string

	// be engine
	BeName               string
	BizName              string
	BeRecallName         string
	BeTableName          string
	BeExposureUserIdName string
	BeExposureItemIdName string

	// feature store
	FeatureStoreName       string
	FeatureStoreModelName  string
	FeatureStoreEntityName string
	FeatureStoreViewName   string

	// graph
	GraphName  string
	InstanceId string
	TableName  string
	UserNode   string
	ItemNode   string
	Edge       string
	// lindorm
	LindormTableName string
	LindormName      string
}
type SceneFeatureConfig struct {
	FeatureLoadConfs []FeatureLoadConfig
	AsynLoadFeature  bool
}
type FeatureLoadConfig struct {
	FeatureDaoConf FeatureDaoConfig
	Features       []FeatureConfig
}
type BeRTCntFieldConfig struct {
	FieldNames []string
	// if not set, Delims is empty string by default
	// which indicates it is single value with no delimiter
	Delims []string
	// if not set, Alias is set to FieldNames[0] by default
	Alias string
}
type FeatureDaoConfig struct {
	DaoConfig
	NoUsePlayTimeField      bool
	FeatureKey              string
	FeatureStore            string // user or item
	UserFeatureKeyName      string
	ItemFeatureKeyName      string
	TimestampFeatureKeyName string
	EventFeatureKeyName     string
	PlayTimeFeatureKeyName  string
	TsFeatureKeyName        string
	UserSelectFields        string
	ItemSelectFields        string
	// FeatureAsyncLoad use async goroutine to load feature
	FeatureAsyncLoad bool
	FeatureType      string // per feature type has different way of build
	SequenceLength   int
	SequenceName     string
	SequenceEvent    string
	SequenceDelim    string
	// SequencePlayTime filter event by as least play time
	// like play event need to large than 10s, so set value is "play:10000", timeunit is ms
	// if has  more than one event to filter, use ';' as delim , like "play:10000;read:50000"
	SequencePlayTime         string
	SequenceOfflineTableName string
	// SequenceDimFields fetch other dimension fields from db
	SequenceDimFields string

	BeItemFeatureKeyName      string
	BeTimestampFeatureKeyName string
	BeEventFeatureKeyName     string
	BePlayTimeFeatureKeyName  string

	BeIsHomeField string
	// separate by , if combined by multiple fields separate by :
	// [depreciated], please set BeRTCntFieldInfo.FieldNames instead
	BeRTCntFields    string
	BeRTCntFieldInfo []BeRTCntFieldConfig
	BeRTTable        string

	RTCntWins     string
	RTCntMaxKey   int
	RTCntWinDelay int

	OutRTCntFeaPattern     string
	OutHomeRTCntFeaPattern string

	// [depreciated], please set BeRTCntFieldInfo.Alias instead
	OutRTCntFieldAlias string
	OutRTCntWinNames   string
	OutEventName       string

	// CacheFeaturesName When use UserFeatureConfs, can cache load features to the user cacheFeatures map.
	// Only valid when FeatureStore = user
	CacheFeaturesName string

	// LoadFromCacheFeaturesName load features from user cacheFeatures map
	// Only valid when FeatureStore = user
	LoadFromCacheFeaturesName string
}
type FeatureConfig struct {
	FeatureType         string
	FeatureName         string
	FeatureSource       string
	FeatureValue        string
	FeatureStore        string // user or item
	RemoveFeatureSource bool   // delete feature source
	Normalizer          string
	Expression          string
}

type AlgoConfig struct {
	Name          string
	Type          string
	EasConf       EasConfig
	VectorConf    VectorConfig
	MilvusConf    MilvusConfig
	LookupConf    LookupConfig
	SeldonConf    SeldonConfig
	TFservingConf TFservingConfig
}

type PIDControllerConfig struct {
	SyncPIDStatus          bool
	AllocateExperimentWise bool
	MaxItemCacheSize       int
	MaxItemCacheTime       int
	RedisName              string
	RedisKeyPrefix         string
	TimeWindow             int
	DefaultKp              float32
	DefaultKi              float32
	DefaultKd              float32
	Timestamp              int64
	BoostScoreConditions   []BoostScoreCondition
}

type LookupConfig struct {
	FieldName string
}

type EasConfig struct {
	Processor        string
	Url              string
	Auth             string
	EndpointType     string
	SignatureName    string
	Timeout          int
	RetryTimes       int
	ResponseFuncName string
	Outputs          []string
	ModelName        string
}
type TFservingConfig struct {
	Url              string
	SignatureName    string
	Timeout          int
	RetryTimes       int
	ResponseFuncName string
	Outputs          []string
}
type SeldonConfig struct {
	Url              string
	ResponseFuncName string
}
type VectorConfig struct {
	ServerAddress string
	Timeout       int64
}
type MilvusConfig struct {
	ServerAddress string
	Timeout       int64
}
type RecallConfig struct {
	Name         string
	RecallType   string
	RecallCount  int
	RecallAlgo   string
	ItemType     string
	CacheAdapter string
	CacheConfig  string
	CachePrefix  string
	CacheTime    int // cache time by seconds
	Triggers     []TriggerConfig

	HologresVectorConf       HologresVectorConfig
	BeVectorConf             BeVectorConfig
	MilvusVectorConf         MilvusVectorConfig
	UserCollaborativeDaoConf UserCollaborativeDaoConfig
	ItemCollaborativeDaoConf ItemCollaborativeDaoConfig
	User2ItemDaoConf         User2ItemDaoConfig
	UserTopicDaoConf         UserTopicDaoConfig
	DaoConf                  DaoConfig
	VectorDaoConf            VectorDaoConfig
	ColdStartDaoConf         ColdStartDaoConfig
	RealTimeUser2ItemDaoConf RealTimeUser2ItemDaoConfig
	UserFeatureConfs         []FeatureLoadConfig // get user features

	// be recall config
	BeConf         BeConfig
	GraphConf      GraphConf
	OpenSearchConf OpenSearchConf
}

type GraphConf struct {
	GraphName   string
	ItemId      string
	QueryString string
	Params      []string
}

type OpenSearchConf struct {
	OpenSearchName string
	AppName        string
	ItemId         string
	RequestParams  map[string]any
	Params         []string
}

type BeConfig struct {
	Count             int
	BeName            string
	BizName           string
	BeRecallType      string
	RecallNameMapping map[string]RecallNameMappingConfig
	BeRecallParams    []BeRecallParam
	BeFilterNames     []string
	BeABParams        map[string]interface{}
}
type RecallNameMappingConfig struct {
	Format string
	Fields []string
}
type BeRecallParam struct {
	Count        int
	Priority     int
	RecallType   string
	RecallName   string
	ScorerClause string
	TriggerType  string // user or be or fixvalue or user_vector
	UserTriggers []TriggerConfig
	TriggerValue string //
	TriggerParam BeTriggerParam
	//RecallParamName   string
	UserVectorTrigger            UserVectorTriggerConfig
	UserTriggerDaoConf           UserTriggerDaoConfig               // online table for u2i
	UserTriggerRulesConf         UserTriggerRulesConfig             // be recall diversity trigger, trigger have diff recall count
	UserCollaborativeDaoConf     UserCollaborativeDaoConfig         // offline table for u2i
	UserRealtimeEmbeddingTrigger UserRealtimeEmbeddingTriggerConfig // get user feature and invoke eas model, get item embedding sink to be
	UserEmbeddingO2OTrigger      UserEmbeddingO2OTriggerConfig

	ItemIdName      string
	TriggerIdName   string
	RecallTableName string
	DiversityParam  string
	CustomParams    map[string]interface{}
}
type UserTriggerRulesConfig struct {
	DefaultValue  int
	TriggerCounts []int
}
type UserVectorTriggerConfig struct {
	CacheTime        int
	CachePrefix      string
	RecallAlgo       string
	UserFeatureConfs []FeatureLoadConfig // get user features
}
type UserEmbeddingO2OTriggerConfig struct {
	BizName             string
	RecallName          string
	BeName              string
	SeqDelimiter        string              // seq feature delimiter
	MultiValueDelimiter string              // multi value feature delimiter
	UserFeatureConfs    []FeatureLoadConfig // get user features
}
type UserRealtimeEmbeddingTriggerConfig struct {
	Debug              bool
	DebugLogDatahub    string
	EmbeddingNum       int
	RecallAlgo         string
	DistinctParamName  string
	DistinctParamValue string
	UserFeatureConfs   []FeatureLoadConfig // get user features
}

type BeTriggerParam struct {
	BizName   string
	FieldName string
}
type ColdStartDaoConfig struct {
	SqlDaoConfig
	TimeInterval int // second
}
type SqlDaoConfig struct {
	DaoConfig
	Limit        int
	WhereClause  string
	PrimaryKey   string
	SelectFields string
}
type RealTimeUser2ItemDaoConfig struct {
	UserTriggerDaoConf    UserTriggerDaoConfig
	Item2ItemTable        string
	SimilarItemIdField    string
	SimilarItemScoreField string
}
type UserTriggerDaoConfig struct {
	SqlDaoConfig
	NoUsePlayTimeField bool
	ItemCount          int
	TriggerCount       int
	EventPlayTime      string
	EventWeight        string
	WeightExpression   string
	WeightMode         string
	PropertyFields     []string
	DiversityRules     []TriggerDiversityRuleConfig

	BeItemFeatureKeyName      string
	BeTimestampFeatureKeyName string
	BeEventFeatureKeyName     string
	BePlayTimeFeatureKeyName  string
}
type TriggerDiversityRuleConfig struct {
	Dimensions []string
	Size       int
}
type HologresVectorConfig struct {
	HologresName         string
	VectorTable          string // example: "item_emb_{partition}", '{partition}' will be replaced by partition info
	VectorKeyField       string
	VectorEmbeddingField string
	WhereClause          string
	TimeInterval         int
}
type BeVectorConfig struct {
	BizName              string //
	VectorKeyField       string
	VectorEmbeddingField string
}
type MilvusVectorConfig struct {
	VectorKeyField       string
	VectorEmbeddingField string
	CollectionName       string
	MetricType           string
	SearchParams         map[string]interface{}
}
type UserCollaborativeDaoConfig struct {
	DaoConfig
	User2ItemTable string
	Item2ItemTable string

	Normalization string // set "on" to enable it, otherwise set "off", enabled by default
}
type ItemCollaborativeDaoConfig struct {
	DaoConfig
	Item2ItemTable string
}
type User2ItemDaoConfig struct {
	DaoConfig
	User2ItemTable string
	Item2ItemTable string
}
type UserTopicDaoConfig struct {
	DaoConfig
	UserTopicTable string
	TopicItemTable string
	IndexName      string
}

type VectorDaoConfig struct {
	DaoConfig
	EmbeddingField string
	KeyField       string

	// set the following fields to get partition info,
	// if not set, '{partition}' in table name won't be replaced (if it exists)
	PartitionInfoTable string
	PartitionInfoField string
}

type GraphConfig struct {
	Host     string
	UserName string
	Passwd   string
}

type RedisConfig struct {
	Host           string
	Port           int
	Password       string
	DbNum          int
	MaxIdle        int
	ConnectTimeout int
	ReadTimeout    int
	WriteTimeout   int
}
type MysqlConfig struct {
	DSN string
}
type ClickHouseConfig struct {
	DSN string
}
type HologresConfig struct {
	DSN string
}
type LindormConfig struct {
	Url      string
	User     string
	Password string
	Database string
}
type FeatureStoreConfig struct {
	AccessId  string
	AccessKey string
	RegionId  string

	ProjectName       string
	FeatureDBUsername string
	FeatureDBPassword string
	HologresPort      int
}
type KafkaConfig struct {
	BootstrapServers string
	Topic            string
}
type DatahubConfig struct {
	AccessId    string
	AccessKey   string
	Endpoint    string
	ProjectName string
	TopicName   string
	Schemas     []DatahubTopicSchema
}
type BEConfig struct {
	Username    string
	Password    string
	Endpoint    string
	ReleaseType string // values: product or dev or prepub
}
type Ha3EngineConfig struct {
	Username   string
	Password   string
	Endpoint   string
	InstanceId string
}
type OpenSearchConfig struct {
	EndPoint        string
	AccessKeyId     string
	AccessKeySecret string
}
type DatahubTopicSchema struct {
	Field string

	//Type is the type of the datahub tuple field,valid value is string, integer
	Type string
}
type HBaseConfig struct {
	ZKQuorum string
}
type HBaseThriftConfig struct {
	Host     string
	User     string
	Password string
}
type TableStoreConfig struct {
	EndPoint        string
	InstanceName    string
	AccessKeyId     string
	AccessKeySecret string
	RoleArn         string
}
type SlsConfig struct {
	EndPoint        string
	AccessKeyId     string
	AccessKeySecret string
	ProjectName     string
	LogstoreName    string
}
type SceneConfig struct {
	Categories []string
}
type CategoryConfig struct {
	RecallNames []string
}
type RankConfig struct {
	RankAlgoList    []string
	RankScore       string
	Processor       string
	ContextFeatures []string
	BatchCount      int
	ScoreRewrite    map[string]string
	ASTType         string
}
type ActionConfig struct {
	ActionType string
	ActionName string
}
type OperatorValueConfig struct {
	Type string // "property", "function"
	Name string
	From string // item or user
}
type LogConfig struct {
	RetensionDays int
	DiskSize      int    // unit : G, if value = 20, the true size is 20G
	LogLevel      string // valid value is DEBUG, INFO , ERROR , FATAL
	Output        string // valid value is file, console
	SLSName       string
}
type ABTestConfig struct {
	Host  string
	Token string
}
type FilterConfig struct {
	Name                      string
	FilterType                string
	DaoConf                   DaoConfig
	MaxItems                  int
	TimeInterval              int // second
	RetainNum                 int
	ShuffleItem               bool
	WriteLog                  bool
	ClearLogIfNotEnoughScene  string
	Dimension                 string
	ScoreWeight               float64
	GroupMinNum               int
	GroupMaxNum               int
	GroupWeightStrategy       string
	GroupWeightDimensionLimit map[string]int
	WriteLogExcludeScenes     []string
	GenerateItemDataFuncName  string
	AdjustCountConfs          []AdjustCountConfig
	ItemStateDaoConf          ItemStateDaoConfig
	FilterEvaluableExpression string
	FilterParams              []FilterParamConfig
	DiversityDaoConf          DiversityDaoConfig
	DiversityMinCount         int
	EnsureDiversity           bool
	FilterVal                 FilterValue
	ItemStateCacheSize        int
	ItemStateCacheTime        int
	Conditions                []FilterParamConfig

	ConditionFilterConfs struct {
		FilterConfs []struct {
			Conditions []FilterParamConfig
			FilterName string
		}
		DefaultFilterName string
	}
}
type BeFilterConfig struct {
	FilterConfig
}
type FilterValue struct {
	SelectCol   string
	WhereClause string
}
type SortConfig struct {
	Debug                         bool
	RemainItem                    bool
	Name                          string
	SortType                      string
	SortByField                   string
	SwitchThreshold               float64
	DiversitySize                 int
	Size                          int
	DPPConf                       DPPSortConfig
	SSDConf                       SSDSortConfig
	PIDConf                       PIDControllerConfig
	MixSortRules                  []MixSortConfig
	BoostScoreConditionsFilterAll bool
	BoostScoreConditions          []BoostScoreCondition
	DistinctIdConditions          []DistinctIdCondition
	Conditions                    []FilterParamConfig
	ExcludeRecalls                []string
	DiversityRules                []DiversityRuleConfig
	TimeInterval                  int
	BoostScoreByWeightDao         BoostScoreByWeightDaoConfig
}

type BoostScoreByWeightDaoConfig struct {
	DaoConfig
	ItemFieldName   string
	WeightFieldName string
}

type MixSortConfig struct {
	MixStrategy   string // fix_position, random_position
	Positions     []int
	PositionField string
	Number        int
	NumberRate    float64
	RecallNames   []string
	Conditions    []FilterParamConfig
}
type DiversityRuleConfig struct {
	Dimensions    []string
	IntervalSize  int
	WindowSize    int
	FrequencySize int
}
type FilterParamConfig struct {
	Name     string
	Domain   string
	Operator string
	Type     string // string, int, int64
	Value    interface{}
}
type BoostScoreCondition struct {
	Expression string
	Conditions []FilterParamConfig
}
type DistinctIdCondition struct {
	DistinctId int
	Conditions []FilterParamConfig
}
type ItemStateDaoConfig struct {
	DaoConfig
	ItemFieldName string
	WhereClause   string
	SelectFields  string
}
type DiversityDaoConfig struct {
	DaoConfig
	ItemKeyField       string
	DistinctFields     []string
	CacheTimeInMinutes int
}
type AdjustCountConfig struct {
	RecallName string
	Count      int
	Type       string
}
type CallBackConfig struct {
	DataSource      DataSourceConfig
	RankConf        RankConfig
	RawFeatures     bool
	RawFeaturesRate int
}
type EmbeddingConfig struct {
	DataSource DataSourceConfig
	RankConf   RankConfig
}
type GeneralRankConfig struct {
	FeatureLoadConfs []FeatureLoadConfig
	RankConf         RankConfig
	ActionConfs      []ActionConfig
}
type ColdStartGeneralRankConfig struct {
	GeneralRankConfig
	RecallNames []string
}
type ColdStartRankConfig struct {
	RecallName           string
	AlgoName             string
	OnlyEmbeddingFeature bool
}

type DataSourceConfig struct {
	Name string
	Type string
}

type TriggerConfig struct {
	TriggerKey   string
	DefaultValue string
	Boundaries   []int
}

type DPPSortConfig struct {
	Name               string
	DaoConf            DaoConfig
	TableName          string
	TableSuffixParam   string
	TablePKey          string
	EmbeddingColumn    string
	EmbeddingSeparator string
	Alpha              float64
	CacheTimeInMinutes int
	EmbeddingHookNames []string
	NormalizeEmb       string
	WindowSize         int
	AbortRunCount      int
	CandidateCount     int
	MinScorePercent    float64
	EmbMissedThreshold float64
	FilterRetrieveIds  []string
	EnsurePositiveSim  string
}
type SSDSortConfig struct {
	Name               string
	DaoConf            DaoConfig
	TableName          string
	TableSuffixParam   string
	TablePKey          string
	EmbeddingColumn    string
	EmbeddingSeparator string
	Gamma              float64
	UseSSDStar         bool
	CacheTimeInMinutes int
	NormalizeEmb       string
	WindowSize         int
	AbortRunCount      int
	CandidateCount     int
	MinScorePercent    float64
	EmbMissedThreshold float64
	FilterRetrieveIds  []string
	EnsurePositiveSim  string
	Condition          *BoostScoreCondition
}
type DebugConfig struct {
	Rate       int
	DebugUsers []string
	// OutputType represent log write to console or datahub or file
	OutputType  string
	DatahubName string
	KafKaName   string
	FilePath    string
	MaxFileNum  int
}

type PipelineConfig struct {
	Name              string
	RecallNames       []string
	FilterNames       []string
	GeneralRankConf   GeneralRankConfig
	FeatureLoadConfs  []FeatureLoadConfig
	RankConf          RankConfig
	ColdStartRankConf ColdStartRankConfig
	SortNames         []string
}

func newRecommendConfig() *RecommendConfig {
	conf := RecommendConfig{
		RunMode: "daily",
		ListenConf: ListenConfig{
			HttpAddr: "",
			HttpPort: 8000,
		},
	}

	return &conf
}

func CopyConfig(src, dst *RecommendConfig, filters ...func(string) bool) {
	srcVal := reflect.ValueOf(src).Elem()
	srcType := reflect.TypeOf(src).Elem()
	dstVal := reflect.ValueOf(dst).Elem()

	numOfFields := srcVal.NumField()
	for i := 0; i < numOfFields; i++ {
		fieldType := srcType.Field(i)
		flag := true
		for _, filter := range filters {
			flag = filter(fieldType.Name)
			if !flag {
				break
			}
		}
		if !flag {
			continue
		}
		elemField := dstVal.FieldByName(fieldType.Name)
		if elemField.CanSet() {
			fieldVal := srcVal.Field(i)
			elemField.Set(fieldVal)
		}
	}
}

func loadConfigFromFile(filePath string) error {
	configer, err := config.NewConfig(adapterName, filePath)
	if err != nil {
		return err
	}
	rawdata := configer.RawData()

	err = json.Unmarshal(rawdata, Config)
	if err != nil {
		return err
	}

	return nil
}

// LoadConfig load config from file or pairec config server
// First check the environment CONFIG_NAME, if exist, load config data from pairec config server
func LoadConfig(filePath string) error {

	if filePath == "" {
		filePath = os.Getenv("CONFIG_PATH")
	}

	if filePath == "" {
		return errors.New("config file path empty")
	}

	return loadConfigFromFile(filePath)
}

var notifyCh = make([]chan *RecommendConfig, 0)

func Subscribe() <-chan *RecommendConfig {
	ch := make(chan *RecommendConfig)
	notifyCh = append(notifyCh, ch)
	return ch
}

func UpdateConf(conf *RecommendConfig) {
	Config = conf
	go func() {
		for _, ch := range notifyCh {
			ch <- conf
		}
	}()
}
