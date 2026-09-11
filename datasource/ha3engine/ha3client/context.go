package ha3client

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
)

// SearchRestWithContext binds the HTTP request and response body to ctx.
// Retries are managed by the caller.
func (client *Client) SearchRestWithContext(ctx context.Context, indexName *string, request *SearchRequestModel, runtime *util.RuntimeOptions) (*SearchResponseModel, error) {
	if runtime != nil {
		options := *runtime
		options.Autoretry = tea.Bool(false)
		runtime = &options
	}
	return client.searchRestWithContext(ctx, indexName, request, runtime)
}

type contextHTTPClient struct {
	ctx     context.Context
	client  *Client
	key     string
	timeout time.Duration
}

func (client *Client) contextHTTPClient(ctx context.Context, runtime map[string]interface{}) (*contextHTTPClient, error) {
	key, err := json.Marshal(runtime)
	if err != nil {
		return nil, err
	}
	return &contextHTTPClient{
		ctx: ctx, client: client, key: string(key),
		timeout: time.Duration(runtime["connectTimeout"].(int)+runtime["readTimeout"].(int)) * time.Millisecond,
	}, nil
}

func (client *contextHTTPClient) Call(request *http.Request, transport *http.Transport) (*http.Response, error) {
	key := request.URL.Scheme + "://" + request.URL.Host + client.key
	pooled, _ := client.client.httpClients.LoadOrStore(key, &http.Client{
		Transport: transport,
		Timeout:   client.timeout,
	})
	httpClient := pooled.(*http.Client)
	response, err := httpClient.Do(request.WithContext(client.ctx))
	if err != nil && client.ctx.Err() != nil {
		// Release abandoned connection attempts without interrupting active requests.
		httpClient.CloseIdleConnections()
	}
	return response, err
}
