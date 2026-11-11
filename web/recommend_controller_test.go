package web

import (
	"encoding/json"
	"testing"
)

func TestRecommendControllerUsesCallerRequestId(t *testing.T) {
	controller := &RecommendController{
		Controller: Controller{RequestId: "generated-request-id"},
	}
	controller.RequestBody = []byte(`{"uid":"user-1","request_id":"caller-request-id"}`)
	if err := controller.CheckParameter(); err != nil {
		t.Fatal(err)
	}
	if controller.RequestId != "caller-request-id" {
		t.Fatalf("RequestId = %q, want caller-request-id", controller.RequestId)
	}
}

func TestRecommendControllerKeepsGeneratedRequestId(t *testing.T) {
	controller := &RecommendController{
		Controller: Controller{RequestId: "generated-request-id"},
	}
	controller.RequestBody = []byte(`{"uid":"user-1"}`)
	if err := controller.CheckParameter(); err != nil {
		t.Fatal(err)
	}
	if controller.RequestId != "generated-request-id" {
		t.Fatalf("RequestId = %q, want generated-request-id", controller.RequestId)
	}
}

func TestItemDataIncludesScore(t *testing.T) {
	contents, err := json.Marshal(ItemData{ItemId: "item-1", Score: 0.75})
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]interface{}
	if err := json.Unmarshal(contents, &value); err != nil {
		t.Fatal(err)
	}
	if value["score"] != 0.75 {
		t.Fatalf("score = %v, want 0.75", value["score"])
	}
}
