package shoppingknowledge

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/alibaba/pairec/v2/log"
	recallsvc "github.com/alibaba/pairec/v2/service/recall"
)

// Splitting long records keeps each payload manageable without omitting any data.
const knowledgeLogChunkBytes = 4096

// LogSearchResult includes records rejected by hit validation and fields not
// selected by ModelFields. Call it before handling a search result's error.
func LogSearchResult(requestID string, result *recallsvc.KnowledgeSearchResult) {
	if result == nil {
		return
	}
	for index, item := range result.RawItems {
		logKnowledgeRecord(requestID, "raw", index+1, len(result.RawItems), item)
	}
}

// LogModelView logs exactly the configured, deduplicated model-visible records.
func (e *Evidence) LogModelView(requestID string) {
	if e == nil {
		return
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=ShoppingKnowledge\tevent=model_view\tcount=%d\tmodelBytes=%d",
		requestID, len(e.values), len(e.promptJSON)))
	for index, value := range e.values {
		logKnowledgeRecord(requestID, "model", index+1, len(e.values), value)
	}
}

func logKnowledgeRecord(requestID, view string, index, count int, value interface{}) {
	payload, err := json.Marshal(value)
	if err != nil {
		log.Warning(fmt.Sprintf("requestId=%s\tmodule=ShoppingKnowledge\tevent=record_encode_error\tview=%s\trecordIndex=%d\terr=%v",
			requestID, view, index, err))
		return
	}
	var chunks [][]byte
	for len(payload) > 0 {
		end := min(len(payload), knowledgeLogChunkBytes)
		if end < len(payload) {
			for !utf8.RuneStart(payload[end]) {
				end--
			}
		}
		chunks = append(chunks, payload[:end])
		payload = payload[end:]
	}
	for part, chunk := range chunks {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=ShoppingKnowledge\tevent=record\tview=%s\trecordIndex=%d\trecordCount=%d\tpartIndex=%d\tpartCount=%d\tpayload=%s",
			requestID, view, index, count, part+1, len(chunks), chunk))
	}
}
