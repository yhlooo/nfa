## Context

当前 NFA Agent 使用 ACP 协议与 TUI 交互，通过 Genkit 框架调用 LLM。系统提示词在 `pkg/agents/prompts.go` 中通过 Go 模板渲染，每轮对话时被 Genkit 发送给模型。对话历史保存在 session 中，但跨会话完全无记忆。

本次设计需要在现有架构上增加长期记忆能力，核心挑战是：
- 如何在系统提示词中注入记忆内容
- 如何在不阻塞用户对话的前提下异步总结
- 如何在用户快速连续对话时正确取消和重新调度总结

## Goals / Non-Goals

**Goals:**
- 程序启动时加载 `~/.nfa/MEMORY.md`，内容注入每轮对话的系统提示词
- 每轮对话结束后，经过 1 分钟防抖等待后，异步调用主模型对全部会话历史进行总结
- 总结时同时参考现有记忆内容，避免重复记录，更新 MEMORY.md
- 防抖等待期间或总结期间，新对话开始则取消，等新对话完成后再重新计时
- 记忆内容聚焦用户画像、偏好和可复用知识，排除时间敏感数据

**Non-Goals:**
- 不实现向量化记忆或 RAG 检索
- 不改变 session 持久化逻辑
- 不改变 ACP 协议
- 不在 TUI 中暴露记忆管理界面
- 不引入新的模型配置字段（使用现有 Primary 模型做总结）

## Decisions

### 1. 记忆内容通过 Context 传递给系统提示函数

**选择**：在 `Prompt()` 中将 memory 内容通过 `context.WithValue` 注入，系统提示函数从 context 读取。

**理由**：系统提示函数签名是 `func(context.Context, any) (string, error)`，context 是唯一的外部数据通道。不需要修改 Genkit 框架代码或变更函数签名。

**替代方案**：将 memory 引用直接闭包捕获到 `AnalystSystemPrompt` 函数中。但该函数在 `InitGenkit` 时创建，此时 memory 已加载，同样可行。选择 context 方式更解耦，后续如果 memory 在运行时更新，无需重建 flow。

### 2. 总结使用 1 分钟防抖，而非每轮立即总结

**选择**：每轮对话结束后启动 1 分钟定时器，期间有新对话则重置计时器。

**理由**：用户可能快速连续提问，每轮立即总结会造成不必要的 API 调用和 token 消耗。1 分钟是一个合理的"对话间歇"阈值。

**替代方案**：立即总结或更长的等待时间。立即总结过于激进，更长时间可能导致记忆更新延迟。

### 3. 使用 `time.Timer` + `context.WithCancel` 实现防抖和取消

**选择**：用一个 goroutine 监听 timer 和 context 取消信号，在 Prompt() 入口处取消上一个 context。

**理由**：Go 标准库即可满足，无需引入第三方库。context 取消链自然传播到模型调用（Genkit Generate 支持 context 取消）。

```
Prompt() 入口:
  ├── summarizer.Cancel()          // 取消上次的 timer 或进行中的总结
  ├── ... 对话处理 ...
  └── 对话结束:
       └── go summarizer.Schedule(history, existingMemory)
            │
            ├── select {
            │   case <-timer.C (1 min):  → 开始总结
            │   case <-ctx.Done():       → 取消，退出
            │  }
            │
            └── 总结: genkit.Generate(ctx, ...) → 写入 MEMORY.md
```

### 4. 总结模型调用使用 Genkit Generate（非 Flow）

**选择**：直接用 `genkit.Generate()` 发送总结提示词，不走 Chat Flow。

**理由**：总结是独立任务，不需要工具调用、反射循环等 Flow 机制。简单的单轮 Generate 足够，且 context 取消也能正确中断。

### 5. MEMORY.md 格式使用 Markdown 分节

**选择**：四个核心章节——用户关注、投资偏好、回答风格、经验与知识。

**理由**：Markdown 对人类可读，LLM 也能很好理解和生成。分节结构便于模型在总结时定位和更新对应章节。

### 6. 总结 Prompt 设计

**选择**：System prompt 描述总结任务和记忆原则，User prompt 包含现有记忆和对话历史。

总体原则在 system prompt 中明确：
- 提取用户画像信息（关注标的、板块、投资偏好、风险偏好、风格偏好）
- 记录可复用的知识经验（信息源、分析框架、工具使用技巧）
- 排除时间敏感数据（具体股价、日期相关的行情数据）
- 与现有记忆去重合并，保持简洁

### 7. 记忆模块包结构

```
pkg/memory/
├── memory.go       # Memory 结构体、Load、Update
└── summarizer.go   # Summarizer 结构体、Schedule、Cancel、summarize
```

两个文件，职责清晰：memory.go 处理存储，summarizer.go 处理总结逻辑。

## Risks / Trade-offs

- **总结模型调用失败** → 静默失败，记录日志。不影响主对话流程，下次对话结束后会重新触发总结。
- **MEMORY.md 文件损坏或格式异常** → Load 时容错处理，文件不存在返回空记忆，解析失败记录日志并返回空记忆。
- **总结质量不稳定** → 模型可能漏掉重要信息或引入幻觉。通过精心设计的总结 prompt 引导模型如实总结，不编造信息。
- **总结成本** → 每次总结需要将全部历史消息发送给模型。对于长对话，token 消耗可能较大。后续可考虑仅总结新增消息的增量模式。
- **并发写入** → 同一时间只有一个总结 goroutine 运行（通过 Cancel 保证）。Memory 结构体使用 RWMutex 保护读写。
