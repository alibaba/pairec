package proxy

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/alibaba/pairec/v2"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/web"
)

type ResponseFunc func(w http.ResponseWriter, requestBody, responseBody []byte) bool

type RequestInfo struct {
	Path   string
	Method string

	// Timeout how max millisecond to invoke the Path
	Timeout int

	ResponseFunc ResponseFunc

	ErrorResponseFunc ResponseFunc
}

type ProxyController struct {
	web.Controller
	RequestInfo *RequestInfo
}

func (c *ProxyController) Process(w http.ResponseWriter, r *http.Request) {
	c.Start = time.Now()
	var err error
	c.RequestBody, err = ioutil.ReadAll(r.Body)
	if err != nil {
		c.SendError(w, web.ERROR_PARAMETER_CODE, "read parammeter error")
		return
	}
	if len(c.RequestBody) == 0 {
		c.SendError(w, web.ERROR_PARAMETER_CODE, "request body empty")
		return
	}

	c.doProcess(w, r)
	c.End = time.Now()
	c.LogRequestEnd(r)
}

func (c *ProxyController) doProcess(w http.ResponseWriter, r *http.Request) {
	responseCh := make(chan *http.Response, 1)

	go c.asyncInvoke(responseCh)

	select {
	case response := <-responseCh:
		body, _ := ioutil.ReadAll(response.Body)
		response.Body.Close()
		if response.StatusCode >= http.StatusBadRequest {
			if c.RequestInfo.ErrorResponseFunc(w, c.RequestBody, nil) {
				c.dealResult(w, nil, web.SERVER_ERROR_CODE, "server error")
			}
		} else {
			if c.RequestInfo.ResponseFunc(w, c.RequestBody, body) {
				c.dealResult(w, body, web.SUCCESS_CODE, "success")
			}
		}
	case <-time.After(time.Duration(c.RequestInfo.Timeout) * time.Millisecond):
		log.Warning(fmt.Sprintf("%s async load timeout %dms", c.RequestInfo.Path, c.RequestInfo.Timeout))
		if c.RequestInfo.ErrorResponseFunc(w, c.RequestBody, nil) {
			c.dealResult(w, nil, web.SERVER_ERROR_CODE, "timeout")
		}
	}
}

func (c *ProxyController) asyncInvoke(ch chan<- *http.Response) {
	response := pairec.Forward(c.RequestInfo.Method, c.RequestInfo.Path, string(c.RequestBody))
	ch <- response
}

func (c *ProxyController) dealResult(w http.ResponseWriter, body []byte, code int, message string) {
	if code == web.SUCCESS_CODE {
		io.WriteString(w, string(body))
	} else {
		c.SendError(w, code, message)
	}
}
