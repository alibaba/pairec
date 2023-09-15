package pairec

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"reflect"
)

type handleFunc func(http.ResponseWriter, *http.Request)

type MiddlewareFunc func(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request)

type RouteInfo struct {
	pattern        string
	hf             handleFunc
	initialize     func() ControllerInterface
	controllerType reflect.Type
}

func Route(pattern string, c ControllerInterface) {
	reflectVal := reflect.ValueOf(c)
	t := reflect.Indirect(reflectVal).Type()
	info := RouteInfo{
		pattern:        pattern,
		controllerType: t,
	}

	info.initialize = func() ControllerInterface {
		vc := reflect.New(info.controllerType)
		execController, ok := vc.Interface().(ControllerInterface)
		if !ok {
			panic("controller is not ControllerInterface")
		}

		elemVal := reflect.ValueOf(c).Elem()
		elemType := reflect.TypeOf(c).Elem()
		execElem := reflect.ValueOf(execController).Elem()

		numOfFields := elemVal.NumField()
		for i := 0; i < numOfFields; i++ {
			fieldType := elemType.Field(i)
			elemField := execElem.FieldByName(fieldType.Name)
			if elemField.CanSet() {
				fieldVal := elemVal.Field(i)
				elemField.Set(fieldVal)
			}
		}

		return execController
	}

	PairecApp.Handlers.Register(&info)
}
func HandleFunc(pattern string, hf handleFunc) {
	info := RouteInfo{
		pattern: pattern,
		hf:      hf,
	}

	PairecApp.Handlers.Register(&info)

}

func Forward(method, url, body string) *http.Response {
	readBuf := bytes.NewBufferString(body)
	req := httptest.NewRequest(method, url, readBuf)
	w := httptest.NewRecorder()
	PairecApp.Handlers.ServeHTTP(w, req)
	return w.Result()
}
