package recallenginerecall

import (
	"errors"
	"reflect"
	"testing"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/recallengine"
	"github.com/alibaba/pairec/v2/module"
)

type multiEmbeddingTestTrigger struct{ result *TriggerResult }

func (t multiEmbeddingTestTrigger) GetTriggerKey(*module.User, *context.RecommendContext) *TriggerResult {
	return t.result
}
func TestMultiEmbeddingBuildQueryParams(t *testing.T) {
	queries := [][]float32{{1, 2}, {3, 4}, {5, 6}}
	r := &RecallEngineVectorRecall{returnCount: 10, versionId: "fallback", timeout: 50, triggerKey: multiEmbeddingTestTrigger{&TriggerResult{Queries: queries, Version: "model"}}}
	got, err := r.BuildQueryParams(nil, nil)
	if err != nil || got.Trigger != "" || got.VersionId != "model" || got.Count != 10 || !reflect.DeepEqual(got.Queries, queries) || got.Options.Timeout != 50 {
		t.Fatalf("got %+v, err %v", got, err)
	}
	r.triggerKey = multiEmbeddingTestTrigger{&TriggerResult{Queries: queries}}
	got, err = r.BuildQueryParams(nil, nil)
	if err != nil || got.VersionId != "fallback" {
		t.Fatalf("static version: %+v %v", got, err)
	}
	r.versionId = ""
	if _, err = r.BuildQueryParams(nil, nil); err == nil {
		t.Fatal("missing version accepted")
	}
	r.triggerKey = multiEmbeddingTestTrigger{&TriggerResult{Queries: [][]float32{}}}
	got, err = r.BuildQueryParams(nil, nil)
	if err != nil || got.Count != 0 {
		t.Fatal("empty queries not skipped")
	}
	expected := errors.New("inference failed")
	r.triggerKey = multiEmbeddingTestTrigger{&TriggerResult{Err: expected}}
	if _, err = r.BuildQueryParams(nil, nil); !errors.Is(err, expected) {
		t.Fatalf("lost error: %v", err)
	}
	r.triggerKey = multiEmbeddingTestTrigger{&TriggerResult{TriggerItem: "1,2"}}
	got, err = r.BuildQueryParams(nil, nil)
	if err != nil || got.Queries != nil || got.Trigger != "1,2" || got.Count != 10 {
		t.Fatalf("legacy changed: %+v %v", got, err)
	}
}
func TestMultiEmbeddingServiceBuildErrorAndEmpty(t *testing.T) {
	client := recallengine.NewRecallEngineClient("local", "local", "http://127.0.0.1:1", "")
	expected := errors.New("invalid tensor")
	vector := &RecallEngineVectorRecall{recallName: "vectors", returnCount: 10, triggerKey: multiEmbeddingTestTrigger{&TriggerResult{Err: expected}}}
	service := &RecallEngineServiceRecall{client: client, recallMap: map[string]RecallEngineBaseRecall{"vectors": vector}}
	user := module.NewUser("test-user")
	ctx := context.NewRecommendContext()
	if _, err := service.GetItems(user, ctx); !errors.Is(err, expected) {
		t.Fatalf("lost build error: %v", err)
	}
	vector.triggerKey = multiEmbeddingTestTrigger{&TriggerResult{Queries: [][]float32{}}}
	if items, err := service.GetItems(user, ctx); err != nil || len(items) != 0 {
		t.Fatalf("empty batch made a network call: %v %v", items, err)
	}
}

func TestMultiEmbeddingCloneKeepsVersionAndRealtimeTrigger(t *testing.T) {
	original := &RecallEngineVectorRecall{versionId: "snapshot", recallName: "vectors", cloneInstances: make(map[string]*RecallEngineVectorRecall)}
	params := map[string]interface{}{"Count": 20, "Timeout": 50, "TriggerType": "user_realtime_embedding", "UserRealtimeEmbeddingTrigger": map[string]interface{}{"RecallAlgo": "embedding-model"}}
	clone := original.CloneWithConfig(params).(*RecallEngineVectorRecall)
	trigger, ok := clone.triggerKey.(*UserRealtimeEmbeddingTrigger)
	if !ok || trigger.recallAlgo != "embedding-model" || clone.versionId != "snapshot" || clone.returnCount != 20 || clone.timeout != 50 {
		t.Fatal("clone lost multi-embedding configuration")
	}
	if original.CloneWithConfig(params) != clone {
		t.Fatal("existing clone cache was bypassed")
	}
}
