package rank

import (
	"testing"
)

func metaOf(t *testing.T, data IAlgoData) map[string]string {
	t.Helper()
	d, ok := data.(*EasyrecAlgoData)
	if !ok {
		t.Fatalf("expected EasyrecAlgoData, got %T", data)
	}
	return d.easyrecRequest.MetaData
}

func TestEasyrecGeneratorAlwaysSetsUserIdMetaData(t *testing.T) {
	g := NewEasyrecAlgoDataGenerator(nil)
	g.SetUserId("188569")
	g.AddFeatures(nil, nil, map[string]interface{}{"uid": "188569"})

	meta := metaOf(t, g.GeneratorAlgoData())
	if meta["user_id"] != "188569" {
		t.Fatalf("non-debug meta_data missing user_id: %v", meta)
	}
}

func TestEasyrecGeneratorDebugKeepsRequestIdAndUserId(t *testing.T) {
	g := NewEasyrecAlgoDataGenerator(nil)
	g.SetUserId("188569")
	g.AddFeatures(nil, nil, map[string]interface{}{"uid": "188569"})

	meta := metaOf(t, g.GeneratorAlgoDataDebugWithLevel(100, map[string]string{"request_id": "req-1"}))
	if meta["user_id"] != "188569" || meta["request_id"] != "req-1" {
		t.Fatalf("debug meta_data must carry both user_id and request_id: %v", meta)
	}
}

func TestEasyrecGeneratorWithoutUserIdKeepsOldMetaData(t *testing.T) {
	g := NewEasyrecAlgoDataGenerator(nil)
	g.AddFeatures(nil, nil, map[string]interface{}{"uid": "188569"})

	if meta := metaOf(t, g.GeneratorAlgoData()); len(meta) != 0 {
		t.Fatalf("non-debug meta_data must stay empty without SetUserId: %v", meta)
	}
	if meta := metaOf(t, g.GeneratorAlgoDataDebugWithLevel(100, map[string]string{"request_id": "req-1"})); len(meta) != 1 || meta["request_id"] != "req-1" {
		t.Fatalf("debug meta_data must only carry request_id without SetUserId: %v", meta)
	}
}
