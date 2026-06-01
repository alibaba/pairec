package web

import (
	"sync"
	"sync/atomic"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/service/metrics"
)

var callBackControllerHandler *CallBackControllerHandler
var once sync.Once

func init() {
	metrics.CallbackPendingFunc = CallbackPending
}

type CallBackControllerHandler struct {
	controllerCh       chan *CallBackController
	poolSize           int
	pending            atomic.Int64
	dropOnBackpressure bool
}

func NewCallBackControllerHandler(poolSize int, bufferSize int, dropOnBackpressure bool) *CallBackControllerHandler {
	if poolSize <= 0 {
		poolSize = 20
	}
	if bufferSize <= 0 {
		bufferSize = 5000
	}
	handler := &CallBackControllerHandler{
		controllerCh:       make(chan *CallBackController, bufferSize),
		poolSize:           poolSize,
		dropOnBackpressure: dropOnBackpressure,
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

// InitHandler initializes the callback handler with config values.
// Must be called during service startup after config is loaded.
func InitHandler(poolSize int, bufferSize int, dropOnBackpressure bool) {
	once.Do(func() {
		callBackControllerHandler = NewCallBackControllerHandler(poolSize, bufferSize, dropOnBackpressure)
	})
}

func Send(controller *CallBackController) {
	if callBackControllerHandler == nil {
		once.Do(func() {
			callBackControllerHandler = NewCallBackControllerHandler(20, 5000, false)
		})
	}

	// Increment pending BEFORE enqueue to avoid transient negative values
	h := callBackControllerHandler
	if h.dropOnBackpressure {
		h.pending.Add(1)
		select {
		case h.controllerCh <- controller:
			// already counted
		default:
			h.pending.Add(-1)
			log.Warning("callback channel full, dropping request")
		}
	} else {
		h.pending.Add(1)
		h.controllerCh <- controller // block if full
	}
}
