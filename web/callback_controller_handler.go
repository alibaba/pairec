package web

import (
	"sync"

	"github.com/alibaba/pairec/v2/utils"
)

var callBackControllerHandler *CallBackControllerHandler
var once sync.Once

type CallBackControllerHandler struct {
	controllerCh chan *CallBackController
	poolSize     int
}

func NewCallBackControllerHandler() *CallBackControllerHandler {
	handler := &CallBackControllerHandler{
		controllerCh: make(chan *CallBackController, 5000),
		poolSize:     20,
	}

	handler.start()
	return handler
}

func (h *CallBackControllerHandler) start() {
	for i := 0; i < h.poolSize; i++ {
		go func() {
			for controller := range h.controllerCh {
				controller.doCallbackLog()
			}
		}()
	}
}

func Send(controller *CallBackController) {
	if callBackControllerHandler == nil {
		once.Do(func() {
			callBackControllerHandler = NewCallBackControllerHandler()
		})
	}

	callBackControllerHandler.controllerCh <- controller
}

// SendDirect bypasses the HTTP self-call path used by the external
// /api/callback endpoint. It constructs a CallBackController directly
// from the provided param and enqueues it into the worker channel.
//
// Compared with Send, this entry skips:
//   - bytes.Reader + httptest request/response allocation
//   - io.ReadAll on the request body
//   - json.Marshal on the hook side and json.Unmarshal on the controller side
//
// The downstream worker behavior (makeCallBackContext + doCallbackLog) is
// identical to the HTTP path, so DataHub training logs are byte-for-byte
// equivalent.
func SendDirect(param *CallBackParam) {
	if callBackControllerHandler == nil {
		once.Do(func() {
			callBackControllerHandler = NewCallBackControllerHandler()
		})
	}

	c := &CallBackController{}
	c.param = *param
	if param.RequestId != "" {
		c.RequestId = param.RequestId
	} else {
		c.RequestId = utils.UUID()
	}
	c.makeCallBackContext()
	callBackControllerHandler.controllerCh <- c
}
