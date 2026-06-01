package service

import (
	"testing"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

// stubParam implements context.IParam for tests, returning a fixed scene name.
type stubParam struct {
	scene string
}

func (s *stubParam) GetParameter(name string) interface{} {
	if name == "scene" {
		return s.scene
	}
	return nil
}

// TestCallBackService_Rank_EmptyItems verifies that Rank() does not panic when
// the candidate item list is empty. The HTTP self-call entry used to be
// guarded by CallBackController.CheckParameter ("recommend item list not
// empty"), but the SendDirect entry bypasses that check. Without an explicit
// nil guard inside Rank(), algoGenerator.HasFeatures() returns false,
// algoData stays nil, and the goroutine loop dereferences it ->
// "invalid memory address or nil pointer dereference" at
// algoData.GetFeatures().
func TestCallBackService_Rank_EmptyItems(t *testing.T) {
	const scene = "test_scene"

	ctx := context.NewRecommendContext()
	ctx.Param = &stubParam{scene: scene}
	ctx.RecommendId = "test-req-id"
	ctx.Config = &recconf.RecommendConfig{
		CallBackConfs: map[string]recconf.CallBackConfig{
			scene: {
				RankConf: recconf.RankConfig{
					// Non-empty so we skip the early "no algo" return at
					// the top of Rank() and reach the algoData branch.
					RankAlgoList: []string{"dummy_algo"},
				},
			},
		},
	}

	svc := NewCallBackService()
	svc.User = module.NewUser("test_user")
	svc.Items = nil // the regression: zero candidates

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Rank() panicked with empty items: %v", r)
		}
	}()

	svc.Rank(ctx)
}
