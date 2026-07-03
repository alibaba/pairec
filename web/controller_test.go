package web

import (
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestLogResponseBody_ShortBody(t *testing.T) {
	c := &Controller{RequestId: "test-req-1"}
	r := httptest.NewRequest("GET", "/api/recommend", nil)

	// Should not panic for short body
	c.LogResponseBody(r, "short response")
}

func TestLogResponseBody_LongASCII(t *testing.T) {
	c := &Controller{RequestId: "test-req-2"}
	r := httptest.NewRequest("GET", "/api/recommend", nil)

	// Create a body larger than maxChunkSize (4000)
	body := strings.Repeat("a", 10000)
	c.LogResponseBody(r, body)
}

func TestLogResponseBody_UTF8Boundary(t *testing.T) {
	c := &Controller{RequestId: "test-req-3"}
	r := httptest.NewRequest("GET", "/api/recommend", nil)

	// Create a body with multi-byte UTF-8 characters (Chinese: 3 bytes each)
	// Fill to just over maxChunkSize so the split point may land mid-character
	ch := "中" // 3 bytes
	body := strings.Repeat(ch, 2000) // 6000 bytes, > 4000
	if !utf8.ValidString(body) {
		t.Fatal("test input should be valid UTF-8")
	}

	// Should not panic and should produce valid UTF-8 chunks
	c.LogResponseBody(r, body)
}

func TestLogResponseBody_InvalidUTF8_NoInfiniteLoop(t *testing.T) {
	c := &Controller{RequestId: "test-req-4"}
	r := httptest.NewRequest("GET", "/api/recommend", nil)

	// Create a body with long stretches of continuation bytes (0x80)
	// that have no valid UTF-8 start byte - this would cause infinite loop
	// without the guard clause.
	body := strings.Repeat("\x80", 5000)

	// Should complete without hanging
	c.LogResponseBody(r, body)
}

func TestLogResponseBody_MixedInvalidUTF8(t *testing.T) {
	c := &Controller{RequestId: "test-req-5"}
	r := httptest.NewRequest("GET", "/api/recommend", nil)

	// Mix valid ASCII with invalid continuation bytes at chunk boundary
	body := strings.Repeat("x", 3998) + "\x80\x80\x80\x80" + strings.Repeat("y", 2000)

	// Should complete without panic or hang
	c.LogResponseBody(r, body)
}
