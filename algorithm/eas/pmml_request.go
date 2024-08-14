package eas

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
)

type PMMLRequest struct {
	EasRequest
}

func (r *PMMLRequest) Invoke(requestData interface{}) (body interface{}, err error) {
	data, _ := json.Marshal(requestData)
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err = zw.Write(data)
	if err != nil {
		return nil, err
	}
	zw.Close()

	req, err := http.NewRequest("POST", r.url, &buf)
	if err != nil {
		return
	}

	headers := map[string][]string{
		"Authorization":    {r.auth},
		"Content-Encoding": {"gzip"},
	}
	req.Header = headers

	easClient.Timeout = r.timeout

	response, err := easClient.Do(req)
	if err != nil {
		return
	}

	defer response.Body.Close()

	reqBody, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	body = string(reqBody)

	return
}
