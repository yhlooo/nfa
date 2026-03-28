## Context

当前 Polymarket Watcher 通过 Market Channel WebSocket (`wss://ws-subscriptions-clob.polymarket.com/ws/market`) 获取预测市场的实时数据（订单簿、价格变化、最近成交价）。对于跟踪加密货币价格的市场（如 `btc-updown-5m`），用户需要同时观察底层资产的实时价格才能做出更好的交易判断。

Polymarket 提供了独立的 RTDS (Real Time Data Socket) 服务 (`wss://ws-live-data.polymarket.com`)，支持订阅加密资产实时价格。部分市场的 `resolutionSource` 字段包含 Chainlink 数据流 URL（如 `https://data.chain.link/streams/btc-usd`），可以从中解析出跟踪的资产信息。Up/Down 类型市场在时间窗口开始后，`eventMetadata.priceToBeat` 会提供起始价格。

## Goals / Non-Goals

**Goals:**
- 在现有 Watcher 中集成 RTDS WebSocket 连接，获取底层加密资产实时价格
- 从 `resolutionSource` 字段自动识别市场跟踪的资产，无需用户手动配置
- 定时轮询 `priceToBeat` 直到可用，提供完整的参考价格信息
- UI 层同时展示底层资产价格和市场预测价格

**Non-Goals:**
- 不从非结构化文本（`question`、`description`）中解析资产信息或目标价格
- 不支持 `resolutionSource` 为空的市场（如 "Bitcoin above $X" 类型），这些市场无法自动识别底层资产
- 不实现 RTDS 连接的跨 Watcher 复用（每个 Watcher 独立管理自己的 RTDS 连接）
- 不修改 Market Channel WebSocket 的现有行为

## Decisions

### 1. 资产识别：仅从 resolutionSource 解析

**决定**：只从 `Market` 的 `events[].resolutionSource` 字段解析底层资产信息。

**模式**：`https://data.chain.link/streams/{asset}-{quote}`（如 `btc-usd`、`eth-usd`）

**映射**：
- `btc-usd` → topic: `crypto_prices_chainlink`, symbol: `btc/usd`
- `eth-usd` → topic: `crypto_prices_chainlink`, symbol: `eth/usd`
- 其他 `{asset}-{quote}` → topic: `crypto_prices_chainlink`, symbol: `{asset}/{quote}`

**备选方案**：从 `description` 文本中用正则提取资产符号。拒绝原因：非结构化数据不稳定，维护成本高。

### 2. RTDS 连接生命周期与 Market Channel 绑定

**决定**：RTDS WebSocket 与 Market Channel WebSocket 在同一个 Watcher 中管理，一起创建、一起销毁。

**实现**：
- `Watcher.Start()` 时，如果检测到底层资产，同时建立 RTDS 连接
- `Watcher.Stop()` 时，同时关闭两个连接
- 两个连接共享同一个 `stopCh` 和 `context`

### 3. priceToBeat 通过定时轮询获取

**决定**：启动 Watcher 后，以固定间隔轮询 `GET /markets/slug/{slug}` API，从返回的 `events[].eventMetadata.priceToBeat` 中提取起始价格。

**策略**：
- 轮询间隔：每 10 秒一次
- 获取到 `priceToBeat` 后立即停止轮询并发送事件
- 如果市场已关闭或 `active=false`，也停止轮询
- 轮询使用已有的 `Client.GetMarketBySlug()` 方法

### 4. 新增事件类型通过现有 eventCh 传递

**决定**：`underlying_price` 和 `price_to_beat` 事件通过现有的 `eventCh` 传递，不新增通道。

**理由**：保持 Watcher 对外接口的一致性，消费者只需要处理新的事件类型。

### 5. Watcher 构造函数扩展

**决定**：`NewWatcher` 新增 `market *Market` 参数，用于提取 `resolutionSource` 和 `slug` 信息。

**备选方案**：将解析逻辑放在 `Start()` 中。拒绝原因：构造时就能确定是否有底层资产，避免不必要的资源分配。

## Risks / Trade-offs

- **[resolutionSource 为空的市场无法支持]** → 可接受。这些市场无法通过结构化字段识别底层资产，功能自动降级为仅 Market Channel 监听。未来可扩展其他识别策略。
- **[RTDS 连接失败不影响主功能]** → RTDS 连接失败时仅记录日志，不触发 Watcher 重连逻辑，Market Channel 正常工作。
- **[priceToBeat 轮询产生额外 API 调用]** → 最多每 10 秒一次，获取到后立即停止。对于短期市场（如 5 分钟），最多轮询 30 次。
- **[Market 结构体需要包含 events 嵌套数据]** → 需要扩展 `Market` 的 JSON 反序列化以支持 `events` 字段，该字段包含 `eventMetadata`。
