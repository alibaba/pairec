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
var once sync.Once

func init() {
	metrics.CallbackPendingFunc = CallbackPending
}

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
// equivalent. makeCallBackContext is intentionally NOT invoked here; it is
// called by the worker goroutine inside doCallbackLog, exactly like the
// HTTP path. Calling it here would trigger a duplicated A/B experiment
// match RPC and a duplicated experiment log line.
func SendDirect(param *CallBackParam) {
	// Guard against nil param to avoid a nil pointer dereference panic.
	// SendDirect is invoked from the recommend main path (RecommendCleanHook),
	// so a panic here would crash the request goroutine. Logging the error
	// keeps the misuse visible without taking down the request.
	if param == nil {
		log.Error("event=SendDirect\terror=nil param")
		return
	}

	// Mirror the empty ItemList rejection that CallBackController.CheckParameter
	// enforces on the HTTP /api/callback path (web/callback_controller.go:109).
	// Without this, an empty ItemList reaches CallBackService.Rank and panics
	// at a nil algoData. Owning the invariant here makes SendDirect spec-
	// equivalent to the HTTP entry, so any future caller is automatically safe.
	if len(param.ItemList) == 0 {
		log.Info(fmt.Sprintf("requestId=%s\tevent=SendDirect\tmsg=empty item list, skip", param.RequestId))
		return
	}

	if callBackControllerHandler == nil {
		once.Do(func() {
			callBackControllerHandler = NewCallBackControllerHandler(20, 5000)
		})
	}

	c := &CallBackController{}
	c.param = *param

	// Mirror the merge logic from CheckParameter so both entry paths
	// produce the same param.Features for downstream consumers
	// (NewUserWithContext copies these into user.Properties, which
	// drives the DataHub user_features field).
	if len(c.param.ComplexTypeFeatures.FeaturesMap) > 0 {
		if c.param.Features == nil {
			c.param.Features = make(FeaturesMap)
		}
		for k, v := range c.param.ComplexTypeFeatures.FeaturesMap {
			c.param.Features[k] = v
		}
	}

	// Fall back to a generated UUID when the caller does not set RequestId.
	// Write it back to c.param so doCallbackLog logs the same id as the one
	// used by the controller and the recommend context.
	if c.param.RequestId == "" {
		c.param.RequestId = utils.UUID()
	}
	c.RequestId = c.param.RequestId

	callBackControllerHandler.controllerCh <- c
}
