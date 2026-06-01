package web

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/service/metrics"
	"github.com/alibaba/pairec/v2/utils"
)

var callBackControllerHandler *CallBackControllerHandler
var initOnce sync.Once

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
	initOnce.Do(func() {
		callBackControllerHandler = NewCallBackControllerHandler(poolSize, bufferSize, dropOnBackpressure)
	})
}

func ensureHandler() *CallBackControllerHandler {
	initOnce.Do(func() {
		callBackControllerHandler = NewCallBackControllerHandler(20, 5000, false)
	})
	return callBackControllerHandler
}

func Send(controller *CallBackController) {
	h := ensureHandler()

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
		h.controllerCh <- controller
	}
}

// SendDirect bypasses the HTTP self-call path used by the external
// /api/callback endpoint. It constructs a CallBackController directly
// from the provided param and enqueues it into the worker channel.
func SendDirect(param *CallBackParam) {
	if param == nil {
		log.Error("event=SendDirect\terror=nil param")
		return
	}

	if len(param.ItemList) == 0 {
		log.Info(fmt.Sprintf("requestId=%s\tevent=SendDirect\tmsg=empty item list, skip", param.RequestId))
		return
	}

	h := ensureHandler()

	c := &CallBackController{}
	c.param = *param

	if len(c.param.ComplexTypeFeatures.FeaturesMap) > 0 {
		if c.param.Features == nil {
			c.param.Features = make(FeaturesMap)
		}
		for k, v := range c.param.ComplexTypeFeatures.FeaturesMap {
			c.param.Features[k] = v
		}
	}

	if c.param.RequestId == "" {
		c.param.RequestId = utils.UUID()
	}
	c.RequestId = c.param.RequestId

	h.pending.Add(1)
	h.controllerCh <- c
}
