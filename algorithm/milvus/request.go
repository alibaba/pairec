package milvus

import "github.com/milvus-io/milvus-sdk-go/v2/entity"

type MilvusRequest struct {
	TopK           int
	CollectionName string
	VectorField    string
	OutputFields   []string
	Vectors        []entity.Vector
	Partitions     []string
	Expr           string
	MetricType     entity.MetricType
	SearchParams   *MilvusSearchParams
}

func NewMilvusRequest() *MilvusRequest {
	return &MilvusRequest{
		MetricType: entity.IP,
	}
}

func (r *MilvusRequest) SetMetricType(metricType string) {
	switch metricType {
	case "IP":
		r.MetricType = entity.IP
	case "L2":
		r.MetricType = entity.L2
	}
}

func (r *MilvusRequest) SetSearchParams(params map[string]interface{}) {
	r.SearchParams = &MilvusSearchParams{params: params}
}

func (r *MilvusRequest) AppendVectors(vectors []float32) {
	r.Vectors = append(r.Vectors, entity.FloatVector(vectors))
}

type MilvusSearchParams struct {
	params map[string]interface{}
}

func (sp *MilvusSearchParams) Params() map[string]interface{} {
	return sp.params
}
