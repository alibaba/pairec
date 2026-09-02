package aishopping

import (
	"strconv"
	"strings"
)

type replyStreamer struct {
	writer      *StreamWriter
	resolver    *replyMarkerResolver
	maxItems    int
	used        int
	buffer      string
	seenItemIDs map[string]struct{}
	textState   markerTextState
}

type markerTextState struct {
	afterMarker       bool
	pendingHorizontal string
}

func newReplyStreamer(writer *StreamWriter, indexMap map[int]string, replyItemIDs []string, maxItems int, strictRankOrder bool) *replyStreamer {
	return &replyStreamer{
		writer:      writer,
		resolver:    newReplyMarkerResolver(indexMap, replyItemIDs, maxItems, strictRankOrder),
		maxItems:    maxItems,
		seenItemIDs: make(map[string]struct{}),
	}
}

func (s *replyStreamer) Feed(delta string) error {
	if delta == "" {
		return nil
	}
	s.buffer += delta
	return s.flush(false)
}

func (s *replyStreamer) Finish() error {
	return s.flush(true)
}

func (s *replyStreamer) flush(final bool) error {
	for s.buffer != "" {
		markerStart := firstMarkerStart(s.buffer)
		if markerStart < 0 {
			emitLen := len(s.buffer)
			if !final {
				emitLen -= pendingMarkerSuffixLen(s.buffer)
			}
			if emitLen <= 0 {
				return nil
			}
			if err := s.emitContent(s.buffer[:emitLen]); err != nil {
				return err
			}
			s.buffer = s.buffer[emitLen:]
			continue
		}
		if markerStart > 0 {
			if err := s.emitContent(s.buffer[:markerStart]); err != nil {
				return err
			}
			s.buffer = s.buffer[markerStart:]
			continue
		}
		if strings.HasPrefix(s.buffer, "[ref]") {
			if err := s.emitMarker(s.resolver.nextRef()); err != nil {
				return err
			}
			s.buffer = s.buffer[len("[ref]"):]
			continue
		}
		markerEnd := strings.Index(s.buffer[2:], "]]")
		if markerEnd < 0 {
			if final {
				if err := s.emitContent(s.buffer); err != nil {
					return err
				}
				s.buffer = ""
			}
			return nil
		}
		markerEnd += 2
		marker := s.buffer[:markerEnd+2]
		inner := s.buffer[2:markerEnd]
		if isIndexMarker(inner) {
			index, err := strconv.Atoi(inner)
			if err == nil {
				if err := s.emitMarker(s.resolver.resolveIndex(index)); err != nil {
					return err
				}
			}
		} else if err := s.emitContent(marker); err != nil {
			return err
		}
		s.buffer = s.buffer[markerEnd+2:]
	}
	return nil
}

func (s *replyStreamer) emitContent(content string) error {
	return s.writer.EmitContent(s.textState.consume(content))
}

func (s *replyStreamer) emitMarker(itemID string) error {
	s.textState.mark()
	if itemID == "" || s.used >= s.maxItems {
		return nil
	}
	if _, exists := s.seenItemIDs[itemID]; exists {
		return nil
	}
	if err := s.writer.EmitCitation(itemID); err != nil {
		return err
	}
	s.seenItemIDs[itemID] = struct{}{}
	s.used++
	return nil
}

func (s *markerTextState) mark() {
	s.afterMarker = true
	s.pendingHorizontal = ""
}

func (s *markerTextState) consume(content string) string {
	if content == "" || !s.afterMarker {
		return content
	}
	leading := 0
	for leading < len(content) && isHorizontalWhitespace(content[leading]) {
		leading++
	}
	s.pendingHorizontal += content[:leading]
	if leading == len(content) {
		return ""
	}

	rest := content[leading:]
	s.afterMarker = false
	if rest[0] == '\n' {
		s.pendingHorizontal = ""
		return rest
	}
	result := s.pendingHorizontal + rest
	s.pendingHorizontal = ""
	return result
}

func isHorizontalWhitespace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\r'
}

func firstMarkerStart(value string) int {
	indexStart := strings.Index(value, "[[")
	refStart := strings.Index(value, "[ref]")
	if indexStart < 0 {
		return refStart
	}
	if refStart < 0 || indexStart < refStart {
		return indexStart
	}
	return refStart
}

func pendingMarkerSuffixLen(value string) int {
	prefixes := []string{"[ref", "[re", "[r", "["}
	for _, prefix := range prefixes {
		if strings.HasSuffix(value, prefix) {
			return len(prefix)
		}
	}
	return 0
}

func isIndexMarker(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
