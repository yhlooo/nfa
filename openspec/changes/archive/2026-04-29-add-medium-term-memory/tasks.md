## 1. 中期记忆存储层

- [x] 1.1 在 `pkg/memory/daily_memory.go` 中创建 `DailyMemory` 结构体，支持加载最近 N 天记忆文件（`~/.nfa/memory/yyyy-mm-dd.md`），线程安全
- [x] 1.2 实现 `LoadRecent(ctx, dirPath string, days int) (string, error)` 函数——从当天倒推 N 天，加载存在的文件，按日期降序拼接
- [x] 1.3 实现 `UpdateToday(ctx, dirPath string, content string) error` 函数——写入当天文件，自动创建目录

## 2. 中期记忆总结流程

- [x] 2.1 在 `pkg/memory/daily_summarize_flow.go` 中定义 `DailySummarizeFlow`——Genkit flow，调用 LLM 将当天对话历史总结为中期记忆
- [x] 2.2 编写中期记忆总结的 system prompt——聚焦时间敏感信息（今日关注、今日结论、待关注事件），明确排除长期偏好
- [x] 2.3 复用现有 `historyToText` 等辅助函数（从 `summarize_flow.go` 抽取或直接引用）

## 3. 中期记忆异步总结器

- [x] 3.1 在 `pkg/memory/daily_summarizer.go` 中创建 `DailySummarizer` 结构体——与现有 `Summarizer` 相同的 1 分钟防抖 + 取消模式
- [x] 3.2 实现 `Schedule` 方法——1 分钟防抖后调用 `DailySummarizeFlow`
- [x] 3.3 实现 `Cancel` 方法——取消等待或进行中的总结

## 4. Agent 集成

- [x] 4.1 在 `pkg/agents/agent.go` 中为 `NFAAgent` 添加 `dailySummarizer *memory.DailySummarizer` 字段
- [x] 4.2 在 `pkg/agents/genkit.go` 的 `InitGenkit` 中加载最近 7 天中期记忆，创建 `DailySummarizer`
- [x] 4.3 将中期记忆内容传入 `AnalystSystemPrompt`（需修改函数签名添加 `dailyMemoryContent` 参数）
- [x] 4.4 在 `pkg/agents/prompts.go` 的系统提示词模板中新增"近期动态"章节
- [x] 4.5 在 `pkg/agents/prompts.go` 的 `AgentSystemPromptData` 结构体中新增 `DailyMemory` 字段
- [x] 4.6 在 `pkg/agents/acp.go` 的 `Prompt` 方法中添加中期记忆总结的 Cancel 和 Schedule 调用（与长期记忆并列）

## 5. 测试

- [x] 5.1 在 `pkg/memory/daily_memory_test.go` 中编写 `DailyMemory` 的单元测试——加载最近 N 天、当天更新、并发安全、目录不存在时自动创建
- [x] 5.2 在 `pkg/memory/daily_summarizer_test.go` 中编写 `DailySummarizer` 的单元测试——防抖取消、总结成功、总结失败
- [x] 5.3 运行 `go vet ./...` 和 `go test ./...` 确保无回归
