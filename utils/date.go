package utils

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var dateExprReg *regexp.Regexp
var spaceReg *regexp.Regexp

func init() {
	dateReg := "\\$\\[(.*?)\\]"
	dateExprReg = regexp.MustCompile(dateReg)
	spaceReg = regexp.MustCompile("\\s+")
}

// FormatDate 日期转字符串
func FormatDate(date time.Time, dateStyle string) string {
	layout := dateStyle
	layout = strings.Replace(layout, "yyyy", "2006", 1)
	layout = strings.Replace(layout, "yy", "06", 1)
	layout = strings.Replace(layout, "MM", "01", 1)
	layout = strings.Replace(layout, "dd", "02", 1)
	layout = strings.Replace(layout, "HH", "15", 1)
	layout = strings.Replace(layout, "mm", "04", 1)
	layout = strings.Replace(layout, "ss", "05", 1)
	layout = strings.Replace(layout, "SSS", "000", -1)

	return date.Format(layout)
}

func IsDateExpression(expression string) bool {
	expr := dateExprReg.FindStringSubmatch(expression)
	return len(expr) > 1
}

func EvalDate(expression string) (string, bool) {
	r := dateExprReg.FindStringIndex(expression)
	if len(r) < 2 {
		return expression, false
	}
	format := expression[r[0]+2 : r[1]-1]
	diff := expression[r[1]:]
	diff = spaceReg.ReplaceAllString(diff, "")
	delta, err := strconv.Atoi(diff)
	if err != nil {
		return expression, false
	}
	date := time.Now().AddDate(0, 0, delta)
	result := FormatDate(date, format)
	return result, true
}
