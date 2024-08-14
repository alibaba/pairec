package be

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

var (
	defaultRequestTimeout = 3 * time.Second
	defaultTransport      = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   100 * time.Millisecond,
			KeepAlive: 5 * time.Minute,
		}).DialContext,
		MaxIdleConns:          1000,
		MaxIdleConnsPerHost:   1000,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
	}
	defaultHttpClient = &http.Client{
		Timeout:   defaultRequestTimeout,
		Transport: defaultTransport,
	}
)

func request(client *Client, method, uri string, headers map[string]string, body []byte) (*http.Response, error) {
	return realRequest(client, method, uri, headers, body)
}

// request sends a request to Be Service.
// @note if error is nil, you must call http.Response.Body.Close() to finalize reader
func realRequest(client *Client, method, uri string, headers map[string]string,
	body []byte) (*http.Response, error) {

	headers["Host"] = client.Endpoint

	digest, err := signature(client)
	if err != nil {
		return nil, NewClientError(err)
	}
	auth := fmt.Sprintf("Basic %v", digest)
	headers["Authorization"] = auth

	// Initialize http request
	reader := bytes.NewReader(body)

	// Handle the endpoint
	urlStr := fmt.Sprintf("%s/%s", client.Endpoint, uri)
	req, err := http.NewRequest(method, urlStr, reader)
	if err != nil {
		return nil, NewClientError(err)
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Parse the be error from body.
	if resp.StatusCode != http.StatusOK {
		err := &BadResponseError{}
		err.HTTPCode = resp.StatusCode
		defer resp.Body.Close()
		buf, ioErr := ioutil.ReadAll(resp.Body)
		if ioErr != nil {
			return nil, NewBadResponseError(ioErr.Error(), resp.Header, resp.StatusCode)
		}
		err.RespBody = string(buf)
		err.RespHeader = resp.Header
		return nil, err
	}
	return resp, nil
}

func signature(client *Client) (string, error) {
	if client.UserName == "" || client.PassWord == "" {
		return "", NewClientError(fmt.Errorf("Empty userName or passWord"))
	}
	auth := client.UserName + ":" + client.PassWord
	return base64.StdEncoding.EncodeToString([]byte(auth)), nil
}
