## Why

现有长期记忆只记录可长期复用的用户偏好和知识，明确排除了时间敏感信息。但实际对话中，用户近几天的关注点、得出的结论、以及未来需要关注的事件对后续对话有重要参考价值。这些"中期"信息既不适合放在长期记忆中（会被去重/排除），也不应丢失。需要增加一个中期记忆层来填补这一空白。

## What Changes

- 新增中期记忆存储机制，每天一个文件保存在 `~/.nfa/memory/yyyy-mm-dd.md` 中
- 新增中期记忆异步总结器，在每轮对话结束后防抖等待 1 分钟后调用 LLM 总结当天对话中的时间敏感信息
- 启动时自动加载最近 7 天的中期记忆文件，注入系统提示词中独立的"近期动态"章节
- 中期记忆的加载和长期记忆一样采用 snapshot 模式，会话期间不动态刷新

## Capabilities

### New Capabilities
- `daily-memory`: 中期记忆功能，包括按天存储、异步总结、启动时加载最近 7 天记忆到系统提示词

### Modified Capabilities
<!-- 无现有规范需要修改 —— 中期记忆是新增能力，不改变现有 memory-storage 和 memory-summarization 的需求 -->

## Impact

- `pkg/memory/` — 新增 `DailyMemory` 结构体、`DailySummarizer` 总结器、总结 flow
- `pkg/agents/genkit.go` — InitGenkit 中加载最近 7 天中期记忆，初始化 DailySummarizer
- `pkg/agents/prompts.go` — 系统提示词模板新增"近期动态"章节
- `pkg/agents/agent.go` — NFAAgent 结构体新增 dailySummarizer 字段
- `pkg/agents/acp.go` — Prompt 中新增 cancel/schedule 中期记忆总结
