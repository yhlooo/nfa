## 1. 添加依赖

- [x] 1.1 添加 `github.com/charmbracelet/glamour` 依赖到 go.mod

## 2. 修改 MessageViewport 结构体

- [x] 2.1 在 `MessageViewport` 结构体中添加 glamour 渲染器字段和宽度字段
- [x] 2.2 在 `NewMessageViewport()` 中初始化 glamour 渲染器（使用 `WithAutoStyle()`）
- [x] 2.3 在 `Update()` 方法中处理 `tea.WindowSizeMsg`，更新渲染器宽度

## 3. 实现 Markdown 渲染

- [x] 3.1 在 `viewMessages()` 方法中为 `MessageTypeAgent` 类型添加 glamour 渲染逻辑
- [x] 3.2 添加渲染失败的 fallback 逻辑（输出原始文本）

## 4. 代码质量检查

- [x] 4.1 运行 `go fmt ./...` 格式化代码
- [x] 4.2 运行 `go vet ./...` 静态分析检查
- [x] 4.3 运行 `go test ./...` 单元测试