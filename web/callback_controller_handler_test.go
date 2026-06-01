package web

import (
	"testing"
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
