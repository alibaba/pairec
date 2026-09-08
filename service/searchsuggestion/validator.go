package searchsuggestion

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/alibaba/pairec/v2/log"
	"golang.org/x/text/unicode/norm"
)

var (
	urlPattern      = regexp.MustCompile(`(?i)(https?://|www\.|\b[a-z0-9.-]+\.(com|cn|net|org)\b)`)
	internalPattern = regexp.MustCompile(`(?i)(item[_ -]?id|knowledge[_ -]?id|candidate[_ -]?id|search_goods|emit_suggestions|\[\[citation:|\[[0-9]+\]|【[0-9]+】)`)
	markdownPattern = regexp.MustCompile("(?m)(^\\s{0,3}(#{1,6}|[-*+]\\s|>\\s)|```|`|\\[[^]]*\\]\\([^)]*\\))")
)

func Validate(suggestions []string, count int, currentQuery string, minLength, maxLength int) ([]string, *Error) {
	if len(suggestions) != count {
		return nil, NewError(CodeValidationFailed, true, fmt.Errorf("expected %d suggestions, got %d", count, len(suggestions)))
	}
	baseline := normalizeComparable(currentQuery)
	seen := make(map[string]struct{}, len(suggestions))
	validated := make([]string, 0, len(suggestions))
	for _, suggestion := range suggestions {
		suggestion = strings.TrimSpace(norm.NFKC.String(suggestion))
		length := utf8.RuneCountInString(suggestion)
		normalized := normalizeComparable(suggestion)
		_, duplicate := seen[normalized]
		reason := ""
		switch {
		case length < minLength || length > maxLength:
			reason = "length"
		case strings.IndexFunc(suggestion, unicode.IsControl) >= 0:
			reason = "control"
		case urlPattern.MatchString(suggestion):
			reason = "url"
		case internalPattern.MatchString(suggestion):
			reason = "internal_marker"
		case markdownPattern.MatchString(suggestion) || strings.ContainsAny(suggestion, "{}<>"):
			reason = "markup"
		case normalized == "" || (baseline != "" && normalized == baseline):
			reason = "repeated_query"
		case duplicate:
			reason = "duplicate"
		}
		if reason != "" {
			log.Warning(fmt.Sprintf("module=SearchSuggestion\tevent=filtered\treason=%s", reason))
			continue
		}
		seen[normalized] = struct{}{}
		validated = append(validated, suggestion)
	}
	if len(validated) == 0 {
		return nil, NewError(CodeValidationFailed, true, fmt.Errorf("no valid suggestions"))
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
