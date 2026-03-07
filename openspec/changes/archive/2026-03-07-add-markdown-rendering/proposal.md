## Why

当前 Agent 输出的 Markdown 内容（如标题、列表、代码块、加粗等）直接以原始文本显示，缺乏格式化渲染，导致用户阅读体验不佳。需要实时渲染 Markdown 以提升可读性。

## What Changes

- Agent 消息将使用 glamour 库进行 Markdown 渲染
- 支持终端深/浅色主题自动适配
- 使用终端实际宽度进行换行
- 流式输出时实时渲染（每个 chunk 都重新渲染累积消息）

## Capabilities

### New Capabilities

- `markdown-rendering`: 在 TUI 中渲染 Agent 输出的 Markdown 内容，支持实时流式渲染和终端主题适配

### Modified Capabilities

无

## Impact

- **依赖**: 新增 `github.com/charmbracelet/glamour` 依赖
- **代码变更**:
  - `pkg/ui/chat/message_viewport.go` - 添加 glamour 渲染逻辑
- **用户体验**: Agent 输出的 Markdown 内容将呈现格式化效果（标题、列表、代码块、加粗等）