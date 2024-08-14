package api

import (
	"unicode/utf8"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"

	paifeaturestore "github.com/alibabacloud-go/paifeaturestore-20230621/client"
)

var (

/*
*

	defaultHttpClient = &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   100,
			MaxConnsPerHost:       100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

*
*/
)

// APIClient manages communication with the Pairec Experiment Restful Api API v1.0.0
// In most cases there should be only one, shared, APIClient.
type APIClient struct {
	*paifeaturestore.Client
	cfg    *Configuration
	common service // Reuse a single struct instead of allocating one for each service on the heap.

	instanceId string

	// API Services

	FsProjectApi *FsProjectApiService

	InstanceApi *InstanceApiService

	DatasourceApi *DatasourceApiService

	FeatureEntityApi *FeatureEntityApiService

	FeatureViewApi *FeatureViewApiService

	FsModelApi *FsModelApiService
}

type service struct {
	client *APIClient
}

// NewAPIClient creates a new API client. Requires a userAgent string describing your application.
// optionally a custom http.Client to allow for advanced features such as caching.
func NewAPIClient(cfg *Configuration) (*APIClient, error) {

	c := &APIClient{
		cfg: cfg,
	}
	endpoint := cfg.GetDomain()
	config := &openapi.Config{
		AccessKeyId:     &cfg.AccessKeyId,
		AccessKeySecret: &cfg.AccessKeySecret,
		Endpoint:        &endpoint,
	}

	client, err := paifeaturestore.NewClient(config)
	if err != nil {
		return nil, err
	}

	c.Client = client
	c.common.client = c

	// API Services
	c.FsProjectApi = (*FsProjectApiService)(&c.common)
	c.InstanceApi = (*InstanceApiService)(&c.common)
	c.DatasourceApi = (*DatasourceApiService)(&c.common)
	c.FeatureEntityApi = (*FeatureEntityApiService)(&c.common)
	c.FeatureViewApi = (*FeatureViewApiService)(&c.common)
	c.FsModelApi = (*FsModelApiService)(&c.common)

	return c, nil
}

func (c *APIClient) GetConfig() *Configuration {
	return c.cfg
}

func strlen(s string) int {
	return utf8.RuneCountInString(s)
}
