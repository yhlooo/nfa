## Why

当前 NFA Agent 每次对话都是"无记忆"的——不记得用户之前问过什么、关注哪些股票、有什么投资偏好。Agent 无法从历史对话中学习，也无法为用户提供个性化的回答体验。引入长期记忆模块，让 Agent 能够跨会话记住用户的关键信息和偏好，逐步提供更贴合用户的金融咨询服务。

## What Changes

- 新增 `pkg/memory` 包，实现记忆的加载、保存和异步更新
- 程序启动时加载 `~/.nfa/MEMORY.md` 记忆文件，将其内容注入系统提示词
- 每轮对话结束后，启动异步总结流程：使用主模型回顾全部会话历史，结合现有记忆内容，生成更新后的记忆文件
- 采用 1 分钟防抖机制：用户停止输入 1 分钟后才触发总结，避免频繁对话时反复总结
- 防抖等待期间或异步总结期间，用户开始新对话则取消当前总结，等新对话完成后再重新计时
- 记忆内容排除时间敏感信息（如某日股价），聚焦于用户画像、偏好和可复用的知识经验

## Capabilities

### New Capabilities

- `memory-storage`: 记忆的持久化存储和读取，包括 MEMORY.md 文件的格式定义、加载和写入
- `memory-summarization`: 异步记忆总结，包括防抖等待、模型调用、取消机制和与现有记忆的合并更新

### Modified Capabilities

<!-- 无现有 spec 需求变更 -->

## Impact

- `pkg/agents/agent.go` — NFAAgent 结构体新增 memory 和 summarizer 字段
- `pkg/agents/acp.go` — Prompt() 方法中集成总结取消/启动逻辑，将 memory 注入 context
- `pkg/agents/prompts.go` — 系统提示模板和数据中增加用户记忆章节
- `pkg/agents/genkit.go` — 可能需要传递 memory 内容给系统提示函数
- `pkg/ctxutil/values.go` — 新增 memory context key
- 新增 `pkg/memory/` 包
- `~/.nfa/MEMORY.md` — 新增数据文件
