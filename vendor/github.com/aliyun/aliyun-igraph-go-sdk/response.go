package aliyun_igraph_go_sdk

type Response struct {
	Result []*Result `json:"result"`
}

func NewResponse(results []*Result) *Response {
	return &Response{Result: results}
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
	ErrorInfo []string  `json:"error_info"`
	Result    []*Result `json:"result"`
}

type WriteResult struct {
	Errno int `json:"errno"`
}

type Result struct {
	Data      []map[string]interface{} `json:"data"`
	TraceInfo map[string]string        `json:"trace_info"`
	ErrorInfo []string                 `json:"error_info"`
}
