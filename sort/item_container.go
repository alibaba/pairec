package sort

import (
	"math/rand"

	"github.com/alibaba/pairec/module"
)

// Strategy use to split items by which way
type Strategy int

const (
	// RecallName split Items to ItemMap by recallName
	SplitByRecallName Strategy = iota
)
const (
	ItemDefaultName = "DEFAULT"
)

type SortStrategy int

const (
	SortByScoreStrategy SortStrategy = iota
	SortByRandomStrategy
)

type ItemContainer struct {
	itemMap  map[string]*ItemSlot
	strategy Strategy
	size     int
}

// ItemSlot is a group of items split by the Strategy, like SplitByRecallName
type ItemSlot struct {
	// size in this slot, need return number of items
	size  int
	index int
	// capacity the slot at most have number of items, general, capacity >= size
	// if other slot not have enough items, can use the extra items to supplement
	capacity     int
	sortStrategy SortStrategy
	items        []*module.Item
}

func (s *ItemSlot) full() bool {
	return len(s.items) == s.capacity
}
func NewItemContaienr(strategy Strategy, size int) *ItemContainer {
	container := &ItemContainer{
		itemMap:  make(map[string]*ItemSlot),
		strategy: strategy,
		size:     size,
	}

	return container
}
func (c *ItemContainer) Assign(name string, size int, capacity int, sortStrategy SortStrategy) {
	if _, ok := c.itemMap[name]; !ok {
		slot := &ItemSlot{
			size:         size,
			index:        0,
			capacity:     capacity,
			sortStrategy: sortStrategy,
		}
		c.itemMap[name] = slot
	}

}
func (c *ItemContainer) Split(items []*module.Item) {
	if c.strategy == SplitByRecallName {
		c.splitByRecallName(items)
	}
}

func (c *ItemContainer) splitByRecallName(items []*module.Item) {
	for _, item := range items {
		name := item.GetRecallName()
		slot, ok := c.itemMap[name]
		if ok {
			if !slot.full() {
				slot.items = append(slot.items, item)
			}
		} else {
			// not found, put item in  ItemDefaultName map
			slot = c.itemMap[ItemDefaultName]
			if slot != nil && !slot.full() {
				slot.items = append(slot.items, item)
			}
		}
	}
}
func (c *ItemContainer) Assembly() (ret []*module.Item) {
	newItems := make([]*module.Item, c.size)
	for _, slot := range c.itemMap {
		// 第一轮优先只处理 SortByRandomStrategy, 防止后续需要移动元素位置
		if slot.sortStrategy != SortByRandomStrategy {
			continue
		}
		var itemSet []*module.Item
		if len(slot.items) >= slot.size {
			itemSet = append(itemSet, slot.items[:slot.size]...)
			slot.index = slot.size
		} else {
			itemSet = append(itemSet, slot.items...)
			slot.index = len(slot.items)
		}
		for _, item := range itemSet {
			index := rand.Intn(c.size)
			i := index
			if newItems[i] != nil {
				for i = (index + 1) % c.size; i != index; i = (i + 1) % c.size {
					if newItems[i] == nil {
						break
					}
				}
			}
			if newItems[i] == nil {
				newItems[i] = item
			}
		}
	}

	for _, slot := range c.itemMap {
		// 第二轮不再处理 SortByRandomStrategy, 已经在第一轮处理过
		if slot.sortStrategy == SortByRandomStrategy {
			continue
		}
		var itemSet []*module.Item
		if len(slot.items) >= slot.size {
			itemSet = append(itemSet, slot.items[:slot.size]...)
			slot.index = slot.size
		} else {
			itemSet = append(itemSet, slot.items...)
			slot.index = len(slot.items)
		}
		i := 0
		for _, item := range itemSet {
			for i < c.size && newItems[i] != nil {
				i++
			}
			if i >= c.size {
				break
			}
			newItems[i] = item
			i++
		}
	}
	for _, item := range newItems {
		if item != nil {
			ret = append(ret, item)
		}
	}
	// if ret size not enough
	if len(ret) < c.size {
		diff := c.size - len(ret)
		var diffItems []*module.Item

		// first use ItemDefaultName slot
		if slot, ok := c.itemMap[ItemDefaultName]; ok {
			if len(slot.items) > slot.index {
				diffItems = append(diffItems, slot.items[slot.index:]...)
			}
		}

		if len(diffItems) < diff {
			for name, slot := range c.itemMap {
				if name != ItemDefaultName {
					if len(slot.items) > slot.index {
						diffItems = append(diffItems, slot.items[slot.index:]...)
					}
				}
			}
		}

		if len(diffItems) > diff {
			diffItems = diffItems[:diff]
		}

		ret = append(ret, diffItems...)
	}

	return
}
