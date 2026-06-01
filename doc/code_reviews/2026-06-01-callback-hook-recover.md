# Code Review: callback hook 链路顶层 recover 兜底 (P0)

- 日期: 2026-06-01
- 分支: fix/callback-empty-itemlist-panic
- Review 结论: 可以合并但建议改进
- 改动范围:
  - service/hook/recommend.go (新增 SafeRun helper)
  - service/user_recommend.go (主推荐路径 hook 派发用 SafeRun)
  - service/user_recall_service.go (召回路径 hook 派发用 SafeRun)
  - web/callback_controller_handler.go (worker 池调用 doCallbackLog 加 runCallbackSafely)
  - service/hook/recommend_test.go (新增)
  - web/callback_controller_handler_test.go (新增 worker recover 测试)

## 背景

callback hook 链路上有两处 P0 风险:

1. service/user_recommend.go:173 与 service/user_recall_service.go:43 的
   `go hf(context, user, items)` 没有顶层 recover, 任何下游 hook panic 会
   crash 整个进程.
2. web/callback_controller_handler.go 的 worker 池循环
   `controller.doCallbackLog()` 同样没有 recover, panic 会让 worker 永久
   退出, 进程也会 crash.

## Review 主要意见

### Important (本 PR 已采纳)

1. **删多余 import**: web/callback_controller_handler_test.go 中
   `var _ = context.NewRecommendContext` 仅用于消除未使用 import 报错,
   是死代码. 已删除.
2. **测试触发 panic 的方式不确定**: 原 TestRunCallbackSafely_RecoversFromPanic
   依赖 doCallbackLog 内部某条分支恰好 panic, 不够稳. 已改为通过
   RegisterCallBackProcessFunc 注入一个一定 panic 的回调, 由 SceneId 路由
   触发, panic 路径明确, 即使将来 doCallbackLog 加更多前置守卫测试依然有效.

### Suggestion (留作 follow-up)

3. **同文件其他 unprotected goroutine**:
   service/user_recommend.go:169-170 的
   `go feature_log.FeatureLog(...)` 与
   `go r.featureConsistencyJobService.LogSampleResult(...)`
   也是 user-supplied 实现, 同样会 crash 进程. 同 PR 范围之外, 建议后续
   独立 PR 处理.
4. **类型签名小不一致**: RecommendCleanHookFunc 定义用
   `params ...interface{}`, 新增的 SafeRun 用 `params ...any`. Go 1.18+
   下完全等价, 项目 Go 1.24+ 后可统一为 `any`. 留作后续清理.

## 验证

- `go test ./service/hook/... ./web/...`: 全部通过
- `go vet ./service/...`: 干净
- 已通过 review 的两条 Important 修改后重跑 OK

## Follow-up

- [ ] PR-X: 保护 service/user_recommend.go:169-170 的另两个 goroutine
- [ ] 整理 RecommendCleanHookFunc 等签名为 `any`

## 相关链接

- 主 PR: TBD (本次提交后补)
