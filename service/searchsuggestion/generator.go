package searchsuggestion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/alibaba/pairec/v2/algorithm/aichat"
)

type toolArguments struct {
	Suggestions []string `json:"suggestions"`
}

func Generate(ctx context.Context, runtimeConfig *RuntimeConfig, input *GenerationInput) Outcome {
	if runtimeConfig == nil || runtimeConfig.Model == nil {
		return Outcome{Err: NewError(CodeModelUnavailable, false, fmt.Errorf("model is unavailable"))}
	}
	payload, inputErr := input.JSON()
	if inputErr != nil {
		return Outcome{Err: inputErr}
	}
	temperature := 0.2
	parallelToolCalls := false
	request := &aichat.ChatCompletionRequest{
		Messages: []aichat.Message{
			{Role: "system", Content: runtimeConfig.Prompt},
			{Role: "user", Content: "UNTRUSTED_SUGGESTION_CONTEXT_JSON:\n" + string(payload)},
		},
		Tools: []aichat.Tool{aichat.SuggestionTool(input.SuggestionCount)},
		ToolChoice: map[string]interface{}{
			"type":     "function",
			"function": map[string]string{"name": "emit_suggestions"},
		},
		Stream:            true,
		EnableThinking:    false,
		Temperature:       &temperature,
		ParallelToolCalls: &parallelToolCalls,
		MaxTokens:         384,
	}
	result, err := runtimeConfig.Model.Stream(ctx, request, nil)
	if err != nil {
		if ctx.Err() != nil {
			return Outcome{Err: NewError(CodeTimeout, true, ctx.Err())}
		}
		return Outcome{Err: NewError(CodeModelError, true, err)}
	}
	suggestions, parseErr := parseResult(result, input.SuggestionCount)
	if parseErr != nil {
		return Outcome{Err: parseErr}
	}
	validated, validationErr := Validate(suggestions, input.SuggestionCount, input.CurrentQuery)
	if validationErr != nil {
		return Outcome{Err: validationErr}
	}
	return Outcome{Suggestions: validated}
}

func parseResult(result *aichat.StreamResult, count int) ([]string, *Error) {
	if result == nil || strings.TrimSpace(result.Content) != "" || len(result.ToolCalls) != 1 {
		return nil, NewError(CodeInvalidStructure, true, fmt.Errorf("exactly one tool call and no prose are required"))
	}
	toolCall := result.ToolCalls[0]
	if strings.TrimSpace(toolCall.ID) == "" || toolCall.Type != "function" || toolCall.Function.Name != "emit_suggestions" {
		return nil, NewError(CodeInvalidStructure, true, fmt.Errorf("invalid emit_suggestions tool call"))
	}
	decoder := json.NewDecoder(bytes.NewBufferString(toolCall.Function.Arguments))
	decoder.DisallowUnknownFields()
	var arguments toolArguments
	if err := decoder.Decode(&arguments); err != nil {
		return nil, NewError(CodeInvalidStructure, true, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("multiple JSON values")
		}
		return nil, NewError(CodeInvalidStructure, true, err)
	}
	if len(arguments.Suggestions) != count {
		return nil, NewError(CodeInvalidStructure, true, fmt.Errorf("expected %d suggestions", count))
	}
	return arguments.Suggestions, nil
}
