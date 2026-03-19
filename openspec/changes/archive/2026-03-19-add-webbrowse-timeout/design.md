# Design: WebBrowse 工具超时实现

## 技术方案

### 输入结构修改

在 `BrowseInput` 结构体中添加 `Timeout` 字段：

```go
type BrowseInput struct {
    URL      string `json:"url"`
    Question string `json:"question,omitempty"`
    Timeout  int    `json:"timeout,omitempty"` // 超时时间（秒），默认 60
}
```

### 超时实现

在工具函数入口处使用 `context.WithTimeout` 设置整体超时：

```go
func(ctx *ai.ToolContext, in BrowseInput) (BrowseOutput, error) {
    wb.lock.Lock()
    defer wb.lock.Unlock()

    // 设置整体超时
    timeout := in.Timeout
    if timeout <= 0 {
        timeout = 60
    }
    ctx2, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
    defer cancel()

    // 使用 ctx2 替代 ctx 进行后续操作
    chromeCtx, cancel := chromedp.NewContext(ctx2)
    defer cancel()

    // ... 其余实现保持不变，使用 ctx2
}
```

### 超时作用范围

```
┌─────────────────────────────────────────────────────────────┐
│                   timeout 覆盖的操作                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   1. chromedp.Navigate(url)                                 │
│   2. chromedp.Text("body", ...)                             │
│   3. chromedp.FullScreenshot(...)                           │
│   4. genkit.Generate(...) - 视觉模型调用                    │
│                                                             │
│   任一阶段超时 → 整个操作返回 context.DeadlineExceeded 错误  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 错误处理

超时错误将直接返回给 Agent，Agent 可根据需要决定是否重试或采用其他策略。

## 文件变更清单

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `pkg/tools/webbrowse/web_browse.go` | 修改 | 添加 Timeout 字段和超时逻辑 |
| `docs/guides/web-browse.md` | 修改 | 更新文档说明超时参数 |

## 风险评估

- **兼容性**：新字段为可选字段，不影响现有调用
- **性能**：无性能影响
- **安全**：超时可防止资源长时间占用
