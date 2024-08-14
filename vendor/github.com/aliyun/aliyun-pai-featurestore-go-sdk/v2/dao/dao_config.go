package dao

import (
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/constants"
)

type DaoConfig struct {
	DatasourceType      string
	RedisName           string
	RedisPrefix         string
	RedisDefaultKey     string
	RedisValueDelimeter string
	MysqlName           string
	MysqlTableName      string
	Config              string
	HBasePrefix         string
	HBaseName           string
	HBaseTable          string
	ColumnFamily        string
	Qualifier           string

	PrimaryKeyField string
	EventTimeField  string
	TTL             int

	// hologres
	HologresName      string
	HologresTableName string

	//tablestore
	TableStoreName      string
	TableStoreTableName string

	// clickhouse
	ClickHouseName      string
	ClickHouseTableName string

	// be engine
	BeName               string
	BizName              string
	BeTableName          string
	BeExposureUserIdName string
	BeExposureItemIdName string

	// igraph
	IGraphName        string
	GroupName         string
	LabelName         string
	SaveOriginalField bool

	FieldMap map[string]string
	// redis, tablestore
	FieldTypeMap map[string]constants.FSType

	// redis
	Fields []string

	// hologres sequence tables
	HologresOnlineTableName  string
	HologresOfflineTableName string
	// tablestore sequence tables
	TableStoreOnlineTableName  string
	TableStoreOfflineTableName string
	// igraph sequence table
	IgraphEdgeName string
}
