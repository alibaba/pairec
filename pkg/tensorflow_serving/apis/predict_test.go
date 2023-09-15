package apis

import (
	context "context"
	"fmt"
	"testing"
	"time"

	framework "github.com/alibaba/pairec/pkg/tensorflow/core/framework"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
)

func TestGrpcRequest(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	serverAddr := "saved-model-half-plus-two-cpu.1730760139076263.cn-beijing.pai-eas.aliyuncs.com:80"
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}

	defer conn.Close()
	client := NewPredictionServiceClient(conn)

	spec := &ModelSpec{
		Name:          "saved_model_half_plus_two_cpu",
		SignatureName: "serving_default",
	}
	request := &PredictRequest{
		ModelSpec: spec,
		Inputs:    make(map[string]*framework.TensorProto),
	}
	tp := framework.TensorProto{
		FloatVal: []float32{1, 3, 8},
		Dtype:    framework.DataType_DT_FLOAT,
		TensorShape: &framework.TensorShapeProto{
			Dim: []*framework.TensorShapeProto_Dim{{Size: 3}},
		},
	}

	request.Inputs["x"] = &tp

	md := metadata.New(map[string]string{"Authorization": "ZTJjMjAzZjJjMDY3Y2FjMGE5ODIzZGQzODg4YjYwNWQwOTU4NTk2Mw=="})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := client.Predict(ctx, request)

	if err != nil {
		t.Fatal(err)
	}
	for k, v := range resp.Outputs {
		fmt.Println(k, v)
	}

}
