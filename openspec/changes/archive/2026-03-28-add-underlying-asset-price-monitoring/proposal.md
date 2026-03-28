## Why

当前 Polymarket Watcher 只能监听预测市场自身的价格数据（Yes/No 的 bid/ask），无法展示市场跟踪的底层资产实时价格（如 BTC、ETH 等）。对于像 `btc-updown-5m` 这类跟踪加密货币价格的市场，用户需要同时观察底层资产价格才能做出更好的判断。同时，Up/Down 类型市场在时间窗口开始后会生成 `priceToBeat`（起始价格），这是判断市场走向的关键参考信息，但目前无法获取。

## What Changes

- **集成 RTDS WebSocket**：在现有 Watcher 中新增 Polymarket RTDS (Real Time Data Socket) 连接，用于获取底层加密资产的实时价格
- **自动资产识别**：从市场的 `resolutionSource` 字段解析跟踪的资产信息（如 `https://data.chain.link/streams/btc-usd` → BTC/USD），自动判断是否需要建立 RTDS 连接
- **priceToBeat 轮询**：对于有底层资产的市场，定时轮询 REST API 获取 `eventMetadata.priceToBeat`（起始价格），直到该值可用
- **新增事件类型**：在 `MarketEvent` 中新增 `underlying_price` 和 `price_to_beat` 两种事件类型
- **UI 展示增强**：在监听界面中展示底层资产实时价格和起始价格

## Capabilities

### New Capabilities
- `underlying-asset-price`: 支持通过 RTDS WebSocket 获取 Polymarket 市场跟踪的底层加密资产实时价格，包括 BTC/USD、ETH/USD 等

### Modified Capabilities
- `polymarket-watcher`: 扩展 Watcher，支持同时监听 Market Channel 和 RTDS 两个 WebSocket 连接，并自动从市场元信息中识别底层资产

## Impact

- **pkg/polymarket/watcher.go**：新增 RTDS 连接管理、底层资产识别逻辑、priceToBeat 轮询
- **pkg/polymarket/ws_types.go**：新增 RTDS 相关的消息类型（`UnderlyingPriceEvent`、`PriceToBeatEvent`）
- **pkg/polymarket/types.go**：扩展 `Market` 结构体以包含 `events` 嵌套数据，新增 `ResolutionSourceInfo`、`EventMetadata` 等类型
- **pkg/polymarket/real_client.go**：新增 `ConnectRTDS()` 方法
- **pkg/polymarket/common_client.go**：无变更
- **pkg/ui/polymarketwatcher/ui.go**：增强 UI 展示，新增底层资产价格和起始价格的显示
- **pkg/ui/polymarketwatcher/i18n.go**：新增国际化文本