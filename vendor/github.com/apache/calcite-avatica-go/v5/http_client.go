/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to you under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package avatica

import (
	"bytes"
	"context"
	"database/sql/driver"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"runtime"
	"time"

	avaticaMessage "github.com/apache/calcite-avatica-go/v5/message"
	"google.golang.org/protobuf/proto"
)

var (
	badConnRe = regexp.MustCompile(`org\.apache\.calcite\.avatica\.NoSuchConnectionException`)
)

// httpClient wraps the default http.Client to communicate with the Avatica server.
type httpClient struct {
	host       string
	httpClient *http.Client
}

type avaticaError struct {
	message *avaticaMessage.ErrorResponse
}

func (e avaticaError) Error() string {
	return fmt.Sprintf("avatica encountered an error: %s", e.message.ErrorMessage)
}

// NewHTTPClient creates a new httpClient from a host.
func NewHTTPClient(host string, config *Config) (*httpClient, error) {

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxConnsPerHost:       1,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
		},
	}
	switch config.authentication {
	case digest:
		client = WithDigestAuth(client, config.avaticaUser, config.avaticaPassword)
	case basic:
		client = WithBasicAuth(client, config.avaticaUser, config.avaticaPassword)
	case spnego:
		user := config.principal.username
		realm := config.principal.realm
		cli, err := WithKerberosAuth(client, user, realm, config.keytab, config.krb5Conf, config.krb5CredentialCache)
		if err != nil {
			return nil, fmt.Errorf("can't add kerberos authentication to http client: %w", err)
		}
		client = cli
	}

	c := &httpClient{
		host:       host,
		httpClient: client,
	}

	return c, nil
}

// post posts a protocol buffer message to the Avatica server.
func (c *httpClient) post(ctx context.Context, message proto.Message) (proto.Message, error) {

	wrapped, err := proto.Marshal(message)

	if err != nil {
		return nil, fmt.Errorf("error marshaling request message to protobuf: %w", err)
	}

	wire := &avaticaMessage.WireMessage{
		Name:           classNameFromRequest(message),
		WrappedMessage: wrapped,
	}

	body, err := proto.Marshal(wire)

	if err != nil {
		return nil, fmt.Errorf("error marshaling wire message to protobuf: %w", err)
	}

	req, err := http.NewRequest("POST", c.host, bytes.NewReader(body))

	if err != nil {
		return nil, fmt.Errorf("error creating http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-google-protobuf")

	req = req.WithContext(ctx)

	res, err := c.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("error executing http request: %w", err)
	}

	defer res.Body.Close()

	response, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	result := &avaticaMessage.WireMessage{}

	err = proto.Unmarshal(response, result)

	if err != nil {
		return nil, fmt.Errorf("error unmarshaling wire message: %w", err)
	}

	inner, err := responseFromClassName(result.Name)

	if err != nil {
		return nil, fmt.Errorf("error getting wrapped response from wire message: %w", err)
	}

	err = proto.Unmarshal(result.WrappedMessage, inner)

	if err != nil {
		return nil, fmt.Errorf("error unmarshaling wrapped message: %w", err)
	}

	if v, ok := inner.(*avaticaMessage.ErrorResponse); ok {

		for _, exception := range v.Exceptions {
			if badConnRe.MatchString(exception) {
				return nil, driver.ErrBadConn
			}
		}

		return nil, avaticaError{v}
	}

	return inner, nil
}
