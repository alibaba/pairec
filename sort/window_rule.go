package sort

import (
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
)

type ItemKey string
type ItemKeyFunc func(item *module.Item, context *context.RecommendContext) string
type WindowRule struct {
	Size    int
	Min     int
	Max     int
	ItemKey ItemKeyFunc
	windows []ItemKey
	Total   int
}

// size : window size
// total: window total size, recommend item list size
func NewWindowRule(min, max, size, total int, f ItemKeyFunc) *WindowRule {
	return &WindowRule{
		Size:    size,
		Min:     min,
		Max:     max,
		ItemKey: f,
		windows: make([]ItemKey, total),
		Total:   total,
	}
}

// item add to the window list by the param pos, a int value of position
// pos start from 1
// return true or false, if add success, return true, else  return false
func (w *WindowRule) AddToWindow(pos int, item *module.Item, context *context.RecommendContext) bool {
	if pos > w.Total {
		return false
	}

	key := w.ItemKey(item, context)

	if pos == 1 {
		w.windows[pos-1] = ItemKey(key)
		return true
	}

	// i := (pos - 1) / w.Size
	count := 0
	start := pos - 1 - w.Size
	if start < 0 {
		start = 0
	}

	for start < pos {
		count = 0
		for i := start; i < start+w.Size && i < w.Total; i++ {
			if w.windows[i] == ItemKey(key) {
				count++
				if count == w.Max {
					return false
				}
			}
		}

		start++
	}

	w.windows[pos-1] = ItemKey(key)
	return true
}
