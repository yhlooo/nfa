## 1. 基础设施

- [x] 1.1 新增 `pkg/ctxutil` memory context key（`ContextWithMemory` / `MemoryFromContext`）
- [x] 1.2 新增 `pkg/memory/memory.go`：`Memory` 结构体，实现 `Load`、`Content`、`Update` 方法，含 `sync.RWMutex` 线程安全保护

## 2. 系统提示词集成

- [x] 2.1 修改 `pkg/agents/prompts.go`：`AgentSystemPromptData` 新增 `Memory` 字段，模板新增"## 用户记忆"章节，`AnalystSystemPrompt` 从 context 读取 memory 并填入数据

## 3. 异步总结

- [x] 3.1 新增 `pkg/memory/summarizer.go`：`Summarizer` 结构体，实现 `Schedule`（1 分钟防抖等待）、`Cancel`（取消等待或进行中的总结）、`summarize`（调用主模型生成总结）方法

## 4. Agent 集成

- [x] 4.1 修改 `pkg/agents/agent.go`：`NFAAgent` 结构体新增 `memory` 和 `summarizer` 字段，`Options` 或初始化流程中加载 memory
- [x] 4.2 修改 `pkg/agents/genkit.go`：在 `InitGenkit` 中加载 `~/.nfa/MEMORY.md`，将 memory 和 model caller 注入 summarizer
- [x] 4.3 修改 `pkg/agents/acp.go`：`Prompt()` 方法入口处调用 `summarizer.Cancel()`，将 memory 内容注入 context，对话处理完成后调用 `summarizer.Schedule()`

## 5. 测试与验证

- [x] 5.1 为 `pkg/memory/memory.go` 编写单元测试
- [x] 5.2 为 `pkg/memory/summarizer.go` 编写单元测试（mock 模型调用）
- [x] 5.3 运行 `go fmt ./...`、`go vet ./...`、`go test ./...` 确保代码质量和测试通过
