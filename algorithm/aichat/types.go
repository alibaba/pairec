package aichat

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallId string     `json:"tool_call_id,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id,omitempty"`
	Type     string       `json:"type,omitempty"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type ChatCompletionRequest struct {
	Model             string      `json:"model"`
	Messages          []Message   `json:"messages"`
	Tools             []Tool      `json:"tools,omitempty"`
	ToolChoice        interface{} `json:"tool_choice,omitempty"`
	Stream            bool        `json:"stream"`
	EnableThinking    bool        `json:"enable_thinking"`
	Temperature       *float64    `json:"temperature,omitempty"`
	ParallelToolCalls *bool       `json:"parallel_tool_calls,omitempty"`
	MaxTokens         int         `json:"max_tokens,omitempty"`
}

type StreamResult struct {
	Content      string
	ToolCalls    []ToolCall
	FinishReason string
}

type DeltaHandler func(text string) error
