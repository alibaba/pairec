package aliyun_igraph_go_sdk

import "encoding/json"

type InvalidParamsError struct {
	Message string
}

func (e InvalidParamsError) String() string {
	return e.Message
}

func (e InvalidParamsError) Error() string {
	return e.String()
}

// BadResponseError define be http bad response error
type BadResponseError struct {
	RespBody   string
	RespHeader map[string][]string
	HTTPCode   int
}

func (e BadResponseError) String() string {
	b, err := json.MarshalIndent(e, "", "    ")
	if err != nil {
		return ""
	}
	return string(b)
}

func (e BadResponseError) Error() string {
	return e.String()
}

func NewBadResponseError(body string, header map[string][]string, httpCode int) *BadResponseError {
	return &BadResponseError{
		RespBody:   body,
		RespHeader: header,
		HTTPCode:   httpCode,
	}
}

// ClientError defines be client error
type ClientError struct {
	Message string `json:"errorMessage"`
}

// NewClientError new client error
func NewClientError(err error) *ClientError {
	if err == nil {
		return nil
	}
	if clientError, ok := err.(*ClientError); ok {
		return clientError
	}
	clientError := new(ClientError)
	clientError.Message = err.Error()
	return clientError
}

func (e ClientError) String() string {
	b, err := json.MarshalIndent(e, "", "    ")
	if err != nil {
		return ""
	}
	return string(b)
}

func (e ClientError) Error() string {
	return e.String()
}
