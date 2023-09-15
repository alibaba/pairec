package eas

import (
	"fmt"
	"testing"
	"time"
)

const (
	EndpointName    = "1828488879222746.cn-shanghai.pai-eas.aliyuncs.com"
	PMMLName        = "scorecard_pmml_example"
	PMMLToken       = ""
	TensorflowName  = "mnist_saved_model_example"
	TensorflowToken = ""
	TorchName       = "pytorch_resnet_example"
	TorchToken      = ""
)

func TestString(t *testing.T) {

	client := NewPredictClient(EndpointName, PMMLName)
	client.SetToken(PMMLToken)
	client.Init()
	req := "[{}]"
	client.AddHeader("headerName", "headerValue")
	resp, err := client.StringPredict(req)
	if err != nil {
		t.Fatalf(err.Error())
	} else {
		fmt.Printf("%v\n", resp)
	}
}

func TestTF(t *testing.T) {
	cli := NewPredictClient(EndpointName, TensorflowName)
	cli.SetToken(TensorflowToken)
	cli.Init()

	req := TFRequest{}
	req.SetSignatureName("predict_images")
	req.AddFeedFloat32("images", []int64{1, 784}, make([]float32, 784))

	st := time.Now()
	for i := 0; i < 10; i++ {
		resp, err := cli.TFPredict(req)
		if err != nil {
			t.Fatalf("failed to query tf model: %v", err)
		}
		fmt.Printf("%v\n", resp)
	}

	fmt.Println("average response time : ", time.Since(st)/10)
}

// TestTorch tests pytorch request and response unit test
func TestTorch(t *testing.T) {

	cli := NewPredictClient(EndpointName, TorchName)
	cli.SetTimeout(500)
	cli.SetToken(TorchToken)
	cli.Init()
	req := TorchRequest{}
	req.AddFeedFloat32(0, []int64{1, 3, 224, 224}, make([]float32, 150528))
	req.AddFetch(0)
	st := time.Now()
	for i := 0; i < 10; i++ {
		resp, err := cli.TorchPredict(req)
		if err != nil {
			t.Fatalf("failed to query torch model: %v", err)
		}
		fmt.Println(resp.GetTensorShape(0), resp.GetFloatVal(0))
	}
	fmt.Println("average response time : ", time.Since(st)/10)
}
