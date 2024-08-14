package web

type Response struct {
	Code      int    `json:"code"`
	Message   string `json:"msg"`
	RequestId string `json:"request_id"`
}
