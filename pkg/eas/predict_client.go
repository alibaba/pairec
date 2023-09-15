package eas

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"
)

const (
	// Default endpoint is the gateway mode
	EndpointTypeDefault = "DEFAULT"
	// Vipserver endpoint is only used for services which registered
	// vipsever domains in alibaba internal clusters
	EndpointTypeVipserver = "VIPSERVER"
	// Direct endpoint is used for direct accessing to the services' instances
	// both inside a eas service and in user's client ecs
	EndpointTypeDirect = "DIRECT"

	// use eas docker to deploy service, but not use pai eas
	EndpointTypeDocker = "DOCKER"
)

const (
	ErrorCodeServiceDiscovery = 510
	ErrorCodeCreateRequest    = 511
	ErrorCodePerformRequest   = 512
	ErrorCodeReadResponse     = 513
)

// PredictError is a custom err type
type PredictError struct {
	Code       int
	Message    string
	RequestURL string
}

// Error for error interface
func (err *PredictError) Error() string {
	return fmt.Sprintf("Url: [%v] Code: [%d], Message: [%s]", err.RequestURL, err.Code, err.Message)
}

// NewPredictError constructs an error
func NewPredictError(code int, url string, msg string) *PredictError {
	return &PredictError{
		Code:       code,
		Message:    msg,
		RequestURL: url,
	}
}

// PredictClient for accessing prediction service by creating a fixed size connection pool
// to perform the request through established persistent connections.
type PredictClient struct {
	retryCount         int
	maxConnectionCount int
	token              string
	headers            map[string]string
	host               string
	endpoint           Endpoint
	endpointType       string
	endpointName       string
	serviceName        string
	stop               int32
	client             http.Client
}

// NewPredictClient returns an instance of PredictClient
func NewPredictClient(endpointName string, serviceName string) *PredictClient {
	return &PredictClient{
		endpointName: endpointName,
		serviceName:  serviceName,
		retryCount:   5,
		stop:         0,
		headers:      map[string]string{},
		client: http.Client{
			Timeout: 5000 * time.Millisecond,
			Transport: &http.Transport{
				MaxConnsPerHost: 100,
			},
		},
	}
}

// Init initializes the predict client to create and enable endpoint discovery
func (p *PredictClient) Init() error {
	switch p.endpointType {
	case "":
		p.endpoint = newGatewayEndpoint(p.endpointName)
	case EndpointTypeDefault:
		p.endpoint = newGatewayEndpoint(p.endpointName)
	case EndpointTypeVipserver:
		p.endpoint = newVipServerEndpoint(p.endpointName)
		go p.syncHandler()
	case EndpointTypeDirect:
		p.endpoint = newCacheServerEndpoint(p.endpointName, p.serviceName)
		go p.syncHandler()
	default:
		return NewPredictError(http.StatusBadRequest, "", "Unsupported endpoint type: "+p.endpointType)
	}
	return nil
}

// Shutdown after called this client instance should not be used again
func (p *PredictClient) Shutdown() {
	atomic.StoreInt32(&(p.stop), 1)
}

// syncHandler synchronizes the services's endpoints from the upstream discovery server periodically
func (p *PredictClient) syncHandler() {
	p.endpoint.Sync()
	for {
		select {
		// Sync endpoints from upstream every 3 seconds
		case <-time.NewTimer(time.Second * 3).C:
			if 1 == atomic.LoadInt32(&(p.stop)) {
				break
			}
			p.endpoint.Sync()
		}
	}
}

// SetEndpoint sets service's endpoint for client
func (p *PredictClient) SetEndpoint(endpointName string) {
	p.endpointName = endpointName
}

// SetEndpointType sets endpoint type for client
func (p *PredictClient) SetEndpointType(endpointType string) {
	p.endpointType = endpointType
}

// SetToken function sets service's access token for client
func (p *PredictClient) SetToken(token string) {
	p.token = token
}

func (p *PredictClient) AddHeader(headerName, headerValue string) {
	p.headers[headerName] = headerValue
}

func (p *PredictClient) SetHost(host string) {
	p.host = host
}

// SetRetryCount sets max retry count for client
func (p *PredictClient) SetRetryCount(cnt int) {
	p.retryCount = cnt
}

// SetHttpTransport sets http transport argument for go http client
func (p *PredictClient) SetHttpTransport(transport *http.Transport) {
	p.client.Transport = transport
}

// SetTimeout set the request timeout for client, 5000ms by default
func (p *PredictClient) SetTimeout(timeout int) {
	p.client.Timeout = time.Duration(timeout) * time.Millisecond
}

// SetServiceName sets target service name for client
func (p *PredictClient) SetServiceName(serviceName string) {
	p.serviceName = serviceName
}

func (p *PredictClient) tryNext(host string) string {
	return p.endpoint.TryNext(host)
}

func (p *PredictClient) createUrl(host string) string {
	if len(p.serviceName) != 0 {
		if p.serviceName[len(p.serviceName)-1] == '/' {
			p.serviceName = p.serviceName[:len(p.serviceName)-1]
		}
	}
	return fmt.Sprintf("http://%s/api/predict/%s", host, p.serviceName)
}

// generateSignature computes the signature header using the access token with hmac sha1 algorithm.
// returns the headers including signature header for authentication.
func (p *PredictClient) generateSignature(requestData []byte) map[string]string {
	//canonicalizedResource := fmt.Sprintf("/api/predict/%s", p.serviceName)
	//contentMd5 := md5sum(requestData)
	contentType := "application/octet-stream"
	//currentTime := time.Now().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	//verb := "POST"

	//auth := fmt.Sprintf("%s\n%s\n%s\n%s\n%s", verb, contentMd5, contentType, currentTime, canonicalizedResource)
	//authorization := fmt.Sprintf("EAS %s", hmacSha256(auth, p.token))

	return map[string]string{
		//"Content-MD5":    contentMd5,
		//"Date":           currentTime,
		"Content-Type":   contentType,
		"Content-Length": fmt.Sprintf("%d", len(requestData)),
		//"Authorization":  authorization,
		"Authorization": p.token,
	}
}

// BytesPredict send the raw request data in byte array through http connections,
// retry the request automatically when an error occurs
func (p *PredictClient) BytesPredict(requestData []byte) ([]byte, error) {
	host := p.tryNext("")
	headers := p.generateSignature(requestData)
	for i := 0; i <= p.retryCount; i++ {
		if i != 0 {
			host = p.tryNext(host)
		}

		if len(host) == 0 {
			return nil, NewPredictError(ErrorCodeServiceDiscovery, host,
				fmt.Sprintf("No available endpoint found for service: %v", p.serviceName))
		}

		url := p.createUrl(host)

		req, err := http.NewRequest("POST", url, bytes.NewReader(requestData))
		if err != nil {
			// retry
			if i != p.retryCount {
				continue
			}
			return nil, NewPredictError(ErrorCodeCreateRequest, url, err.Error())
		}
		if p.token != "" {
			for headerName, headerValue := range headers {
				req.Header.Set(headerName, headerValue)
			}
		}

		for headerName, headerValue := range p.headers {
			req.Header.Set(headerName, headerValue)
		}

		if p.host != "" {
			req.Host = p.host
		}

		resp, err := p.client.Do(req)
		if err != nil {
			// retry
			if i != p.retryCount {
				continue
			}
			return nil, NewPredictError(ErrorCodePerformRequest, url, err.Error())
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// retry
			if i != p.retryCount {
				continue
			}
			return nil, NewPredictError(ErrorCodeReadResponse, url, err.Error())
		}
		resp.Body.Close()

		if resp.StatusCode != 200 {
			// retry
			if i != p.retryCount {
				continue
			}
			return body, NewPredictError(resp.StatusCode, url, string(body))
		}
		return body, nil
	}
	return []byte{}, nil
}

type Request interface {
	ToString() (string, error)
}

type Response interface {
	unmarshal(body []byte) error
}

// Predict for request
func (p *PredictClient) Predict(request Request) (Response, error) {
	req, err2 := request.ToString()
	if err2 != nil {
		return nil, err2
	}
	body, err := p.BytesPredict([]byte(req))
	if err != nil {
		return nil, err
	}

	switch request.(type) {
	case TFRequest:
		resp := TFResponse{}
		unmarshalErr := resp.unmarshal(body)
		return &resp, unmarshalErr
	case TorchRequest:
		resp := TorchResponse{}
		unmarshalErr := resp.unmarshal(body)
		return &resp, unmarshalErr
	default:
		return nil, NewPredictError(-1, "", "Unknown request type, currently support StringRequest, TFRequest and TorchRequest.")
	}
}

// StringPredict function send input data and return predicted result
func (p *PredictClient) StringPredict(str string) (string, error) {
	body, err := p.BytesPredict([]byte(str))
	return string(body), err
}

// TorchPredict function send input data and return PyTorch predicted result
func (p *PredictClient) TorchPredict(request TorchRequest) (*TorchResponse, error) {
	resp, err := p.Predict(request)
	if err != nil {
		return nil, err
	}
	return resp.(*TorchResponse), err
}

// TFPredict function send input data and return TensorFlow predicted result
func (p *PredictClient) TFPredict(request TFRequest) (*TFResponse, error) {
	resp, err := p.Predict(request)
	if err != nil {
		return nil, err
	}
	return resp.(*TFResponse), err
}
