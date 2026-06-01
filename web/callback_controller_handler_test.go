package web

import (
	"testing"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
)

// TestSendDirect_NilParam verifies SendDirect does not panic when called with
// a nil *CallBackParam. SendDirect is invoked from the recommend main path
// (RecommendCleanHook), so a nil dereference panic here would crash the
// request goroutine. The guard must short-circuit before the *param deref.
func TestSendDirect_NilParam(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SendDirect(nil) panicked: %v", r)
		}
	}()

	// Should return cleanly without enqueueing anything.
	SendDirect(nil)
}

// TestSendDirect_EmptyItemList verifies SendDirect rejects a CallBackParam
// with an empty ItemList, mirroring CallBackController.CheckParameter on
// the HTTP /api/callback path. Without this guard, the worker would call
// CallBackService.Rank with no items, which would dereference a nil
// algoData and crash the process (see
// service/call_back_service_test.go:TestCallBackService_Rank_EmptyItems).
//
// The handler channel is never reachable from this test (it would require
// initializing the global handler), so the assertion here is "no panic
// and clean return". A passing test means SendDirect returned before the
// channel send on web/callback_controller_handler.go:113.
func TestSendDirect_EmptyItemList(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SendDirect with empty ItemList panicked: %v", r)
		}
	}()

	SendDirect(&CallBackParam{
		SceneId:   "test_scene",
		RequestId: "test-req-id",
		ItemList:  nil, // the regression: zero candidates after upstream filters
	})
}

// TestRunCallbackSafely_RecoversFromInjectedPanic verifies the worker pool
// recover guard captures a deterministic panic raised from inside
// doCallbackLog. We inject a CallBackProcessFunc that always panics and
// route it via SceneId; the doCallbackLog flow reaches the
// callBackProcessFuncMap dispatch on web/callback_controller.go:159 before
// touching any other potentially-panicking site, so the panic is provably
// triggered by our injection (not by an unrelated nil deref).
//
// Without runCallbackSafely, this panic would terminate the worker goroutine
// and crash the process. The test passes iff runCallbackSafely returns
// cleanly.
func TestRunCallbackSafely_RecoversFromInjectedPanic(t *testing.T) {
	const sceneID = "test_panic_inject_scene"

	// Inject a CallBackProcessFunc that always panics. Clean up after the
	// test so we do not leak state into other test cases in this package.
	RegisterCallBackProcessFunc(sceneID, func(user *module.User, items []*module.Item, ctx *context.RecommendContext) {
		panic("intentional panic from injected CallBackProcessFunc")
	})
	t.Cleanup(func() {
		delete(callBackProcessFuncMap, sceneID)
	})

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("runCallbackSafely must swallow injected panic but propagated: %v", r)
		}
	}()

	c := &CallBackController{}
	c.RequestId = "test-recover-injected"
	c.param = CallBackParam{
		SceneId:   sceneID,
		RequestId: "test-recover-injected",
		Uid:       "u1",
		ItemList: []map[string]any{
			{"item_id": "i1"},
		},
	}

	runCallbackSafely(c)
}

// TestRunCallbackSafely_NilController verifies the helper tolerates a nil
// controller without itself panicking before the recover frame is set up.
// This is defensive: callers should never pass nil, but the recover guard
// must not assume non-nil.
func TestRunCallbackSafely_NilController(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("runCallbackSafely(nil) must not panic but did: %v", r)
		}
	}()

	runCallbackSafely(nil)
}
