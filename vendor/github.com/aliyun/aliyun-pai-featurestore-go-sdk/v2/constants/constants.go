package constants

type FSType int

const (
	FS_INT32 FSType = iota + 1 // int32
	FS_INT64                   // int64
	FS_FLOAT
	FS_DOUBLE
	FS_STRING
	FS_BOOLEAN
	FS_TIMESTAMP
)
const (
	Datasource_Type_MaxCompute = "maxcompute"
	Datasource_Type_Hologres   = "hologres"
	Datasource_Type_Redis      = "redis"
	Datasource_Type_Mysql      = "mysql"
	Datasource_Type_IGraph     = "igraph"
	Datasource_Type_Spark      = "spark"
	Datasource_Type_TableStore = "tablestore"
)
const (
	Feature_View_Type_Batch    = "Batch"
	Feature_View_Type_Stream   = "Stream"
	Feature_View_Type_Sequence = "Sequence"
)
