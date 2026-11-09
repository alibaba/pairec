package recall

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"sort"
	"testing"

	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/algorithm/eas"
	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	proto "github.com/golang/protobuf/proto"
)

func TestGenRecEncodeValue(t *testing.T) {
	cases := []struct {
		name      string
		value     interface{}
		separator string
		expected  string
		ok        bool
	}{
		{"nil", nil, "\x1d", "", true},
		{"string", "abc", "\x1d", "abc", true},
		{"empty string", "", "\x1d", "", true},
		{"int", 12, "\x1d", "12", true},
		{"int64", int64(12), "\x1d", "12", true},
		{"float64 integral", float64(18), "\x1d", "18", true},
		{"float64", 1.5, "\x1d", "1.5", true},
		{"string slice", []string{"a", "b"}, "\x1d", "a\x1db", true},
		{"interface slice", []interface{}{float64(18), float64(20)}, "\x1d", "18\x1d20", true},
		{"single element", []interface{}{float64(1)}, "\x1d", "1", true},
		{"empty slice", []interface{}{}, "\x1d", "", true},
		{"duplicates preserved", []string{"a", "a", "b"}, ";", "a;a;b", true},
		{"bytes", []byte("abc"), "\x1d", "abc", true},
		{"bool unsupported", true, "\x1d", "", false},
		{"map unsupported", map[string]string{"a": "b"}, "\x1d", "", false},
		{"slice with unsupported element", []interface{}{true}, "\x1d", "", false},
	}
	for _, c := range cases {
		got, ok := genRecEncodeValue(c.value, c.separator)
		if ok != c.ok || got != c.expected {
			t.Errorf("%s: genRecEncodeValue(%v) = (%q, %v), expected (%q, %v)", c.name, c.value, got, ok, c.expected, c.ok)
		}
	}
}

func TestGenRecEncodeUserFeaturesWhitelist(t *testing.T) {
	recall := NewGenRecRecall(recconf.RecallConfig{
		GenRecConf: recconf.GenRecConfig{
			FeatureNames: []string{"hist", "st_middle_1001", "province", "lt_big_type", "missing_field"},
		},
	})
	features := map[string]interface{}{
		"hist":           "youku:1\x1dyouku:2\x1dyouku:2",
		"st_middle_1001": []interface{}{float64(18), float64(20)},
		"province":       float64(3),
		"lt_big_type":    []interface{}{},
		"other_feature":  "should not be sent",
	}
	encoded := recall.encodeUserFeatures(features, &context.RecommendContext{})

	if len(encoded) != 5 {
		t.Fatalf("expected 5 features, got %d", len(encoded))
	}
	if _, ok := encoded["other_feature"]; ok {
		t.Errorf("feature outside whitelist should not be encoded")
	}
	cases := map[string]string{
		"hist":           "youku:1;youku:2;youku:2",
		"st_middle_1001": "18\x1d20",
		"province":       "3",
		"lt_big_type":    "",
		"missing_field":  "",
	}
	for name, expected := range cases {
		feature, ok := encoded[name]
		if !ok {
			t.Errorf("feature %s not encoded", name)
			continue
		}
		if got := feature.GetStringFeature(); got != expected {
			t.Errorf("feature %s = %q, expected %q", name, got, expected)
		}
	}
}

func TestGenRecEncodeUserFeaturesAllProperties(t *testing.T) {
	recall := NewGenRecRecall(recconf.RecallConfig{})
	features := map[string]interface{}{
		"hist":     []interface{}{"youku:1", "youku:2", "youku:2"},
		"province": float64(3),
		"complex":  map[string]string{"a": "b"},
	}
	encoded := recall.encodeUserFeatures(features, &context.RecommendContext{})

	if len(encoded) != 2 {
		t.Fatalf("expected 2 features, got %d", len(encoded))
	}
	if got := encoded["hist"].GetStringFeature(); got != "youku:1;youku:2;youku:2" {
		t.Errorf("hist = %q", got)
	}
	if got := encoded["province"].GetStringFeature(); got != "3" {
		t.Errorf("province = %q", got)
	}
	if _, ok := encoded["complex"]; ok {
		t.Errorf("unsupported type should be skipped")
	}
}

// TestGenRecSampleRequestReplay encodes the saved user features and compares
// every field against the PBRequest captured in the real request sample.
func TestGenRecSampleRequestReplay(t *testing.T) {
	userFeaturesData, err := os.ReadFile("../../docs/GenRec/user_features.txt")
	if err != nil {
		t.Fatalf("read user features: %v", err)
	}
	var rawFeatures map[string][]interface{}
	if err := json.Unmarshal(userFeaturesData, &rawFeatures); err != nil {
		t.Fatalf("unmarshal user features: %v", err)
	}
	features := make(map[string]interface{}, len(rawFeatures))
	for name, value := range rawFeatures {
		features[name] = value
	}

	requestData, err := os.ReadFile("../../docs/GenRec/request.txt")
	if err != nil {
		t.Fatalf("read request sample: %v", err)
	}
	var httpRequest struct {
		ImageData string `json:"image_data"`
	}
	if err := json.Unmarshal(requestData, &httpRequest); err != nil {
		t.Fatalf("unmarshal request sample: %v", err)
	}
	pbData, err := base64.StdEncoding.DecodeString(httpRequest.ImageData)
	if err != nil {
		t.Fatalf("decode image_data: %v", err)
	}
	expected := &easyrec.PBRequest{}
	if err := proto.Unmarshal(pbData, expected); err != nil {
		t.Fatalf("unmarshal expected PBRequest: %v", err)
	}
	if len(expected.UserFeatures) != 37 {
		t.Fatalf("expected 37 fields in sample, got %d", len(expected.UserFeatures))
	}

	names := make([]string, 0, len(expected.UserFeatures))
	for name := range expected.UserFeatures {
		names = append(names, name)
	}
	sort.Strings(names)

	recall := NewGenRecRecall(recconf.RecallConfig{
		GenRecConf: recconf.GenRecConfig{FeatureNames: names},
	})
	actual := recall.encodeUserFeatures(features, &context.RecommendContext{})
	if len(actual) != len(expected.UserFeatures) {
		t.Fatalf("expected %d encoded fields, got %d", len(expected.UserFeatures), len(actual))
	}
	for name, expectedFeature := range expected.UserFeatures {
		actualFeature, ok := actual[name]
		if !ok {
			t.Errorf("field %s missing in encoded features", name)
			continue
		}
		expectedValue := expectedFeature.GetStringFeature()
		if got := actualFeature.GetStringFeature(); got != expectedValue {
			t.Errorf("field %s = %q, expected %q", name, got, expectedValue)
		}
	}
}

type genRecFakeAlgo struct {
	requestData *eas.GenRecRequestData
	response    *eas.GenRecResponse
}

func (a *genRecFakeAlgo) Init(conf *recconf.AlgoConfig) error { return nil }
func (a *genRecFakeAlgo) Run(algoData interface{}) (interface{}, error) {
	a.requestData = algoData.(*eas.GenRecRequestData)
	return a.response, nil
}

func TestGenRecGetCandidateItems(t *testing.T) {
	newResponse := func() *eas.GenRecResponse {
		response := &eas.GenRecResponse{}
		response.MetaInfo.BeamResults = []eas.GenRecBeamResult{
			{ItemIds: []string{"a", "b"}},
			{ItemIds: []string{"b", "c"}},
			{ItemIds: []string{}},
			{ItemIds: []string{"", "d"}},
		}
		scores := []float64{-0.1, -0.2, -0.3, -0.4}
		for i := range response.MetaInfo.BeamResults {
			response.MetaInfo.BeamResults[i].MetaInfo.SequenceScore = scores[i]
		}
		return response
	}

	fake := &genRecFakeAlgo{response: newResponse()}
	algorithm.RegisterAlgorithm("test_genrec_algo", fake)

	recall := NewGenRecRecall(recconf.RecallConfig{
		Name:        "genrec_test",
		RecallAlgo:  "test_genrec_algo",
		RecallCount: 3,
		ItemType:    "video",
		GenRecConf: recconf.GenRecConfig{
			BeamWidth:    100,
			SidCount:     50,
			MaxNewTokens: 3,
		},
	})

	ctx := &context.RecommendContext{RecommendId: "req-1"}
	items := recall.GetCandidateItems(module.NewUser("u1"), ctx)

	// a, b from the first beam, c from the second (b deduplicated); the empty
	// beam is skipped, then truncated to 3 before d.
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	expectedIds := []string{"a", "b", "c"}
	expectedScores := []float64{-0.1, -0.1, -0.2}
	for i, item := range items {
		if string(item.Id) != expectedIds[i] {
			t.Errorf("item %d id = %s, expected %s", i, item.Id, expectedIds[i])
		}
		if item.Score != expectedScores[i] {
			t.Errorf("item %d score = %v, expected %v", i, item.Score, expectedScores[i])
		}
		if item.RetrieveId != "genrec_test" || item.ItemType != "video" {
			t.Errorf("item %d retrieveId=%s itemType=%s", i, item.RetrieveId, item.ItemType)
		}
	}

	if fake.requestData == nil {
		t.Fatal("algorithm not invoked")
	}
	if fake.requestData.BeamWidth != 100 || fake.requestData.SidCount != 50 || fake.requestData.MaxNewTokens != 3 {
		t.Errorf("unexpected sampling params: %+v", fake.requestData)
	}
	if fake.requestData.Rid != "req-1" {
		t.Errorf("unexpected rid: %s", fake.requestData.Rid)
	}

	// Without truncation all four distinct items are returned in beam order.
	recall = NewGenRecRecall(recconf.RecallConfig{
		Name:       "genrec_test",
		RecallAlgo: "test_genrec_algo",
	})
	items = recall.GetCandidateItems(module.NewUser("u1"), ctx)
	if len(items) != 4 {
		t.Fatalf("expected 4 items without truncation, got %d", len(items))
	}
	if string(items[3].Id) != "d" || items[3].Score != -0.4 {
		t.Errorf("last item = %s/%v, expected d/-0.4", items[3].Id, items[3].Score)
	}
}
