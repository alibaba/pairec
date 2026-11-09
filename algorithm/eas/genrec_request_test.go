package eas

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
	proto "github.com/golang/protobuf/proto"
)

func TestNormalizeGenRecUrl(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"http://host/api/predict/svc", "http://host/api/predict/svc/generate"},
		{"http://host/api/predict/svc/", "http://host/api/predict/svc/generate"},
		{"http://host/api/predict/svc/generate", "http://host/api/predict/svc/generate"},
		{"http://host/api/predict/svc/generate/", "http://host/api/predict/svc/generate"},
	}
	for _, c := range cases {
		if got := normalizeGenRecUrl(c.input); got != c.expected {
			t.Errorf("normalizeGenRecUrl(%q) = %q, expected %q", c.input, got, c.expected)
		}
	}
}

func TestGenrecRequestInvoke(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/generate" {
			t.Errorf("unexpected path: %s", req.URL.Path)
		}
		if req.Header.Get("Authorization") != "test-token" {
			t.Errorf("unexpected Authorization: %s", req.Header.Get("Authorization"))
		}
		if req.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected Content-Type: %s", req.Header.Get("Content-Type"))
		}

		var body genRecHttpRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if len(body.InputIds) != 1 || body.InputIds[0] != 0 {
			t.Errorf("unexpected input_ids: %v", body.InputIds)
		}
		if body.Stream {
			t.Errorf("stream should be false")
		}
		if body.SamplingParams.BeamWidth != 100 || body.SamplingParams.N != 50 || body.SamplingParams.MaxNewTokens != 3 {
			t.Errorf("unexpected sampling params: %+v", body.SamplingParams)
		}
		if body.Rid != "req-1" {
			t.Errorf("unexpected rid: %s", body.Rid)
		}

		pbData, err := base64.StdEncoding.DecodeString(body.ImageData)
		if err != nil {
			t.Errorf("decode image_data: %v", err)
		}
		pbRequest := &easyrec.PBRequest{}
		if err := proto.Unmarshal(pbData, pbRequest); err != nil {
			t.Errorf("unmarshal PBRequest: %v", err)
		}
		if pbRequest.UserFeatures["hist"].GetStringFeature() != "youku:1;youku:2" {
			t.Errorf("unexpected hist: %q", pbRequest.UserFeatures["hist"].GetStringFeature())
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"output_ids":[1,2,3],"meta_info":{"beam_results":[{"output_ids":[1,2,3],"meta_info":{"sequence_score":-0.5},"item_ids":["100"]}]}}`))
	}))
	defer server.Close()

	req := GenrecRequest{}
	req.SetUrl(server.URL)
	req.SetAuth("test-token")
	req.SetTimeout(1000)

	response, err := req.Invoke(&GenRecRequestData{
		Request: &easyrec.PBRequest{
			UserFeatures: map[string]*easyrec.PBFeature{
				"hist": {Value: &easyrec.PBFeature_StringFeature{StringFeature: "youku:1;youku:2"}},
			},
		},
		BeamWidth:    100,
		SidCount:     50,
		MaxNewTokens: 3,
		Rid:          "req-1",
	})
	if err != nil {
		t.Fatalf("invoke: %v", err)
	}

	genRecResponse, ok := response.(*GenRecResponse)
	if !ok {
		t.Fatalf("unexpected response type %T", response)
	}
	if len(genRecResponse.MetaInfo.BeamResults) != 1 {
		t.Fatalf("unexpected beam count: %d", len(genRecResponse.MetaInfo.BeamResults))
	}
	beam := genRecResponse.MetaInfo.BeamResults[0]
	if len(beam.ItemIds) != 1 || beam.ItemIds[0] != "100" {
		t.Errorf("unexpected item ids: %v", beam.ItemIds)
	}
	if beam.MetaInfo.SequenceScore != -0.5 {
		t.Errorf("unexpected score: %v", beam.MetaInfo.SequenceScore)
	}
}

func TestGenrecRequestInvokeHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	req := GenrecRequest{}
	req.SetUrl(server.URL)
	req.SetTimeout(1000)

	if _, err := req.Invoke(&GenRecRequestData{Request: &easyrec.PBRequest{}}); err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestGenRecResponseSample(t *testing.T) {
	data, err := os.ReadFile("../../docs/GenRec/response.txt")
	if err != nil {
		t.Fatalf("read sample response: %v", err)
	}
	response := &GenRecResponse{}
	if err := json.Unmarshal(data, response); err != nil {
		t.Fatalf("unmarshal sample response: %v", err)
	}
	if len(response.MetaInfo.BeamResults) != 50 {
		t.Fatalf("expected 50 beams, got %d", len(response.MetaInfo.BeamResults))
	}
	first := response.MetaInfo.BeamResults[0]
	if len(first.ItemIds) != 1 || first.ItemIds[0] != "11012379717" {
		t.Errorf("unexpected first item ids: %v", first.ItemIds)
	}
	if first.MetaInfo.SequenceScore != -0.1423762639363607 {
		t.Errorf("unexpected first score: %v", first.MetaInfo.SequenceScore)
	}
	for i, beam := range response.MetaInfo.BeamResults {
		if len(beam.ItemIds) == 0 {
			t.Errorf("beam %d has no item ids", i)
		}
		if i > 0 && beam.MetaInfo.SequenceScore > response.MetaInfo.BeamResults[i-1].MetaInfo.SequenceScore {
			t.Errorf("beam %d score %v is greater than previous beam", i, beam.MetaInfo.SequenceScore)
		}
	}
}
