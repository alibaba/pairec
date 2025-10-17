package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/utils/compress"
)

const (
	SUCCESS_CODE         int = 200
	ERROR_PARAMETER_CODE int = 400
	SERVER_ERROR_CODE    int = 500
)

var (
	CODE_MAPS = map[int]string{
		SUCCESS_CODE:         "success",
		ERROR_PARAMETER_CODE: "parammeter error",
		SERVER_ERROR_CODE:    "server error",
	}
)

type ErrorResponse struct {
	Response
}

func (e *ErrorResponse) ToString() string {
	j, _ := json.Marshal(e)
	return string(j)
}

type Controller struct {
	RequestBody []byte
	RequestId   string
	Start       time.Time
	End         time.Time
}

func (c *Controller) cost() int64 {
	duration := c.End.UnixNano() - c.Start.UnixNano()

	return duration / 1e6
}
func (c *Controller) ReadRequestBody(r *http.Request) ([]byte, error) {
	encoding := r.Header.Get("Content-Encoding")
	if encoding != "" {
		switch encoding {
		case "zstd":
			return compress.ZstdDecode(r.Body)
		case "lz4":
			return compress.LZ4Decode(r.Body)
		case "gzip":
			return compress.GzipDecode(r.Body)
		default:
		}
	}
	return io.ReadAll(r.Body)
}

func (c *Controller) LogRequestBegin(r *http.Request) {
	info := fmt.Sprintf("requestId=%s\tevent=begin\turi=%s\taddress=%s\tbody=%s", c.RequestId, r.RequestURI, r.RemoteAddr, string(c.RequestBody))
	log.Info(info)
}
func (c *Controller) LogRequestBeginWithSize(r *http.Request, size int) {
	if len(c.RequestBody) > size {
		info := fmt.Sprintf("requestId=%s\tevent=begin\turi=%s\taddress=%s\tbody=%s\tsize=%d", c.RequestId, r.RequestURI, r.RemoteAddr, string(c.RequestBody[0:size]), len(c.RequestBody))
		log.Info(info)
	} else {
		info := fmt.Sprintf("requestId=%s\tevent=begin\turi=%s\taddress=%s\tbody=%s", c.RequestId, r.RequestURI, r.RemoteAddr, string(c.RequestBody))
		log.Info(info)
	}
}
func (c *Controller) LogRequestEnd(r *http.Request) {
	info := fmt.Sprintf("requestId=%s\tevent=end\turi=%s\tcost=%d", c.RequestId, r.RequestURI, c.cost())
	log.Info(info)
}

func (c *Controller) SendError(w http.ResponseWriter, code int, msg string) {
	errInfo := fmt.Sprintf("requestId=%s\tbody=%s\terr=%s", c.RequestId, string(c.RequestBody), msg)
	log.Error(errInfo)
	e := ErrorResponse{
		Response: Response{
			Code:      code,
			Message:   msg,
			RequestId: c.RequestId,
		},
	}

	io.WriteString(w, e.ToString())
}
