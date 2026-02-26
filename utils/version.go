package utils

import (
	"strconv"
	"strings"
)

// IsVersionGreaterOrEqual compares two semantic version strings.
// Returns true if v1 >= v2.
// Supports formats: "1", "1.2", "1.2.3".
func IsVersionGreaterOrEqual(v1, v2 string) bool {
	parts1 := parseVersion(v1)
	parts2 := parseVersion(v2)

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int
		if i < len(parts1) {
			p1 = parts1[i]
		}
		if i < len(parts2) {
			p2 = parts2[i]
		}
		if p1 > p2 {
			return true
		}
		if p1 < p2 {
			return false
		}
	}
	return true // equal
}

func parseVersion(v string) []int {
	parts := strings.Split(v, ".")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		n, _ := strconv.Atoi(strings.TrimSpace(p))
		result = append(result, n)
	}
	return result
}
