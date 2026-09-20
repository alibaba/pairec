package eas

import (
	"math"
	"reflect"
	"testing"

	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
	pe "github.com/alibaba/pairec/v2/pkg/eas"
	"github.com/alibaba/pairec/v2/recconf"
)

func mindResponse(shape []int64, values []float32) *easyrec.TorchRecPBResponse {
	return &easyrec.TorchRecPBResponse{MapOutputs: map[string]*easyrec.ArrayProto{
		"noise":          {Dtype: easyrec.ArrayDataType_DT_FLOAT, ArrayShape: &easyrec.ArrayShape{Dim: []int64{1, 1}}, FloatVal: []float32{999}},
		"user_tower_emb": {Dtype: easyrec.ArrayDataType_DT_FLOAT, ArrayShape: &easyrec.ArrayShape{Dim: shape}, FloatVal: values},
	}, PassThroughData: map[string]string{"model_version": "items-v1"}}
}
func TestTorchrecMindEmbeddingResponse(t *testing.T) {
	parse := torchrecMindEmbeddingResponseFunc("user_tower_emb")
	got, err := parse(mindResponse([]int64{1, 3, 2}, []float32{1, 2, 0, 0, -3, 4}))
	if err != nil {
		t.Fatal(err)
	}
	r := got[0].(*TorchrecEmbeddingResponse)
	if !reflect.DeepEqual(r.GetInterests(), [][]float32{{1, 2}, {-3, 4}}) || r.GetEmbeddingSize() != 2 || r.GetPassThroughData()["model_version"] != "items-v1" {
		t.Fatalf("unexpected response: %#v", r)
	}
	got, err = parse(mindResponse([]int64{1, 2}, []float32{0, 0}))
	if err != nil {
		t.Fatal(err)
	}
	if q := got[0].(*TorchrecEmbeddingResponse).GetInterests(); q == nil || len(q) != 0 {
		t.Fatalf("all-zero interests must remain non-nil empty: %v", q)
	}
	for _, tt := range []struct {
		name   string
		shape  []int64
		values []float32
	}{
		{"batch", []int64{2, 1, 2}, []float32{1, 2, 3, 4}},
		{"length", []int64{1, 3, 2}, []float32{1, 2}},
		{"zero dimension", []int64{1, 1, 0}, nil},
		{"negative interest count", []int64{1, -1, 2}, nil},
		{"nan", []int64{1, 2}, []float32{1, float32(math.NaN())}},
		{"infinity", []int64{1, 2}, []float32{1, float32(math.Inf(1))}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parse(mindResponse(tt.shape, tt.values)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
	if _, err := torchrecMindEmbeddingResponseFunc("missing")(mindResponse([]int64{1, 2}, []float32{1, 2})); err == nil {
		t.Fatal("missing named output accepted")
	}
	bad := mindResponse([]int64{1, 2}, []float32{1, 2})
	bad.MapOutputs["user_tower_emb"].Dtype = easyrec.ArrayDataType_DT_DOUBLE
	if _, err := parse(bad); err == nil {
		t.Fatal("wrong dtype accepted")
	}
}
func TestTorchrecMindModelOutputBinding(t *testing.T) {
	for _, endpoint := range []string{pe.EndpointTypeDocker, ""} {
		conf := recconf.AlgoConfig{EasConf: recconf.EasConfig{Processor: Eas_Processor_EASYREC, EndpointType: endpoint, Url: "http://123.cn-shanghai.pai-eas.aliyuncs.com/api/predict/test", ResponseFuncName: "torchrecMindEmbeddingResponseFunc", Outputs: []string{"noise"}}}
		m := NewEasModel("mind")
		if err := m.Init(&conf); err != nil {
			t.Fatal(err)
		}
		got, err := m.request.GetResponseFunc()(mindResponse([]int64{1, 2}, []float32{1, 2}))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got[0].(*TorchrecEmbeddingResponse).GetInterests(), [][]float32{{999}}) {
			t.Fatal("output binding ignored")
		}
		conf.EasConf.Outputs = nil
		if err := NewEasModel("invalid").Init(&conf); err == nil {
			t.Fatal("missing output config accepted")
		}
	}
}
