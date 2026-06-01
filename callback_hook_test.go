package pairec

import (
	"testing"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

// hookStubParam implements context.IParam for tests, returning a fixed scene.
type hookStubParam struct {
	scene string
}

func (s *hookStubParam) GetParameter(name string) interface{} {
	if name == "scene" {
		return s.scene
	}
	return nil
}

// TestCallBackHookFunc_EmptyItems_EarlyReturn verifies the hook returns
// cleanly when the recommend pipeline produced zero candidate items, instead
// of constructing a CallBackParam with an empty ItemList and forwarding it to
// the worker via web.SendDirect. The downstream worker (CallBackService.Rank)
// previously panicked on a nil algoData when given an empty item list, see
// service/call_back_service_test.go:TestCallBackService_Rank_EmptyItems.
//
// To prove the early return fires before any further work, this test passes
// a nil *module.User as params[0]. The hook accesses user.MakeUserFeatures
// only after the items length check; if the guard were missing, the nil user
// would be dereferenced and the test would panic.
func TestCallBackHookFunc_EmptyItems_EarlyReturn(t *testing.T) {
	const scene = "test_scene"

	ctx := context.NewRecommendContext()
	ctx.Param = &hookStubParam{scene: scene}
	ctx.RecommendId = "test-req-id"
	ctx.Debug = true // bypass the !Debug && !callbackFlag short-circuit
	ctx.Config = &recconf.RecommendConfig{
		CallBackConfs: map[string]recconf.CallBackConfig{
			scene: {},
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("CallBackHookFunc panicked with empty items: %v", r)
		}
	}()

	var nilUser *module.User // intentional: proves we return before user access
	emptyItems := []*module.Item{}

	CallBackHookFunc(ctx, nilUser, emptyItems)
}
