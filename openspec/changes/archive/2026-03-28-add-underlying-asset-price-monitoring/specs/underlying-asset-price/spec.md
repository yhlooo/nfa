## ADDED Requirements

### Requirement: resolutionSource 资产解析
系统 SHALL 从市场的 `events[].resolutionSource` 字段解析底层资产信息。支持的格式为 `https://data.chain.link/streams/{asset}-{quote}`，其中 `{asset}` 为资产符号（如 `btc`、`eth`），`{quote}` 为计价货币（如 `usd`）。解析结果 SHALL 映射到 RTDS 订阅参数：topic 为 `crypto_prices_chainlink`，symbol 为 `{asset}/{quote}` 格式。当 `resolutionSource` 为空或不匹配已知模式时，系统 SHALL 视为无底层资产。

#### Scenario: 解析 Chainlink resolutionSource
- **WHEN** 市场的 `resolutionSource` 为 `https://data.chain.link/streams/btc-usd`
- **THEN** 系统 SHALL 解析出 asset=`btc`、quote=`usd`，映射为 topic=`crypto_prices_chainlink`、symbol=`btc/usd`

#### Scenario: 解析 ETH 资产
- **WHEN** 市场的 `resolutionSource` 为 `https://data.chain.link/streams/eth-usd`
- **THEN** 系统 SHALL 解析出 asset=`eth`、quote=`usd`，映射为 topic=`crypto_prices_chainlink`、symbol=`eth/usd`

#### Scenario: resolutionSource 为空
- **WHEN** 市场的 `resolutionSource` 为空字符串
- **THEN** 系统 SHALL 判定为无底层资产，不建立 RTDS 连接

#### Scenario: resolutionSource 不匹配
- **WHEN** 市场的 `resolutionSource` 为 `https://example.com/other-source`
- **THEN** 系统 SHALL 判定为无底层资产，不建立 RTDS 连接

### Requirement: RTDS WebSocket 连接
系统 SHALL 支持 Polymarket RTDS WebSocket 连接到 `wss://ws-live-data.polymarket.com`。连接建立后 SHALL 发送 JSON 格式的订阅消息：`{"topic": "<topic>", "symbols": ["<symbol>"]}`。系统 SHALL 每 5 秒发送 `PING` 文本消息保持连接活跃。

#### Scenario: 建立 RTDS 连接并订阅
- **WHEN** Watcher 启动且检测到底层资产
- **THEN** 系统 SHALL 连接到 `wss://ws-live-data.polymarket.com`，并发送订阅消息 `{"topic": "crypto_prices_chainlink", "symbols": ["btc/usd"]}`

#### Scenario: RTDS 心跳
- **WHEN** RTDS 连接已建立
- **THEN** 系统 SHALL 每 5 秒发送 `PING` 文本消息

#### Scenario: RTDS 连接失败不阻塞主功能
- **WHEN** RTDS 连接建立失败
- **THEN** 系统 SHALL 记录错误日志，Market Channel WebSocket 正常工作，不触发 Watcher 重连

### Requirement: 底层资产实时价格推送
系统 SHALL 解析 RTDS 推送的价格更新消息（格式：`{"topic":"crypto_prices_chainlink","type":"update","timestamp":...,"payload":{"symbol":"btc/usd","timestamp":...,"value":87534.50}}`），并通过 `MarketEvent`（`Type` 为 `"underlying_price"`）推送到底层资产价格。

#### Scenario: 接收并推送价格更新
- **WHEN** RTDS 推送 `{"topic":"crypto_prices_chainlink","type":"update","payload":{"symbol":"btc/usd","value":87534.50}}` 消息
- **THEN** 系统 SHALL 向 eventCh 发送 `MarketEvent{Type: "underlying_price", Data: &UnderlyingPriceEvent{Symbol: "btc/usd", Value: 87534.50}}`

#### Scenario: 忽略非 update 类型消息
- **WHEN** RTDS 推送 `{"type":"subscribe_ack"}` 等非 `update` 类型消息
- **THEN** 系统 SHALL 忽略该消息

### Requirement: priceToBeat 定时轮询
系统 SHALL 在 Watcher 启动后，对有底层资产的市场以 10 秒间隔轮询 `GET /markets/slug/{slug}` API，从返回的 `events[].eventMetadata.priceToBeat` 中提取起始价格。获取到 priceToBeat 后 SHALL 立即停止轮询，并通过 `MarketEvent`（`Type` 为 `"price_to_beat"`）推送。当市场关闭时 SHALL 停止轮询。

#### Scenario: 轮询获取到 priceToBeat
- **WHEN** 轮询返回 `eventMetadata.priceToBeat` 为 `66394.37`
- **THEN** 系统 SHALL 停止轮询，向 eventCh 发送 `MarketEvent{Type: "price_to_beat", Data: &PriceToBeatEvent{PriceToBeat: 66394.37}}`

#### Scenario: priceToBeat 尚未可用
- **WHEN** 轮询返回 `eventMetadata` 为空对象 `{}`
- **THEN** 系统 SHALL 继续下一次轮询

#### Scenario: 市场已关闭停止轮询
- **WHEN** 轮询返回市场 `closed=true`
- **THEN** 系统 SHALL 停止轮询

#### Scenario: Watcher 停止时终止轮询
- **WHEN** Watcher.Stop() 被调用
- **THEN** 系统 SHALL 立即终止轮询 goroutine
