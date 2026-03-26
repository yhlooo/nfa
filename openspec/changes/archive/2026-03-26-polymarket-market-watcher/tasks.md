# 任务：PolyMarket 市场实时监听功能

## 任务列表

### 1. 添加依赖
- [x] 添加 `github.com/gorilla/websocket` 依赖到 `go.mod`

### 2. 数据类型定义
- [x] 在 `pkg/polymarket/types.go` 中添加 `Market` 结构体
- [x] 创建 `pkg/polymarket/ws_types.go` 定义 WebSocket 消息类型
  - `SubscriptionRequest`
  - `BookEvent`
  - `PriceChangeEvent`
  - `PriceLevel`
  - `PriceChange`

### 3. Client 扩展
- [x] 在 `pkg/polymarket/types.go` 中扩展 `GammaAPIClient` 接口，添加 `GetMarketBySlug` 方法
- [x] 在 `pkg/polymarket/real_client.go` 中实现 `GetMarketBySlug` 方法
- [x] 在 `pkg/polymarket/types.go` 中添加 `CLOBWebSocketClient` 接口
- [x] 在 `pkg/polymarket/real_client.go` 中实现 `ConnectMarketWebSocket` 方法

### 4. PolyMarketWatcher 实现
- [x] 创建 `pkg/polymarket/watcher.go`
  - 实现 `Watcher` 接口
  - 实现消息读取循环
  - 实现心跳保活（每 10 秒发送 PING）
  - 实现断线重连（指数退避）
  - 定义事件类型和状态类型

### 5. UI 实现
- [x] 创建 `pkg/ui/polymarketwatcher/` 目录
- [x] 创建 `pkg/ui/polymarketwatcher/ui.go`
  - 实现 Bubbletea Model
  - 实现市场信息渲染（标题、描述）
  - 实现实时价格显示（Yes/No 的 bid/ask）
  - 实现连接状态显示
  - 实现断连状态提示
- [x] 创建 `pkg/ui/polymarketwatcher/i18n.go`
  - 定义国际化消息

### 6. CLI 命令
- [x] 在 `pkg/commands/root.go` 中注册 `tools` 子命令（如果不存在）
- [x] 创建 `pkg/commands/polymarket.go`
  - 定义 `nfa tools polymarket watch [MARKET_SLUG]` 命令
  - 解析命令参数
  - 初始化 Client、Watcher、UI
  - 启动 UI 运行循环

### 7. 国际化
- [x] 在 `pkg/commands/i18n.go` 中添加命令描述的国际化消息
- [x] 运行 `i18n-translate` skill 更新翻译文件

### 8. 测试与验证
- [x] 运行 `go fmt ./...`
- [x] 运行 `go vet ./...`
- [x] 运行 `go test ./...`
- [x] 手动测试命令功能

## 文件变更汇总

### 新增文件
| 文件路径 | 说明 |
|----------|------|
| `pkg/polymarket/ws_types.go` | WebSocket 消息类型定义 |
| `pkg/polymarket/watcher.go` | 市场监听器实现 |
| `pkg/ui/polymarketwatcher/ui.go` | TUI 实现 |
| `pkg/ui/polymarketwatcher/i18n.go` | UI 国际化消息 |
| `pkg/commands/polymarket.go` | CLI 命令实现 |

### 修改文件
| 文件路径 | 修改内容 |
|----------|----------|
| `go.mod` | 添加 gorilla/websocket 依赖 |
| `pkg/polymarket/types.go` | 添加 Market 结构体，扩展接口 |
| `pkg/polymarket/real_client.go` | 实现 GetMarketBySlug 和 ConnectMarketWebSocket |
| `pkg/commands/root.go` | 注册 tools 子命令 |
| `pkg/commands/i18n.go` | 添加国际化消息 |
| `pkg/i18n/active.en.yaml` | 英文翻译 |
| `pkg/i18n/active.zh.yaml` | 中文翻译 |
