package pairec

import (
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/alibaba/pairec/v2/log"
)

type ControllerInterface interface {
	Process(http.ResponseWriter, *http.Request)
}
type ControllerRegister struct {
	routeInfos map[string]*RouteInfo

	Middlewares []MiddlewareFunc
}

func NewControllerRegister() *ControllerRegister {
	cr := &ControllerRegister{
		routeInfos: make(map[string]*RouteInfo),
	}
	return cr
}

func (c *ControllerRegister) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			stack := string(debug.Stack())
			log.Error(fmt.Sprintf("error=%v, stack=%s", err, strings.ReplaceAll(stack, "\n", "\t")))
			Error(rw, 500, "server error")
		}
	}()

	rw = NewResponseWriter(rw)

	uri := req.RequestURI

	if index := strings.Index(req.RequestURI, "?"); index != -1 {
		uri = req.RequestURI[:index]
	}

	if info, exist := c.routeInfos[uri]; exist {

		if info.initialize != nil {
			c := info.initialize()
			c.Process(rw, req)
		} else if info.hf != nil {
			info.hf(rw, req)
		} else {
			// return 404
			Error(rw, 404, "controller not found")
			return
		}

	} else {
		// return 404
		Error(rw, 404, "url not found route info")
		return
	}
}

func (c *ControllerRegister) Register(routeInfo *RouteInfo) {
	c.routeInfos[routeInfo.pattern] = routeInfo
}

func (c *ControllerRegister) GetRoutePath() (paths []string) {
	for p := range c.routeInfos {
		paths = append(paths, p)
	}
	return
}

func (c *ControllerRegister) ApplyMiddlewares() {
	for p, r := range c.routeInfos {
		c.routeInfos[p] = applyMiddleware(r, c.Middlewares...)
	}
}

func Error(rw http.ResponseWriter, code int, msg string) {
	rw.WriteHeader(code)
	io.WriteString(rw, msg)
}

type MiddlewareController struct {
	ControllerInterface
	Middlewares []MiddlewareFunc
}

func (c MiddlewareController) Process(resp http.ResponseWriter, req *http.Request) {
	var hf handleFunc
	for i := len(c.Middlewares) - 1; i >= 0; i-- {
		if hf == nil {
			hf = c.Middlewares[i](c.ControllerInterface.Process)
		} else {
			hf = c.Middlewares[i](hf)
		}
	}
	hf(resp, req)
}

func applyMiddleware(info *RouteInfo, middleware ...MiddlewareFunc) *RouteInfo {
	if len(middleware) == 0 {
		return info
	}

	if info.initialize != nil {
		initialize := info.initialize

		info.initialize = func() ControllerInterface {
			return &MiddlewareController{
				ControllerInterface: initialize(),
				Middlewares:         middleware,
			}
		}
	}
	if info.hf != nil {
		for i := len(middleware) - 1; i >= 0; i-- {
			info.hf = middleware[i](info.hf)
		}
	}

	return info
}

type ResponseWriter struct {
	http.ResponseWriter

	statusCode int
	size       int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *ResponseWriter) Write(b []byte) (n int, err error) {
	n, err = w.ResponseWriter.Write(b)
	w.size += n

	return
}

func (w *ResponseWriter) StatusCode() int {
	return w.statusCode
}

func (w *ResponseWriter) Size() int {
	return w.size
}
