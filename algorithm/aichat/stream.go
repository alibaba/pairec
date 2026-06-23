package aichat

import (
	"bufio"
	"encoding/json"
	"io"
	"sort"
	"strings"
)

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content   string `json:"content"`
			ToolCalls []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

func parseStream(reader io.Reader, onDelta DeltaHandler) (*StreamResult, error) {
	result := &StreamResult{}
	type acc struct {
		id        string
		callType  string
		name      string
		arguments string
	}
	toolCalls := map[int]*acc{}
	scanner := bufio.NewScanner(reader)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}
		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return nil, err
		}
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				result.Content += choice.Delta.Content
				if onDelta != nil {
					if err := onDelta(choice.Delta.Content); err != nil {
						return nil, err
					}
				}
			}
			if choice.FinishReason != "" {
				result.FinishReason = choice.FinishReason
			}
			for _, tc := range choice.Delta.ToolCalls {
				slot := toolCalls[tc.Index]
				if slot == nil {
					slot = &acc{}
					toolCalls[tc.Index] = slot
				}
				if tc.ID != "" {
					slot.id = tc.ID
				}
				if tc.Type != "" {
					slot.callType = tc.Type
				}
				slot.name += tc.Function.Name
				slot.arguments += tc.Function.Arguments
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	indexes := make([]int, 0, len(toolCalls))
	for index := range toolCalls {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)
	for _, index := range indexes {
		call := toolCalls[index]
		if call.callType == "" {
			call.callType = "function"
		}
		result.ToolCalls = append(result.ToolCalls, ToolCall{
			ID:   call.id,
			Type: call.callType,
			Function: FunctionCall{
				Name:      call.name,
				Arguments: call.arguments,
			},
		})
	}
	return result, nil
}
