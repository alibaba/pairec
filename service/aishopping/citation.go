package aishopping

import (
	"regexp"
	"sort"
	"strconv"

	"github.com/alibaba/pairec/v2/algorithm/aichat"
)

var (
	indexMarkerRegexp = regexp.MustCompile(`\[\[(\d+)\]\]`)
	itemMarkerRegexp  = regexp.MustCompile(`\[\[item_id:[^\]]+\]\]`)
	replyRefRegexp    = regexp.MustCompile(`\[\[(\d+)\]\]|\[ref\]`)
)

type ReplyEvent struct {
	Content string
	ItemId  string
}

func resolveReplyEvents(reply string, indexMap map[int]string, maxItems int) (string, []ReplyEvent) {
	events := make([]ReplyEvent, 0)
	canonical := ""
	pos := 0
	used := 0
	refIDs := orderedItemIDs(indexMap, maxItems)
	refPos := 0
	matches := replyRefRegexp.FindAllStringSubmatchIndex(reply, -1)
	for _, match := range matches {
		text := reply[pos:match[0]]
		if text != "" {
			events = append(events, ReplyEvent{Content: text})
			canonical += text
		}
		pos = match[1]
		itemID := ""
		if match[2] >= 0 {
			n, _ := strconv.Atoi(reply[match[2]:match[3]])
			itemID = indexMap[n]
		} else if refPos < len(refIDs) {
			itemID = refIDs[refPos]
			refPos++
		}
		if itemID == "" || used >= maxItems {
			continue
		}
		used++
		canonical += "[[item_id:" + itemID + "]]"
		events = append(events, ReplyEvent{ItemId: itemID})
	}
	if pos < len(reply) {
		text := reply[pos:]
		events = append(events, ReplyEvent{Content: text})
		canonical += text
	}
	return canonical, events
}

func orderedItemIDs(indexMap map[int]string, maxItems int) []string {
	if maxItems <= 0 || len(indexMap) == 0 {
		return nil
	}
	indexes := make([]int, 0, len(indexMap))
	for index := range indexMap {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)
	itemIDs := make([]string, 0, len(indexes))
	for _, index := range indexes {
		if itemID := indexMap[index]; itemID != "" {
			itemIDs = append(itemIDs, itemID)
			if len(itemIDs) >= maxItems {
				break
			}
		}
	}
	return itemIDs
}

func maskHistoryMessages(messages []aichat.Message) []aichat.Message {
	copied := make([]aichat.Message, 0, len(messages))
	for _, msg := range messages {
		if msg.Role == "assistant" && msg.Content != "" {
			msg.Content = itemMarkerRegexp.ReplaceAllString(msg.Content, "[ref]")
		}
		copied = append(copied, msg)
	}
	return copied
}

func messagesWithPrompt(messages []aichat.Message, prompt string) []aichat.Message {
	if len(messages) == 0 {
		return []aichat.Message{{Role: "system", Content: prompt}}
	}
	messages[0] = aichat.Message{Role: "system", Content: prompt}
	return messages
}
