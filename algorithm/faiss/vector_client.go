package faiss

import (
	"context"
	"time"

	pb "github.com/alibaba/pairec/v2/algorithm/faiss/pai_web"

	"google.golang.org/grpc"
)

type VectorClient struct {
	ServerAddress string
	Timeout       time.Duration
	VectorClient  pb.VectorRetrievalClient
}

func NewVectorClient(address string, timeout time.Duration) (*VectorClient, error) {

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := VectorClient{
		ServerAddress: address,
		Timeout:       timeout,
		VectorClient:  pb.NewVectorRetrievalClient(conn),
	}
	return &client, nil
}

func (c *VectorClient) Search(requestData interface{}) (*pb.VectorReply, error) {
	request := requestData.(*pb.VectorRequest)
	ctx, cancle := context.WithTimeout(context.Background(), c.Timeout)
	defer cancle()
	reply, err := c.VectorClient.Search(ctx, request)
	if err != nil {
		return nil, err
	}

	return reply, nil
}
