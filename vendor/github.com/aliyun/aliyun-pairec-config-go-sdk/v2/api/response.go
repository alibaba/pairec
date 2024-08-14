package api

type BaseResponse struct {
	RequestId string `json:"request_id,omitempty"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
}
type Response struct {
	BaseResponse
	Data map[string]interface{} `json:"data,omitempty"`
}
