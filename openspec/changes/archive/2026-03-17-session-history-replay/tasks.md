# 任务清单

## 1. 消息转换辅助函数

- [x] 1.1 在 `pkg/agents/session_store.go` 添加 `replayHistory` 函数框架
- [x] 1.2 实现用户消息回放 `replayUserMessage`
- [x] 1.3 实现模型消息回放 `replayModelMessage`（包含 Text、Reasoning、ToolRequest、ToolResponse）
- [x] 1.4 实现工具消息回放 `replayToolMessage`（如有必要）

## 2. Agent 侧集成

- [x] 2.1 在 `NFAAgent` 结构体添加 `replayHistory` 方法
- [x] 2.2 修改 `LoadSession` 方法，调用 `replayHistory`
- [x] 2.3 错误处理：发送失败时清理会话并返回错误

## 3. 测试

- [x] 3.1 添加单元测试验证消息转换逻辑
- [x] 3.2 手动测试：创建会话 → 退出 → 恢复 → 验证历史显示
- [x] 3.3 测试工具调用历史正确显示（先 in_progress 后 completed）
- [x] 3.4 测试思考过程历史正确显示
