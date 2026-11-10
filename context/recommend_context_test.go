package context

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestLoadOrStoreContextParam(t *testing.T) {
	for _, ctx := range []*RecommendContext{NewRecommendContext(), {}} {
		value := new(int)
		if actual, loaded := ctx.LoadOrStoreContextParam("key", value); loaded || actual != value {
			t.Fatal("首次写入未返回传入值")
		}
		if actual, loaded := ctx.LoadOrStoreContextParam("key", new(int)); !loaded || actual != value {
			t.Fatal("已有值被覆盖")
		}
		ctx.AddContextParam("nil", nil)
		if actual, loaded := ctx.LoadOrStoreContextParam("nil", value); !loaded || actual != nil {
			t.Fatal("已存 nil 应当视为已有条目")
		}
		ctx.DeleteContextParam("key")
		if actual, loaded := ctx.LoadOrStoreContextParam("key", value); loaded || actual != value {
			t.Fatal("删除后不能重新写入")
		}
	}
}

func TestLoadOrStoreContextParamConcurrent(t *testing.T) {
	ctx := NewRecommendContext()
	var winners atomic.Int32
	var wg sync.WaitGroup
	actuals := make([]interface{}, 64)
	for i := range actuals {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			actual, loaded := ctx.LoadOrStoreContextParam("shared", new(int))
			actuals[i] = actual
			if !loaded {
				winners.Add(1)
			}
			ctx.GetContextParam("shared")
		}(i)
	}
	wg.Wait()
	if winners.Load() != 1 {
		t.Fatalf("赢家数=%d，期望1", winners.Load())
	}
	for _, actual := range actuals {
		if actual != actuals[0] {
			t.Fatal("并发调用得到不同条目")
		}
	}
	if actual, loaded := ctx.LoadOrStoreContextParam("other", 123); loaded || actual != 123 {
		t.Fatal("不同 key 发生串值")
	}
}
