## 1. 会话存储层

- [x] 1.1 创建 `pkg/agents/session_store.go`，定义 `SessionData` 结构体和会话文件路径常量
- [x] 1.2 实现 `SaveSession` 函数：将会话历史保存到 `~/.nfa/sessions/{sessionId}/session.json`
- [x] 1.3 实现 `LoadSessionData` 函数：从文件加载会话历史
- [x] 1.4 添加单元测试验证序列化/反序列化正确性

## 2. Agent 侧实现

- [x] 2.1 在 `NFAAgent` 结构体中添加 `sessionsDir` 字段
- [x] 2.2 实现 `acp.AgentLoader` 接口的 `LoadSession` 方法
- [x] 2.3 在 `Initialize` 方法中设置 `AgentCapabilities.LoadSession = true`
- [x] 2.4 在 `Prompt` 方法成功返回前调用 `SaveSession` 保存会话
- [x] 2.5 添加集成测试验证会话保存和加载

## 3. 命令行参数

- [x] 3.1 在 `pkg/commands/root.go` 的 `Options` 结构体中添加 `Resume string` 字段
- [x] 3.2 添加 `--resume` 命令行参数绑定
- [x] 3.3 将 `Resume` 参数传递给 UI

## 4. UI 侧实现

- [x] 4.1 在 `ChatUI` 结构体中添加 `resumeSessionID` 字段
- [x] 4.2 在 `Options` 结构体中添加 `ResumeSessionID` 字段
- [x] 4.3 实现 `loadSession` 方法调用 `conn.LoadSession`
- [x] 4.4 修改 `Init` 方法：根据 `resumeSessionID` 决定调用 `newSession` 或 `loadSession`
- [x] 4.5 在退出时打印会话恢复提示信息

## 5. 测试与验证

- [x] 5.1 运行 `go fmt ./...` 格式化代码
- [x] 5.2 运行 `go vet ./...` 静态检查
- [x] 5.3 运行 `go test ./...` 单元测试
- [x] 5.4 手动测试：新会话 → 对话 → 退出 → 恢复会话
