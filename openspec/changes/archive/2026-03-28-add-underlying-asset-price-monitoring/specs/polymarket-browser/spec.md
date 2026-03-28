## MODIFIED Requirements

### Requirement: Market 实时监听页
系统 SHALL 通过 WebSocket 实时展示市场的各 outcome Bid/Ask 价格。对于有底层资产的市场（`resolutionSource` 可解析），系统 SHALL 同时展示底层资产的实时价格和起始价格（`priceToBeat`）。进入 Market 页时 SHALL 创建并启动 Watcher，离开 Market 页时 SHALL 停止 Watcher 并关闭所有 WebSocket 连接（包括 Market Channel 和 RTDS）。页面 SHALL 显示连接状态和最后更新时间。

#### Scenario: 实时价格更新
- **WHEN** 用户进入 Market 页
- **THEN** 系统创建 Watcher 并订阅该市场，实时展示各 outcome 的 Bid/Ask 价格

#### Scenario: 底层资产价格展示
- **WHEN** 市场有底层资产（resolutionSource 可解析）且 Watcher 推送 `underlying_price` 事件
- **THEN** 系统 SHALL 在页面中展示底层资产的当前实时价格，格式为 "{Asset} 价格: ${value}"（如 "BTC/USD 价格: $87,534.50"）

#### Scenario: 起始价格展示
- **WHEN** 市场有底层资产且 Watcher 推送 `price_to_beat` 事件
- **THEN** 系统 SHALL 在页面中展示起始价格，格式为 "起始价格: ${value}"（如 "起始价格: $86,321.88"）

#### Scenario: 无底层资产的市场
- **WHEN** 市场的 resolutionSource 为空或不匹配
- **THEN** 系统 SHALL 不展示底层资产价格区域，行为与改动前一致

#### Scenario: 退出 Market 页停止监听
- **WHEN** 用户在 Market 页按 Esc
- **THEN** 系统 SHALL 调用 Watcher.Stop() 停止所有 WebSocket 连接（Market Channel 和 RTDS），pop 回 Event 页
