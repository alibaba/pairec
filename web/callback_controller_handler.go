package web

import (
	"sync"
	"sync/atomic"

	"github.com/alibaba/pairec/v2/log"
)

var callBackControllerHandler *CallBackControllerHandler
var once sync.Once

type CallBackControllerHandler struct {
	controllerCh chan *CallBackController
	poolSize     int
	pending      atomic.Int64
}

func NewCallBackControllerHandler(poolSize int, bufferSize int) *CallBackControllerHandler {
	if poolSize <= 0 {
		poolSize = 20
	}
	if bufferSize <= 0 {
		bufferSize = 5000
	}
	handler := &CallBackControllerHandler{
		controllerCh: make(chan *CallBackController, bufferSize),
		poolSize:     poolSize,
	}
	handler.start()
	return handler
}

func (h *CallBackControllerHandler) start() {
	for i := 0; i < h.poolSize; i++ {
		go func() {
			for controller := range h.controllerCh {
				controller.doCallbackLog()
				h.pending.Add(-1)
			}
		}()
	}
}

func (h *CallBackControllerHandler) Pending() int64 {
	return h.pending.Load()
}

func CallbackPending() int64 {
	if callBackControllerHandler == nil {
		return 0
	}
	return callBackControllerHandler.Pending()
}

func Send(controller *CallBackController) {
	if callBackControllerHandler == nil {
		once.Do(func() {
			callBackControllerHandler = NewCallBackControllerHandler(20, 5000)
		})
	}

	select {
	case callBackControllerHandler.controllerCh <- controller:
		callBackControllerHandler.pending.Add(1)
	default:
		log.Warning("callback channel full, dropping request")
	}
}
