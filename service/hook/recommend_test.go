package hook

import (
	"errors"
	"sync"
	"testing"

	"github.com/alibaba/pairec/v2/context"
)

// TestSafeRun_RecoversFromPanic verifies that a panicking hook does not
// propagate out of SafeRun. Hooks are dispatched in their own goroutines on
// the recommend main path; without this guard a single faulty hook would
// crash the entire process.
func TestSafeRun_RecoversFromPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SafeRun must swallow downstream panic but propagated: %v", r)
		}
	}()

	hf := func(ctx *context.RecommendContext, params ...any) {
		panic(errors.New("boom"))
	}

	SafeRun(hf, context.NewRecommendContext())
}

// TestSafeRun_NilContext verifies SafeRun does not panic when the hook
// itself panics with a nil context. The recover handler reads ctx.RecommendId
// for the log line, so a nil ctx must be tolerated.
func TestSafeRun_NilContext(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SafeRun(nil ctx) must not panic but did: %v", r)
		}
	}()

	hf := func(ctx *context.RecommendContext, params ...any) {
		panic("nil ctx panic")
	}

	SafeRun(hf, nil)
}

// TestSafeRun_PassesParamsThrough verifies SafeRun forwards variadic params
// to the wrapped hook. Existing call sites pass (user, items); a regression
// here would silently break every clean hook.
func TestSafeRun_PassesParamsThrough(t *testing.T) {
	var got []any
	hf := func(ctx *context.RecommendContext, params ...any) {
		got = params
	}

	SafeRun(hf, context.NewRecommendContext(), "a", 42, true)

	if len(got) != 3 || got[0] != "a" || got[1] != 42 || got[2] != true {
		t.Fatalf("SafeRun did not forward params; got=%v", got)
	}
}

// TestSafeRun_NormalReturn verifies SafeRun runs the hook synchronously and
// returns cleanly when no panic occurs.
func TestSafeRun_NormalReturn(t *testing.T) {
	var ran bool
	hf := func(ctx *context.RecommendContext, params ...any) {
		ran = true
	}

	SafeRun(hf, context.NewRecommendContext())

	if !ran {
		t.Fatal("SafeRun did not invoke the hook")
	}
}

// TestSafeRun_GoroutineIsolation models the production call site: SafeRun is
// dispatched in its own goroutine. A panic must not terminate the parent;
// here we approximate that by waiting on the goroutine and asserting it
// finishes without dragging the process down.
func TestSafeRun_GoroutineIsolation(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		hf := func(ctx *context.RecommendContext, params ...any) {
			panic("downstream bug")
		}
		SafeRun(hf, context.NewRecommendContext())
	}()

	wg.Wait()
}
