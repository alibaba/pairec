package aliyun_igraph_go_sdk

import (
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"net/url"
	"strings"
	"time"
)

const (
	REQUEST_METHOD = "GET"
)

type Client struct {
	Endpoint   string
	UserName   string
	PassWord   string
	Src        string
	httpClient *fasthttp.Client
}

func NewClient(endpoint string, userName string, passWord string, src string) *Client {
	if len(src) == 0 {
		src = userName + "_" + endpoint
	}
	return &Client{
		Endpoint:   endpoint,
		UserName:   userName,
		PassWord:   passWord,
		Src:        src,
		httpClient: defaultHttpClient,
	}
}

// WithRequestTimeout with custom timeout for a request
func (c *Client) WithRequestTimeout(timeout time.Duration) *Client {
	c.httpClient.ReadTimeout = timeout
	c.httpClient.WriteTimeout = timeout
	return c
}

func (c *Client) buildReadUrl(readRequest *ReadRequest) url.URL {
	uri := url.URL{Path: "app"}
	src := c.Src
	rawUrl := readRequest.BuildUri()
	uri.RawQuery = strings.Join([]string{"app=gremlin", "src=" + src, rawUrl}, "&")
	return uri
}

func (c *Client) Read(readRequest *ReadRequest) (*Response, error) {
	vErr := readRequest.Validate()
	if vErr != nil {
		return nil, vErr
	}
	if len(c.Src) == 0 {
		return nil, InvalidParamsError{"Src is empty"}
	}

	buildUri := c.buildReadUrl(readRequest)
	uri := buildUri.RequestURI()
	headers := map[string]string{}

	body, statusCode, err := request(c, REQUEST_METHOD, uri, headers, nil)

	if err != nil {
		return nil, err
	}

	readResult := ReadResult{}
	if jErr := json.Unmarshal(body, &readResult); jErr != nil {
		fmt.Println(jErr)
		return nil, NewBadResponseError("Illegal readResult:"+string(body), nil, statusCode)
	}

	var resp *Response
	if len(readResult.ErrorInfo) == 0 {
		resp = NewResponse(readResult.Result)
	} else {
		return nil, NewBadResponseError(fmt.Sprintf("Failed to read, message:%v",
			readResult.ErrorInfo), nil, statusCode)
	}
	return resp, nil
}

func (c *Client) Write(writeRequest *WriteRequest) (*Response, error) {
	vErr := writeRequest.Validate()
	if vErr != nil {
		return nil, vErr
	}
	buildUri := writeRequest.BuildUri()
	uri := buildUri.RequestURI()
	headers := map[string]string{}

	body, statusCode, err := request(c, REQUEST_METHOD, uri, headers, nil)
	if err != nil {
		return nil, err
	}
	writeResult := WriteResult{}
	if jErr := json.Unmarshal(body, &writeResult); jErr != nil {
		fmt.Println(jErr)
		return nil, NewBadResponseError("Illegal writeResult:"+string(body), nil, statusCode)
	}

	switch writeResult.Errno {
	case 0:
		results := []*Result{}
		return NewResponse(results), nil
	case 1:
		return nil, NewBadResponseError(fmt.Sprintf("Failed to write, illegal reqeust body, errorCode[%v], resp:[%v]",
			writeResult.Errno, string(body)), nil, statusCode)
	default:
		return nil, NewBadResponseError(fmt.Sprintf("Failed to write, errorCode[%v], resp:[%v]",
			writeResult.Errno, string(body)), nil, statusCode)
	}

}
