package aliyun_igraph_go_sdk

import (
	"encoding/base64"
	"fmt"
	"github.com/valyala/fasthttp"
	"net/http"
	"time"
)

var (
	defaultRequestTimeout = 3 * time.Second
	defaultHttpClient     = &fasthttp.Client{
		MaxConnsPerHost: 200,
		ReadTimeout:     defaultRequestTimeout,
		WriteTimeout:    defaultRequestTimeout,
	}
)

func request(client *Client, method, uri string, headers map[string]string, body []byte) ([]byte, int, error) {
	return realRequest(client, method, uri, headers, body)
}

// request sends a request to Be Service.
// @note if error is nil, you must call http.Response.Body.Close() to finalize reader
func realRequest(client *Client, method, uri string, headers map[string]string,
	body []byte) ([]byte, int, error) {

	headers["Host"] = client.Endpoint
	digest, err := signature(client)
	if err != nil {
		return nil, 0, NewClientError(err)
	}
	auth := fmt.Sprintf("Basic %v", digest)
	headers["Authorization"] = auth

	// Initialize http request
	// Handle the endpoint
	urlStr := fmt.Sprintf("%s/%s", client.Endpoint, uri)

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer func() {
		// release resource
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()
	req.SetRequestURI(urlStr)
	req.Header.SetMethod(method)
	req.SetBody(body)
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	err = client.httpClient.Do(req, resp)
	if err != nil {
		return nil, 0, err
	}

	respBody := resp.Body()
	buf := make([]byte, len(respBody))
	copy(buf, respBody)
	// Parse the be error from body.
	if resp.StatusCode() != http.StatusOK {
		err := &BadResponseError{}
		err.HTTPCode = resp.StatusCode()
		err.RespBody = string(buf)
		return nil, resp.StatusCode(), err
	}
	return buf, resp.StatusCode(), nil
}

func signature(client *Client) (string, error) {
	if client.UserName == "" || client.PassWord == "" {
		return "", NewClientError(fmt.Errorf("Empty userName or passWord"))
	}
	auth := client.UserName + ":" + client.PassWord
	return base64.StdEncoding.EncodeToString([]byte(auth)), nil
}
