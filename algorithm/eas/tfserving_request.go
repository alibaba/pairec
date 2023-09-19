package eas

import (
	"context"
	"errors"
	"time"

	tensorflow_serving "github.com/alibaba/pairec/v2/pkg/tensorflow_serving/apis"
	"google.golang.org/grpc/metadata"
)

type TFServingRequest struct {
	EasRequest
	SignatureName string
	ModelName     string
	Outputs       []string
	Client        tensorflow_serving.PredictionServiceClient
}

func (r *TFServingRequest) SetSignatureName(name string) {
	r.SignatureName = name
}
func (r *TFServingRequest) SetModelName(name string) {
	r.ModelName = name
}
func (r *TFServingRequest) SetOutputs(outputs []string) {
	r.Outputs = outputs
}
func (r *TFServingRequest) Invoke(requestData interface{}) (response interface{}, err error) {
	request, ok := requestData.(*tensorflow_serving.PredictRequest)
	if !ok {
		err = errors.New("requestData is not tensorflow_serving.PredictRequest type")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout*time.Millisecond)
	defer cancel()

	md := metadata.New(map[string]string{"Authorization": r.auth})
	ctx = metadata.NewOutgoingContext(ctx, md)

	response, err = r.Client.Predict(ctx, request)

	return
}
