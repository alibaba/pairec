package aishopping

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type StreamWriter struct {
	requestId string
	sessionId string
	writer    http.ResponseWriter
	flusher   http.Flusher
}

type StreamEvent struct {
	RequestId string                 `json:"request_id"`
	SessionId string                 `json:"session_id"`
	Result    map[string]interface{} `json:"result"`
}

func NewStreamWriter(requestId, sessionId string, writer http.ResponseWriter, flusher http.Flusher) *StreamWriter {
	return &StreamWriter{requestId: requestId, sessionId: sessionId, writer: writer, flusher: flusher}
}

func (w *StreamWriter) EmitStep(step string) error {
	return w.emit(map[string]interface{}{"step_info": map[string]string{"step": step}})
}

func (w *StreamWriter) EmitContent(content string) error {
	if content == "" {
		return nil
	}
	return w.emit(map[string]interface{}{"content": content})
}

func (w *StreamWriter) EmitCitation(itemId string) error {
	if itemId == "" {
		return nil
	}
	return w.emit(map[string]interface{}{"citation": map[string]string{"type": "item", "_id": itemId}})
}

func (w *StreamWriter) EmitSuggestions(suggestions []string) error {
	return w.emit(map[string]interface{}{
		"payload": map[string]interface{}{"suggestions": suggestions},
	})
}

func (w *StreamWriter) EmitSuggestionError(code, message string, retryable bool) error {
	return w.emit(map[string]interface{}{
		"payload": map[string]interface{}{
			"suggestions_error": map[string]interface{}{
				"code":      code,
				"message":   message,
				"retryable": retryable,
			},
		},
	})
}

func (w *StreamWriter) EmitStop(reason, errorCode string) error {
	result := map[string]interface{}{"stop_reason": reason}
	if errorCode != "" {
		result["error_code"] = errorCode
	}
	return w.emit(result)
}

func (w *StreamWriter) emit(result map[string]interface{}) error {
	event := StreamEvent{RequestId: w.requestId, SessionId: w.sessionId, Result: result}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w.writer, "data: %s\n\n", payload); err != nil {
		return err
	}
	if w.flusher != nil {
		w.flusher.Flush()
	}
	return nil
}
