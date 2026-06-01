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
