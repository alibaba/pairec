package searchsuggestion

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

var (
	urlPattern      = regexp.MustCompile(`(?i)(https?://|www\.|\b[a-z0-9.-]+\.(com|cn|net|org)\b)`)
	internalPattern = regexp.MustCompile(`(?i)(item[_ -]?id|knowledge[_ -]?id|candidate[_ -]?id|search_goods|emit_suggestions|\bprompt\b|\btool\b|citation|\bK[0-9]+\b|\[[0-9]+\]|【[0-9]+】)`)
	markdownPattern = regexp.MustCompile("(?m)(^\\s{0,3}(#{1,6}|[-*+]\\s|>\\s)|```|`|\\[[^]]*\\]\\([^)]*\\))")
)

func Validate(suggestions []string, count int, currentQuery string) ([]string, *Error) {
	if len(suggestions) != count {
		return nil, NewError(CodeValidationFailed, true, fmt.Errorf("expected %d suggestions, got %d", count, len(suggestions)))
	}
	baseline := normalizeComparable(currentQuery)
	seen := make(map[string]struct{}, len(suggestions))
	validated := make([]string, 0, len(suggestions))
	for _, suggestion := range suggestions {
		suggestion = strings.TrimSpace(norm.NFKC.String(suggestion))
		length := utf8.RuneCountInString(suggestion)
		if length < 4 || length > 80 {
			return nil, NewError(CodeValidationFailed, true, fmt.Errorf("suggestion length out of range"))
		}
		for _, r := range suggestion {
			if r == '\n' || r == '\r' || unicode.IsControl(r) {
				return nil, NewError(CodeValidationFailed, true, fmt.Errorf("suggestion contains control characters"))
			}
		}
		if urlPattern.MatchString(suggestion) || internalPattern.MatchString(suggestion) || markdownPattern.MatchString(suggestion) ||
			strings.ContainsAny(suggestion, "{}<>") {
			return nil, NewError(CodeValidationFailed, true, fmt.Errorf("suggestion contains forbidden syntax"))
		}
		normalized := normalizeComparable(suggestion)
		if normalized == "" || (baseline != "" && normalized == baseline) {
			return nil, NewError(CodeValidationFailed, true, fmt.Errorf("suggestion repeats current query"))
		}
		if _, exists := seen[normalized]; exists {
			return nil, NewError(CodeValidationFailed, true, fmt.Errorf("duplicate suggestions"))
		}
		seen[normalized] = struct{}{}
		validated = append(validated, suggestion)
	}
	return validated, nil
}

func normalizeComparable(value string) string {
	value = strings.ToLower(norm.NFKC.String(value))
	var builder strings.Builder
	space := false
	for _, r := range value {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			space = false
		case unicode.IsSpace(r):
			if builder.Len() > 0 && !space {
				builder.WriteByte(' ')
				space = true
			}
		}
	}
	return strings.TrimSpace(builder.String())
}
