package aishopping

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alibaba/pairec/v2/service/searchsuggestion"
)

func TestAssistantMemoryRoundTripAndContext(t *testing.T) {
	intent := &searchsuggestion.SearchIntent{Keywords: []string{"blue"}}
	blob := &SessionBlob{}
	blob.recordTurn("蓝色长袖衬衫", "改成粉色，保留长袖，对吗？", intent)
	blob.recordTurn("对", "已保留长袖。", nil)
	payload, err := json.Marshal(blob)
	if err != nil {
		t.Fatal(err)
	}
	var loaded SessionBlob
	if err := json.Unmarshal(payload, &loaded); err != nil {
		t.Fatal(err)
	}
	messages := loaded.messages("还有别的吗？")
	if len(messages) != 3 || messages[2].Content != "还有别的吗？" {
		t.Fatalf("unexpected message contract: %+v", messages)
	}
	var context SessionBlob
	if err := json.Unmarshal([]byte(strings.TrimPrefix(messages[1].Content, "Previous session context (JSON):\n")), &context); err != nil {
		t.Fatal(err)
	}
	if len(context.UserQueries) != 2 || context.UserQueries[0].AssistantReply != "改成粉色，保留长袖，对吗？" || context.UserQueries[1].AssistantReply != "已保留长袖。" {
		t.Fatalf("lost paired replies: %+v", context.UserQueries)
	}
	if loaded.LastSearchTurnID != 1 || loaded.LastSearch.Keywords[0] != "blue" {
		t.Fatal("clarification changed the search snapshot")
	}
}

func TestAssistantMemoryLegacySession(t *testing.T) {
	var blob SessionBlob
	if err := json.Unmarshal([]byte(`{"turn_count":1,"user_queries":[{"turn_id":1,"query":"衬衫"}],"last_search":{"keywords":["blue"]},"last_search_turn_id":1}`), &blob); err != nil {
		t.Fatal(err)
	}
	if blob.UserQueries[0].AssistantReply != "" {
		t.Fatal("invented legacy reply")
	}
	payload, _ := json.Marshal(blob)
	if strings.Contains(string(payload), "assistant_reply") {
		t.Fatal("empty legacy field should be omitted")
	}
	if len(blob.messages("粉色")) != 3 {
		t.Fatal("legacy messages changed")
	}
}

func TestAssistantMemoryTrimCountsRepliesAndKeepsPairs(t *testing.T) {
	blob := &SessionBlob{}
	for _, query := range []string{"first", "second", "third"} {
		blob.recordTurn(query, strings.Repeat(query, 200), nil)
	}
	blob.trim(32, 250)
	if len(blob.UserQueries) != 1 || blob.UserQueries[0].Query != "third" || blob.UserQueries[0].AssistantReply != strings.Repeat("third", 200) {
		t.Fatalf("reply budget or pairing lost: %+v", blob.UserQueries)
	}
}

func TestAssistantMemoryDoesNotChangeKnowledgeQuery(t *testing.T) {
	blob := &SessionBlob{LastSearch: &searchsuggestion.SearchIntent{Keywords: []string{"blue"}}, LastSearchTurnID: 1, UserQueries: []SessionQuery{{TurnID: 1, Query: "蓝色长袖衬衫"}, {TurnID: 2, Query: "粉色"}}}
	before := blob.knowledgeQuery("party", blob.LastSearch)
	blob.UserQueries[0].AssistantReply = "派对裙、爱心、蝴蝶结"
	blob.UserQueries[1].AssistantReply = "改短袖，对吗？"
	if after := blob.knowledgeQuery("party", blob.LastSearch); after != before {
		t.Fatalf("knowledge input changed: %q -> %q", before, after)
	}
}

func TestAssistantMemoryCapturesOnlyEmittedText(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewStreamWriter("request", "session", recorder, recorder)
	if err := writer.EmitContent("请问"); err != nil {
		t.Fatal(err)
	}
	if err := writer.EmitStep("reply"); err != nil {
		t.Fatal(err)
	}
	if err := writer.EmitCitation("item"); err != nil {
		t.Fatal(err)
	}
	if err := writer.EmitContent("图案不限吗？"); err != nil {
		t.Fatal(err)
	}
	if err := writer.EmitSuggestions([]string{"衬衫"}); err != nil {
		t.Fatal(err)
	}
	if writer.Content() != "请问图案不限吗？" {
		t.Fatalf("non-text leaked: %q", writer.Content())
	}
	failed := NewStreamWriter("request", "session", failingMemoryWriter{}, nil)
	if err := failed.EmitContent("not delivered"); err == nil {
		t.Fatal("expected write error")
	}
	if failed.Content() != "" {
		t.Fatal("failed write captured")
	}
}

type failingMemoryWriter struct{}

func (failingMemoryWriter) Header() http.Header       { return make(http.Header) }
func (failingMemoryWriter) WriteHeader(int)           {}
func (failingMemoryWriter) Write([]byte) (int, error) { return 0, errors.New("disconnected") }
