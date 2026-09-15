package aishopping

import (
	"encoding/json"
	"strings"

	"github.com/alibaba/pairec/v2/log"
)

// Repair only missing values and standalone null-like literals belonging to
// top-level optional prices. The JSON decoder still validates all other syntax.
func normalizeOptionalPriceLiterals(arguments string) string {
	var output strings.Builder
	depth, copied := 0, 0
	for i := 0; i < len(arguments); i++ {
		switch arguments[i] {
		case '{', '[':
			depth++
		case '}', ']':
			depth--
		case '"':
			start := i
			i++
			for i < len(arguments) && arguments[i] != '"' {
				if arguments[i] == '\\' {
					i++
				}
				i++
			}
			if i >= len(arguments) {
				return arguments
			}
			if depth != 1 {
				continue
			}
			var key string
			if json.Unmarshal([]byte(arguments[start:i+1]), &key) != nil || (key != "min_price" && key != "max_price") {
				continue
			}
			colon := skipJSONSpace(arguments, i+1)
			if colon == len(arguments) || arguments[colon] != ':' {
				continue
			}
			valueStart := skipJSONSpace(arguments, colon+1)
			valueEnd := valueStart
			for valueEnd < len(arguments) && !strings.ContainsRune(" \t\r\n,}]", rune(arguments[valueEnd])) {
				valueEnd++
			}
			reason := "null_like_literal"
			switch strings.ToLower(arguments[valueStart:valueEnd]) {
			case "":
				if valueStart == len(arguments) || (arguments[valueStart] != ',' && arguments[valueStart] != '}') {
					continue // Do not repair a truncated object or mismatched delimiter.
				}
				reason = "missing_value"
			case "none", "nan", "+nan", "-nan", "infinity", "+infinity", "-infinity":
			default:
				continue
			}
			output.WriteString(arguments[copied:valueStart])
			output.WriteString("null")
			copied = valueEnd
			i = valueEnd - 1
			log.Info("module=AIShoppingChat\tevent=optional_price_ignored\tfield=" + key + "\treason=" + reason)
		}
	}
	if copied == 0 {
		return arguments
	}
	output.WriteString(arguments[copied:])
	return output.String()
}

func skipJSONSpace(value string, start int) int {
	for start < len(value) && strings.ContainsRune(" \t\r\n", rune(value[start])) {
		start++
	}
	return start
}
