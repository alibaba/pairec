package aishopping

import (
	"strconv"
	"strings"
)

type replyStreamer struct {
	writer   *StreamWriter
	indexMap map[int]string
	refIDs   []string
	refPos   int
	maxItems int
	used     int
	buffer   string
}

func newReplyStreamer(writer *StreamWriter, indexMap map[int]string, maxItems int) *replyStreamer {
	return &replyStreamer{
		writer:   writer,
		indexMap: indexMap,
		refIDs:   orderedItemIDs(indexMap, maxItems),
		maxItems: maxItems,
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
			if err := s.writer.EmitContent(s.buffer[:emitLen]); err != nil {
				return err
			}
			s.buffer = s.buffer[emitLen:]
			continue
		}
		if markerStart > 0 {
			if err := s.writer.EmitContent(s.buffer[:markerStart]); err != nil {
				return err
			}
			s.buffer = s.buffer[markerStart:]
			continue
		}
		if strings.HasPrefix(s.buffer, "[ref]") {
			if s.refPos < len(s.refIDs) && s.used < s.maxItems {
				if err := s.writer.EmitCitation(s.refIDs[s.refPos]); err != nil {
					return err
				}
				s.used++
			}
			s.refPos++
			s.buffer = s.buffer[len("[ref]"):]
			continue
		}
		markerEnd := strings.Index(s.buffer[2:], "]]")
		if markerEnd < 0 {
			if final {
				if err := s.writer.EmitContent(s.buffer); err != nil {
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
			index, _ := strconv.Atoi(inner)
			if itemID := s.indexMap[index]; itemID != "" && s.used < s.maxItems {
				if err := s.writer.EmitCitation(itemID); err != nil {
					return err
				}
				s.used++
			}
		} else if err := s.writer.EmitContent(marker); err != nil {
			return err
		}
		s.buffer = s.buffer[markerEnd+2:]
	}
	return nil
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
