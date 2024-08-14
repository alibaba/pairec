package be

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type RecallType string

const (
	RecallTypeX2I    RecallType = "X2I"
	RecallTypeVector RecallType = "Vector"
)

type FilterClause struct {
	Filter Filter `json:"filter"`
	Clause string `json:"clause"`
}

func NewFilterClause(filter Filter) *FilterClause {
	return &FilterClause{Filter: filter}
}

func (c *FilterClause) GetFilter() Filter {
	return c.Filter
}

func (c *FilterClause) SetFilter(filter *Filter) *FilterClause {
	c.Filter = *filter
	return c
}

func (c *FilterClause) SetClause(clause string) *FilterClause {
	c.Clause = clause
	return c
}

func (c *FilterClause) BuildParams() string {
	queryClause := ""
	if c.Clause != "" {
		queryClause = c.Clause
	} else {
		queryClause = c.Filter.GetConditionValue()
	}
	return url.QueryEscape(queryClause)
}

type ExposureClause struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

func NewExposureClause(values []string) *ExposureClause {
	return &ExposureClause{Name: "user_id", Values: values}
}

func (c *ExposureClause) BuildParams() string {
	if len(c.Values) == 0 {
		return ""
	}
	return strings.Join(c.Values[:], ",")
}

type ScorerClause struct {
	Clause string `json:"clause"`
}

func NewScorerClause(clause string) *ScorerClause {
	return &ScorerClause{Clause: clause}
}

type RecallParam struct {
	RecallName   string        `json:"recall_name"`
	TriggerItems []string      `json:"trigger_items"`
	RecallType   RecallType    `json:"recall_type"`
	ReturnCount  int           `json:"return_count"`
	ScorerClause *ScorerClause `json:"scorer_clause"`
}

func NewRecallParam() *RecallParam {
	return &RecallParam{}
}

func (p *RecallParam) Validate() error {
	if len(p.TriggerItems) == 0 {
		return InvalidParamsError{fmt.Sprintf("Empty trigger items for recall[%s]", p.RecallName)}
	}
	if p.RecallName != "" && p.ReturnCount <= 0 {
		return InvalidParamsError{fmt.Sprintf("Return count should be greater than 0 for recall[%s]", p.RecallName)}
	}
	return nil
}

func (p *RecallParam) SetRecallName(name string) *RecallParam {
	p.RecallName = strings.TrimSpace(name)
	return p
}

func (p *RecallParam) SetTriggerItems(items []string) *RecallParam {
	p.TriggerItems = items
	return p
}

func (p *RecallParam) SetRecallType(recallType RecallType) *RecallParam {
	p.RecallType = recallType
	return p
}

func (p *RecallParam) SetScorerClause(clause *ScorerClause) *RecallParam {
	p.ScorerClause = clause
	return p
}

func (p *RecallParam) SetReturnCount(returnCount int) *RecallParam {
	p.ReturnCount = returnCount
	return p
}

func (p *RecallParam) flatTriggers() string {
	if p.RecallType == RecallTypeX2I {
		return strings.Join(p.TriggerItems[:], ",")
	} else {
		return strings.Join(p.TriggerItems[:], ";")
	}
}

func (p *RecallParam) getTriggerKey() string {
	if p.RecallName == "" {
		return "trigger_list"
	} else {
		return p.RecallName + "_trigger_list"
	}
}

func (p *RecallParam) getScorerKey() string {
	if p.RecallName == "" {
		return "score_rule"
	} else {
		return p.RecallName + "_score_rule"
	}
}

func (p RecallParam) getReturnCountKey() string {
	if p.RecallName == "" {
		return "return_count"
	} else {
		return p.RecallName + "_return_count"
	}
}

type ReadRequest struct {
	ReturnCount    int               `json:"return_count"`
	BizName        string            `json:"biz_name"`
	FilterClause   *FilterClause     `json:"filter_clause"`
	RecallParams   []RecallParam     `json:"recall_params"`
	ExposureClause *ExposureClause   `json:"exposure_clause"`
	QueryParams    map[string]string `json:"query_params"`
	IsRawRequest   bool              `json:"is_raw_request"`
	OutFmt         string            `json:"out_fmt"`
	IsPost         bool              `json:"is_post"`
}

func NewReadRequest(bizName string, returnCount int) *ReadRequest {
	return &ReadRequest{ReturnCount: returnCount, BizName: bizName, QueryParams: map[string]string{}, IsRawRequest: false, OutFmt: "fb2", IsPost: false}
}

func (r *ReadRequest) Validate() error {
	if r.ReturnCount <= 0 {
		return InvalidParamsError{"Return count should be greater than 0"}
	}
	if r.BizName == "" {
		return InvalidParamsError{"Empty biz name"}
	}
	if r.IsRawRequest {
		return nil
	}
	if len(r.RecallParams) == 0 {
		return InvalidParamsError{"Empty recall params"}
	}
	recallNames := map[string]bool{}
	for _, param := range r.RecallParams {
		recallError := param.Validate()
		if recallError != nil {
			return recallError
		}
		if recallNames[param.RecallName] {
			return InvalidParamsError{fmt.Sprintf("Duplicate recall name[%s] in RecallParams", param.RecallName)}
		}
		recallNames[param.RecallName] = true
	}
	return nil
}

func (r *ReadRequest) SetReturnCount(returnCount int) *ReadRequest {
	r.ReturnCount = returnCount
	return r
}

func (r *ReadRequest) SetBizName(bizName string) *ReadRequest {
	r.BizName = bizName
	return r
}

func (r *ReadRequest) SetFilterClause(clause *FilterClause) *ReadRequest {
	r.FilterClause = clause
	return r
}

func (r *ReadRequest) SetRecallParams(recallParams []RecallParam) *ReadRequest {
	r.RecallParams = recallParams
	return r
}

func (r *ReadRequest) AddRecallParam(param *RecallParam) *ReadRequest {
	r.RecallParams = append(r.RecallParams, *param)
	return r
}

func (r *ReadRequest) AddQueryParam(key string, value string) *ReadRequest {
	r.QueryParams[key] = value
	return r
}

func (r *ReadRequest) SetQueryParams(params map[string]string) *ReadRequest {
	r.QueryParams = params
	return r
}

func (r *ReadRequest) BuildParams() string {
	query := map[string]string{}
	query["biz_name"] = "searcher"
	query["p"] = r.BizName
	query["s"] = r.BizName
	query["return_count"] = strconv.Itoa(r.ReturnCount)

	_, exist := r.QueryParams["outfmt"]
	if !exist {
		r.QueryParams["outfmt"] = r.OutFmt
	}

	if r.FilterClause != nil && r.FilterClause.BuildParams() != "" {
		query["filter_rule"] = r.FilterClause.BuildParams()
	}
	for _, recallParam := range r.RecallParams {
		query[recallParam.getTriggerKey()] = recallParam.flatTriggers()
		if recallParam.RecallName != "" {
			query[recallParam.getReturnCountKey()] = strconv.Itoa(recallParam.ReturnCount)
		}
		if recallParam.ScorerClause != nil {
			query[recallParam.getScorerKey()] = url.QueryEscape(recallParam.ScorerClause.Clause)
		}
	}

	if r.ExposureClause != nil {
		query[r.ExposureClause.Name] = r.ExposureClause.BuildParams()
	}

	if len(r.QueryParams) != 0 {
		for k, v := range r.QueryParams {
			query[k] = v
		}
	}

	var params []string
	for k, v := range query {
		params = append(params, k+"="+v)
	}
	return strings.Join(params[:], "&")
}

func (r *ReadRequest) BuildUri() url.URL {
	uri := url.URL{Path: "be"}
	if r.IsPost {
		return uri
	}
	uri.RawQuery = r.BuildParams()
	return uri
}
