## Context

当前 `MessageViewport` 组件负责渲染聊天消息，Agent 消息直接以原始文本输出，未进行 Markdown 格式化。需要集成 glamour 库实现实时 Markdown 渲染。

**当前代码结构**:
- `pkg/ui/chat/message_viewport.go` - 消息视窗组件
- `viewMessages()` 方法处理不同消息类型的渲染
- `MessageTypeAgent` 类型直接输出原始文本
- 窗口宽度通过 `tea.WindowSizeMsg` 事件传递给 `viewStyle`

## Goals / Non-Goals

**Goals:**
- Agent 消息使用 glamour 渲染 Markdown
- 支持终端深/浅色主题自动检测
- 流式输出时实时渲染（每个 chunk 重新渲染累积消息）

**Non-Goals:**
- 不渲染用户消息的 Markdown
- 不自定义 glamour 样式（使用默认样式）
- 不处理增量渲染优化
- 不显式设置宽度（glamour 自动检测终端宽度）

## Decisions

### 1. 渲染器初始化位置

**决定**: 在 `MessageViewport` 结构体中添加 `mdRenderer` 字段

**理由**:
- 渲染器是有状态对象，需要复用以保持一致性
- 在 `NewMessageViewport()` 中初始化，与组件生命周期一致
- 使用 `glamour.WithAutoStyle()` 自动检测终端主题

**实现方式**:
```go
func NewMessageViewport(ctx context.Context) (MessageViewport, error) {
    r, err := glamour.NewTermRenderer(glamour.WithAutoStyle())
    if err != nil {
        return MessageViewport{}, fmt.Errorf("new markdown renderer error: %w", err)
    }
    return MessageViewport{
        viewStyle:  lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderLeft(true),
        mdRenderer: r,
        logger:     logr.FromContextOrDiscard(ctx),
    }, nil
}
```

**替代方案**:
- 每次渲染创建新渲染器 → 性能开销大
- 全局单例渲染器 → 生命周期管理复杂

### 2. 宽度管理

**决定**: 不显式设置宽度，使用 glamour 默认行为

**理由**:
- glamour 默认会自动检测终端宽度进行换行
- 无需在窗口大小变化时重新创建渲染器
- 实现更简洁

### 3. 渲染时机

**决定**: 每个 chunk 都重新渲染整个累积消息

**理由**:
- glamour 不支持增量渲染
- 重新渲染开销可接受（消息通常不长）
- 实现简单可靠

**替代方案**:
- 缓冲后渲染 → 增加延迟感
- 增量渲染 → glamour 不支持

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| 不完整的 Markdown 显示原始字符（如 `**重` 显示星号） | 可接受，用户理解消息还在输出中 |
| 渲染失败时无输出 | 添加 fallback，失败时输出原始文本 |
| 大量文本渲染性能问题 | 暂不处理，按需优化 |