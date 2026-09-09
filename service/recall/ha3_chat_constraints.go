package recall

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/alibaba/pairec/v2/algorithm/aichat"
	"github.com/alibaba/pairec/v2/recconf"
)

func cloneSearchGoodsConfig(conf *recconf.SearchGoodsConfig) *recconf.SearchGoodsConfig {
	if conf == nil {
		return nil
	}
	data, _ := json.Marshal(conf)
	var copy recconf.SearchGoodsConfig
	_ = json.Unmarshal(data, &copy)
	return &copy
}

func validateSearchGoodsConfig(conf *recconf.SearchGoodsConfig) {
	if conf == nil {
		return
	}
	names := map[string]bool{}
	for _, param := range conf.ConstraintParams {
		if !validHa3FieldName(param.Name) || names[param.Name] {
			panic("invalid or duplicate search constraint name")
		}
		names[param.Name] = true
		switch param.Kind {
		case "enum", "option_set":
			if !validHa3FieldName(param.Field) || len(param.Values) == 0 || len(param.EqualValues) != 0 {
				panic("search constraint requires a field and enum values")
			}
			for name, values := range param.Values {
				if strings.TrimSpace(name) == "" || len(values) == 0 {
					panic("empty search constraint value")
				}
				for _, value := range values {
					validateConstraintLiteral(value)
				}
			}
		case "all_eq":
			if param.Field != "" || len(param.Values) != 0 || len(param.EqualValues) == 0 {
				panic("all_eq requires named equal-field groups")
			}
			for name, fields := range param.EqualValues {
				if strings.TrimSpace(name) == "" || len(fields) == 0 {
					panic("empty all_eq constraint")
				}
				for field, value := range fields {
					if !validHa3FieldName(field) {
						panic("invalid constraint field")
					}
					validateConstraintLiteral(value)
				}
			}
		default:
			panic("unsupported search constraint kind")
		}
	}
}

func validateConstraintLiteral(value string) {
	if strings.TrimSpace(value) == "" || strings.ContainsAny(value, "\x00\r\n\t\\\"") {
		panic("invalid search constraint literal")
	}
}

func (r *Ha3ChatRecall) SearchGoodsTool() aichat.Tool {
	return aichat.FieldAwareSearchGoodsTool(r.conf.SearchGoodsConf)
}

func (r *Ha3ChatRecall) ValidateSearchGoodsRequest(req SearchGoodsRequest) error {
	_, err := r.buildConstraintExpr(req.Constraints)
	return err
}

func (r *Ha3ChatRecall) attachConstraintEvidence(result *SearchGoodsResult, constraints map[string]json.RawMessage) {
	if result == nil || len(constraints) == 0 || r.conf.SearchGoodsConf == nil {
		return
	}
	for _, param := range r.conf.SearchGoodsConf.ConstraintParams {
		raw, requested := constraints[param.Name]
		if !requested || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			continue
		}
		for i := range result.Hits {
			hit := &result.Hits[i]
			if hit.ConstraintEvidence == nil {
				hit.ConstraintEvidence = make(map[string]interface{})
			}
			if param.Kind != "all_eq" {
				hit.ConstraintEvidence[param.Name] = hit.Properties[param.Field]
				continue
			}
			var value string
			_ = json.Unmarshal(raw, &value)
			evidence := make(map[string]interface{})
			for field := range param.EqualValues[value] {
				evidence[field] = hit.Properties[field]
			}
			hit.ConstraintEvidence[param.Name] = evidence
		}
	}
}

func (r *Ha3ChatRecall) buildConstraintExpr(constraints map[string]json.RawMessage) (string, error) {
	params := map[string]recconf.SearchConstraintConfig{}
	if r.conf.SearchGoodsConf != nil {
		for _, param := range r.conf.SearchGoodsConf.ConstraintParams {
			params[param.Name] = param
		}
	}
	names := make([]string, 0, len(constraints))
	for name := range constraints {
		names = append(names, name)
	}
	sort.Strings(names)
	clauses := make([]string, 0, len(names))
	for _, name := range names {
		param, ok := params[name]
		if !ok {
			return "", fmt.Errorf("unknown constraint %q", name)
		}
		raw := bytes.TrimSpace(constraints[name])
		if bytes.Equal(raw, []byte("null")) {
			continue
		}
		clause, err := compileSearchConstraint(param, raw)
		if err != nil {
			return "", fmt.Errorf("constraint %s: %w", name, err)
		}
		if clause != "" {
			clauses = append(clauses, clause)
		}
	}
	return strings.Join(clauses, " AND "), nil
}

func compileSearchConstraint(param recconf.SearchConstraintConfig, raw json.RawMessage) (string, error) {
	if param.Kind == "all_eq" {
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return "", fmt.Errorf("expected a configured string value")
		}
		fields, ok := param.EqualValues[value]
		if !ok {
			return "", fmt.Errorf("unknown value %q", value)
		}
		clauses := make([]string, 0, len(fields))
		for field, literal := range fields {
			clauses = append(clauses, field+" = "+strconv.Quote(literal))
		}
		sort.Strings(clauses)
		return "(" + strings.Join(clauses, " AND ") + ")", nil
	}
	var selection struct {
		Any     []string `json:"any"`
		Exclude []string `json:"exclude"`
		Known   *bool    `json:"known"`
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&selection); err != nil {
		return "", err
	}
	if selection.Known != nil && !*selection.Known {
		return "", fmt.Errorf("known must be true or null")
	}
	if len(selection.Any) == 0 && len(selection.Exclude) == 0 && selection.Known == nil {
		return "", nil
	}
	allowed := map[string]bool{}
	if len(selection.Any) == 0 {
		for _, values := range param.Values {
			for _, value := range values {
				allowed[value] = true
			}
		}
	}
	for _, name := range selection.Any {
		values, ok := param.Values[name]
		if !ok {
			return "", fmt.Errorf("unknown value %q", name)
		}
		for _, value := range values {
			allowed[value] = true
		}
	}
	for _, name := range selection.Exclude {
		values, ok := param.Values[name]
		if !ok {
			return "", fmt.Errorf("unknown excluded value %q", name)
		}
		for _, value := range values {
			delete(allowed, value)
		}
	}
	if len(allowed) == 0 {
		return "", fmt.Errorf("conditions have no allowed value; do not drop a hard condition")
	}
	clauses := make([]string, 0, len(allowed))
	for value := range allowed {
		clauses = append(clauses, param.Field+" = "+strconv.Quote(value))
	}
	sort.Strings(clauses)
	// Positive membership requires evidence and binds any/exclude to the same option.
	return "(" + strings.Join(clauses, " OR ") + ")", nil
}
