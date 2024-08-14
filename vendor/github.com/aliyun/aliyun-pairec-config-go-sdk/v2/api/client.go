package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/pairecservice"
)

var (
	defaultTransport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			d := net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}
			return d.DialContext(ctx, "tcp4", addr)
		},
		MaxIdleConns:          200,
		MaxIdleConnsPerHost:   200,
		MaxConnsPerHost:       200,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
)

// APIClient manages communication with the Pairec Experiment Restful Api API v1.0.0
// In most cases there should be only one, shared, APIClient.
type APIClient struct {
	*pairecservice.Client

	common service // Reuse a single struct instead of allocating one for each service on the heap.

	region string

	domain string

	// API Services
	ExperimentApi *ExperimentApiService

	ExperimentGroupApi *ExperimentGroupApiService

	ExperimentRoomApi *ExperimentRoomApiService

	LayerApi *LayerApiService

	SceneApi *SceneApiService

	ParamApi *ParamApiService

	CrowdApi *CrowdApiService

	FlowCtrlApi *FlowCtrlApiService

	FeatureConsistencyCheckApi *FeatureConsistencyCheckService
}

type service struct {
	client     *APIClient
	instanceId string
}

// NewAPIClient creates a new API client. Requires a userAgent string describing your application.
// optionally a custom http.Client to allow for advanced features such as caching.
func NewAPIClient(instanceId, region, accessId, accessKey string) (*APIClient, error) {
	client, err := pairecservice.NewClientWithAccessKey(region, accessId, accessKey)
	if err != nil {
		return nil, err
	}
	client.SetTransport(defaultTransport)
	c := &APIClient{
		Client: client,
		region: region,
	}
	c.common.client = c
	c.common.instanceId = instanceId

	// API Services
	c.ExperimentApi = (*ExperimentApiService)(&c.common)
	c.ExperimentGroupApi = (*ExperimentGroupApiService)(&c.common)
	c.ExperimentRoomApi = (*ExperimentRoomApiService)(&c.common)
	c.LayerApi = (*LayerApiService)(&c.common)
	c.SceneApi = (*SceneApiService)(&c.common)
	c.ParamApi = (*ParamApiService)(&c.common)
	c.CrowdApi = (*CrowdApiService)(&c.common)
	c.FlowCtrlApi = (*FlowCtrlApiService)(&c.common)
	c.FeatureConsistencyCheckApi = (*FeatureConsistencyCheckService)(&c.common)

	return c, nil
}

func (c *APIClient) GetDomain() string {
	if c.domain == "" {
		c.domain = fmt.Sprintf("pairecservice-vpc.%s.aliyuncs.com", c.region)
	}

	return c.domain
}

func (c *APIClient) SetDomain(domain string) {
	c.domain = domain
}

/**
func (c *APIClient) Init(accessId, accessKey string) error {
	endpoint := c.GetDomain()
	protol := "http"
	config := &openapi.Config{
		AccessKeyId:     &accessId,
		AccessKeySecret: &accessKey,
		Endpoint:        &endpoint,
		Protocol:        &protol,
	}

	client, err := pairecserviceV2.NewClient(config)

	if err != nil {
		return err
	}

	c.v2Client = client
	return nil
}

**/
