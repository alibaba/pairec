package sort

import (
	"testing"
	"time"

	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

// TestRandomPositionStrategyDeadLoop 测试 randomPositionStrategy.BuildItems 是否存在死循环
// 当 items 数组已满时，代码会进入死循环
func TestRandomPositionStrategyDeadLoop(t *testing.T) {
	// 创建一个较小的 totalSize，便于触发死循环
	totalSize := 5

	// 创建 randomPositionStrategy
	config := &recconf.MixSortConfig{
		Number: 5, // 尝试放置 5 个 item
	}
	strategy := newRandomPositionStrategy(config, totalSize)
	strategy.totalSize = totalSize

	// 预填充一些 item 到 strategy
	for i := 0; i < 5; i++ {
		item := module.NewItem(string(rune('a' + i)))
		strategy.AppendItem(item)
	}

	// 创建一个已经部分填充的 items 数组
	items := make([]*module.Item, totalSize)
	// 预先填满所有位置
	for i := 0; i < totalSize; i++ {
		items[i] = module.NewItem(string(rune('x' + i)))
	}

	// 使用 channel 来检测是否超时（死循环）
	done := make(chan bool)
	go func() {
		strategy.BuildItems(items)
		done <- true
	}()

	// 等待最多 2 秒，如果超时则认为存在死循环
	select {
	case <-done:
		// 正常完成，测试通过（没有死循环）
		t.Log("BuildItems completed without dead loop")
	case <-time.After(2 * time.Second):
		// 超时，存在死循环
		t.Fatal("Dead loop detected: BuildItems did not complete within 2 seconds")
	}
}

// TestRandomPositionStrategyPartialFilled 测试部分填充时的行为
func TestRandomPositionStrategyPartialFilled(t *testing.T) {
	totalSize := 10
	config := &recconf.MixSortConfig{
		Number: 3,
	}
	strategy := newRandomPositionStrategy(config, totalSize)
	strategy.totalSize = totalSize

	// 添加 3 个 item
	for i := 0; i < 3; i++ {
		item := module.NewItem(string(rune('a' + i)))
		strategy.AppendItem(item)
	}

	// 创建空的 items 数组
	items := make([]*module.Item, totalSize)

	// 应该正常完成
	done := make(chan bool)
	go func() {
		strategy.BuildItems(items)
		done <- true
	}()

	select {
	case <-done:
		// 检查是否放置了 3 个 item
		count := 0
		for i, item := range items {
			if item != nil {
				count++
				t.Logf("item at index %d: %s", i, item.String())
			}
		}
		if count != 3 {
			t.Errorf("Expected 3 items placed, got %d", count)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Dead loop detected")
	}
}
