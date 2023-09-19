package sort

import (
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
)

type MatrixDotFunc func(threshold float64, pre, cur *module.Item) bool
type MatrixScatterRule struct {
	Threshold float64
	dotFunc   MatrixDotFunc
	items     []*module.Item
}

func NewMatrixScatterRule(threshold float64, items []*module.Item, f MatrixDotFunc) *MatrixScatterRule {
	rule := &MatrixScatterRule{
		Threshold: threshold,
		dotFunc:   f,
		items:     items,
	}

	return rule
}

// pos start from 1
// return true or false, if return true, the item can add into the items
func (r *MatrixScatterRule) Match(pos int, item *module.Item, context *context.RecommendContext) bool {

	if pos < 2 {
		return true
	}
	preItem := r.items[pos-2]
	if preItem == nil {
		return true
	}

	return r.dotFunc(r.Threshold, preItem, item)
}
