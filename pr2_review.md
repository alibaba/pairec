## 📋 Review Summary

本次 PR 通过引入对象复用（`sync.Pool`）、减少切片拷贝、批量调优以及基于阈值的限流与背压机制，显著降低了精排阶段的 P99 延迟及内存分配（GC 压力）。整体优化思路清晰且切中痛点，但在资源回收逻辑和配置的初始化上存在一些可以优化的细节。

## 🔍 General Feedback

- **配置加载逻辑 (`configloader.go`)**: `initCallbackHandler` 方法在遍历 `config.CallBackConfs` (map) 时遇到第一个配置就进行初始化并 `return`。由于 map 的遍历是无序的，当存在多个场景的回调配置时，这会导致每次启动时 WorkerPool 的配置不可预测。既然 WorkerPool 是全局的，建议将其提取为全局配置，而不是依附于某个随机的场景。
- **并发初始化冲突 (`web/callback_controller_handler.go`)**: `InitHandler` 和 `Send` 使用了同一个 `sync.Once` 进行防并发初始化。如果在 `configloader.go` 执行前就有流量触发了 `Send`，则会按照硬编码的默认值（20/5000/false）初始化，使得后续配置文件中的设定永久失效。建议在 `Send` 触发默认初始化时加一条 Error 日志，或者调整服务启动顺序以保证 `InitHandler` 绝对先行。
- **`sync.Pool` 使用细节 (`algorithm/eas/easyrec_request.go`)**: 处理 Oversized buffer 时，当容量大于 `1<<20` 时，直接 return 或不执行 put 即可，没有必要主动去分配一个新的 `proto.NewBuffer` 再塞回 Pool 中。主动 new 再 put 反而增加了没有必要的 GC 压力。

### Code Suggestion for `algorithm/eas/easyrec_request.go`
```go
	// Cap oversized buffers to avoid retaining large allocations in pool
	if cap(buf.Bytes()) <= 1<<20 {
		marshalBufPool.Put(buf)
	}
```
