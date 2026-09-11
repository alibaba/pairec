package shoppingknowledge

import (
	"fmt"
	"unicode/utf8"

	"github.com/alibaba/pairec/v2/log"
)

// Split long payloads without omitting any model-visible fields.
const knowledgeLogChunkBytes = 4096

// LogModelView logs the same knowledge JSON sent to the model.
func (e *Evidence) LogModelView(requestID string) {
	if e == nil {
		return
	}
	payload := e.PromptJSON()
	var chunks []string
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
		log.Info(fmt.Sprintf("requestId=%s\tmodule=ShoppingKnowledge\tevent=model_view\tpartIndex=%d\tpartCount=%d\tpayload=%s",
			requestID, part+1, len(chunks), chunk))
	}
}
