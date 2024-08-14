package web

import "sync"

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
			for {
				select {
				case controller := <-h.controllerCh:
					controller.doCallbackLog()
				}
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
