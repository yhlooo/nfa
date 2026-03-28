## 1. 类型定义

- [x] 1.1 在 `pkg/polymarket/types.go` 中新增 `MarketEvent` 嵌套结构体（包含 `EventMetadata`），扩展 `Market` 结构体以支持 `events` JSON 字段
- [x] 1.2 在 `pkg/polymarket/types.go` 中新增 `EventMetadata` 结构体（包含 `PriceToBeat *float64` 字段）
- [x] 1.3 在 `pkg/polymarket/types.go` 中新增 `ResolutionSourceInfo` 结构体（Asset、Quote、Topic、Symbol 字段）和 `ParseResolutionSource` 函数
- [x] 1.4 在 `pkg/polymarket/ws_types.go` 中新增 `UnderlyingPriceEvent` 和 `PriceToBeatEvent` 结构体
- [x] 1.5 在 `pkg/polymarket/ws_types.go` 中新增 RTDS 订阅请求结构体 (`RTDSSubscription`) 和 RTDS 消息结构体 (`RTDSMessage`、`RTDSPayload`)

## 2. RTDS 客户端

- [x] 2.1 在 `pkg/polymarket/real_client.go` 中新增 `RTDSWebSocketEndpoint` 常量和 `ConnectRTDS` 方法
- [x] 2.2 为 `ParseResolutionSource` 编写单元测试，覆盖 Chainlink URL 匹配、空值、不匹配 URL 等场景

## 3. Watcher 集成

- [x] 3.1 修改 `Watcher` 结构体：新增 `market`、`underlyingAsset`、`rtdsConn`、`rtdsConnMu` 字段
- [x] 3.2 修改 `NewWatcher` 函数签名：新增 `market *Market` 参数，调用 `ParseResolutionSource` 确定底层资产信息
- [x] 3.3 修改 `Watcher.Start()`：如果有底层资产，启动 RTDS 连接、RTDS 读取循环和 priceToBeat 轮询
- [x] 3.4 实现 `Watcher.connectRTDS()`：连接 RTDS WebSocket 并发送订阅消息
- [x] 3.5 实现 `Watcher.readRTDSLoop()`：读取 RTDS 消息，解析 `underlying_price` 事件并推送到 eventCh
- [x] 3.6 实现 `Watcher.pollPriceToBeat()`：定时轮询 REST API 获取 `priceToBeat`，推送 `price_to_beat` 事件
- [x] 3.7 修改 `Watcher.heartbeatLoop()`：同时为 RTDS 连接发送 PING（间隔 5 秒）
- [x] 3.8 修改 `Watcher.Stop()` 和 `Watcher.handleDisconnect()`：正确关闭 RTDS 连接

## 4. UI 更新

- [x] 4.1 更新 `pkg/ui/polymarketwatcher/ui.go`：新增 `underlyingPrice` 和 `priceToBeat` 状态字段
- [x] 4.2 在 `handleMarketEvent` 中处理 `underlying_price` 和 `price_to_beat` 事件类型
- [x] 4.3 在 `View()` 中渲染底层资产价格区域（仅在有底层资产时显示）
- [x] 4.4 在 `pkg/ui/polymarketwatcher/i18n.go` 中新增底层资产价格相关的国际化文本

## 5. 验证

- [x] 5.1 更新所有调用 `NewWatcher` 的代码以适配新签名（`pkg/ui/polymarketwatcher/` 和 `pkg/ui/polymarket/` 中的调用点）
- [x] 5.2 运行 `go fmt ./...`、`go vet ./...`、`go test ./...` 确保代码质量
