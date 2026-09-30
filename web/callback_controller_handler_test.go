package web

import (
	"fmt"
	"sync"
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
		unregisterCallBackProcessFunc(sceneID)
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

// TestCallBackProcessFuncMap_ConcurrentAccess verifies that concurrent
// Register / lookup / unregister against callBackProcessFuncMap is race-free.
// Production code reads this map from worker goroutines while user code may
// still be registering new scenes at startup, so the synchronization in
// callback_controller_func.go is load-bearing. Run with `go test -race` to
// detect any regression that drops the mutex.
func TestCallBackProcessFuncMap_ConcurrentAccess(t *testing.T) {
	const workers = 32
	const opsPerWorker = 200

	noop := func(user *module.User, items []*module.Item, ctx *context.RecommendContext) {}

	// Make sure none of the scenes used here survive the test.
	t.Cleanup(func() {
		for i := 0; i < workers; i++ {
			unregisterCallBackProcessFunc(fmt.Sprintf("race_scene_%d", i))
		}
	})

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			scene := fmt.Sprintf("race_scene_%d", id)
			for j := 0; j < opsPerWorker; j++ {
				RegisterCallBackProcessFunc(scene, noop)
				if _, ok := lookupCallBackProcessFunc(scene); !ok {
					t.Errorf("scene %q: lookup miss right after register", scene)
					return
				}
				unregisterCallBackProcessFunc(scene)
			}
		}(i)
	}
	wg.Wait()
}
