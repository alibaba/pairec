package recall

import (
	"encoding/json"
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

func validateToolParamLiteral(value string) {
	if strings.TrimSpace(value) == "" || strings.ContainsAny(value, "\x00\r\n\t\\\"") {
		panic("invalid search tool parameter literal")
	}
}

func (r *Ha3ChatRecall) SearchGoodsTool() aichat.Tool {
	return aichat.FieldAwareSearchGoodsTool(r.conf.SearchGoodsConf)
}

func (r *Ha3ChatRecall) ValidateSearchGoodsRequest(req SearchGoodsRequest) error {
	_, err := r.buildToolParamsExpr(req.ToolParams)
	return err
}

func (r *Ha3ChatRecall) attachToolParamEvidence(result *SearchGoodsResult, values map[string]json.RawMessage) {
	if result == nil {
		return
	}
	for _, param := range r.toolParams() {
		terms, err := toolTermValues(values[param.Name])
		if err != nil || len(terms) == 0 {
			continue
		}
		for i := range result.Hits {
			hit := &result.Hits[i]
			if hit.ConstraintEvidence == nil {
				hit.ConstraintEvidence = make(map[string]interface{})
			}
			hit.ConstraintEvidence[param.Name] = hit.Properties[param.Field]
		}
	}
}
