package be

type Response struct {
	Result Result `json:"result"`
}

func NewResponse(result Result) *Response {
	return &Response{Result: result}
}

type MatchItem struct {
	FieldNames  []string        `json:"field_names"`
	FieldValues [][]interface{} `json:"field_values"`
}

func (m MatchItem) getItems(i int) map[string]interface{} {
	count := m.getResultCount()
	if i >= count || i < 0 {
		return nil
	}
	values := m.FieldValues[i]
	itemMap := make(map[string]interface{})
	for i := 0; i < len(m.FieldNames); i++ {
		itemMap[m.FieldNames[i]] = values[i]
	}
	return itemMap
}

func (m MatchItem) getResultCount() int {
	return len(m.FieldValues)
}

type ReadResult struct {
	ErrorCode    int         `json:"error_code"`
	ErrorMessage string      `json:"error_message"`
	MatchItems   MatchItem   `json:"match_items"`
	TraceInfo    interface{} `json:"trace_info"`
}

type WriteResult struct {
	Errno int `json:"errno"`
}

type Result struct {
	MatchItems *MatchItem   `json:"match_items"`
	TraceInfo  *interface{} `json:"trace_info"`
}
