## Purpose

提供一个交互式的 PolyMarket 浏览器，让用户能够浏览热门市场、搜索事件、查看市场详情并实时监听价格变化。
## Requirements
### Requirement: 命令入口
系统 SHALL 在执行 `nfa tools polymarket`（无子命令）时直接进入交互式 PolyMarket 浏览器模式。现有的 `watch` 功能 SHALL 作为子命令 `nfa tools polymarket watch [MARKET_SLUG]` 继续可用。

#### Scenario: 直接进入浏览模式
- **WHEN** 用户执行 `nfa tools polymarket`（不带子命令）
- **THEN** 系统启动交互式 TUI 浏览器，显示首页

#### Scenario: 通过子命令监听市场
- **WHEN** 用户执行 `nfa tools polymarket watch [MARKET_SLUG]`
- **THEN** 系统行为与现有 watch 功能一致

### Requirement: 首页热门列表
系统 SHALL 在首页自动加载热门 Series 和 Event 的混合列表。列表 SHALL 按 `volume24hr` 降序排列。每个列表项 SHALL 标记其类型为 `[Series]` 或 `[Event]`，并显示标题和交易量摘要。

#### Scenario: 首页自动加载
- **WHEN** 浏览器启动
- **THEN** 系统并行请求 `GET /series?closed=false` 和 `GET /events?active=true`，合并去重后按 volume24hr 降序展示混合列表

#### Scenario: 列表项类型标记
- **WHEN** 首页列表渲染完成
- **THEN** 属于 Series 的项 SHALL 标记为 `[Series]`，独立事件 SHALL 标记为 `[Event]`

### Requirement: 首页搜索
系统 SHALL 在首页提供搜索输入框。用户 SHALL 通过 Tab 键切换焦点到搜索框。在搜索框中输入并按 Enter 后 SHALL 触发搜索（调用 `GET /public-search`），搜索结果 SHALL 替换当前列表显示。搜索结果 SHALL 只展示 events 和 series，不展示 profiles 或 tags。清空搜索框后 SHALL 恢复默认热门列表。

#### Scenario: 切换焦点到搜索框
- **WHEN** 用户在首页按 Tab 键
- **THEN** 焦点 SHALL 切换到搜索输入框

#### Scenario: 执行搜索
- **WHEN** 用户在搜索框输入关键词并按 Enter
- **THEN** 系统调用 `GET /public-search?q=<关键词>`，搜索结果替换当前列表

#### Scenario: 清空搜索恢复热门列表
- **WHEN** 用户清空搜索框内容并按 Enter
- **THEN** 系统恢复显示默认热门列表

### Requirement: 首页导航
用户 SHALL 使用上下键在列表中选择条目，按 Enter 进入对应的详情页。Series 条目 SHALL 进入 Series 页，Event 条目 SHALL 进入 Event 页。

#### Scenario: 选择并进入 Series
- **WHEN** 用户选中一个 `[Series]` 类型条目并按 Enter
- **THEN** 系统 push Series 页到页面栈，展示该系列的详情

#### Scenario: 选择并进入 Event
- **WHEN** 用户选中一个 `[Event]` 类型条目并按 Enter
- **THEN** 系统 push Event 页到页面栈，展示该事件的详情

### Requirement: Series 详情页
系统 SHALL 展示系列的标题、描述和交易量等信息，并列出该系列下的所有事件。用户 SHALL 使用上下键选择事件，按 Enter 进入事件详情页。

#### Scenario: 展示 Series 详情
- **WHEN** 用户从首页进入一个 Series
- **THEN** 系统展示 Series 标题、描述、交易量信息，以及该系列下的事件列表

#### Scenario: 从 Series 页进入 Event
- **WHEN** 用户在 Series 页选中一个事件并按 Enter
- **THEN** 系统 push Event 页到页面栈

### Requirement: Event 详情页
系统 SHALL 展示事件的标题、描述、volume24h、liquidity 等信息，并列出该事件下的所有市场及其价格摘要。用户 SHALL 使用上下键选择市场，按 Enter 进入市场实时监听页。

#### Scenario: 展示 Event 详情
- **WHEN** 用户从首页或 Series 页进入一个 Event
- **THEN** 系统展示 Event 标题、描述、volume24h、liquidity，以及该事件下的市场列表

#### Scenario: 从 Event 页进入 Market
- **WHEN** 用户在 Event 页选中一个市场并按 Enter
- **THEN** 系统 push Market 页到页面栈，启动 WebSocket 实时监听

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

### Requirement: 页面栈导航
系统 SHALL 使用页面栈管理多页面导航。每层页面的状态（光标位置、滚动位置等）SHALL 在 push/pop 时独立保持。在首页按 Esc SHALL 退出程序，在其他页面按 Esc SHALL 返回上一级页面。

#### Scenario: Esc 返回上一级
- **WHEN** 用户在非首页页面按 Esc
- **THEN** 系统 pop 当前页面，显示上一级页面，且上一级页面状态（光标位置等）保持不变

#### Scenario: Esc 在首页退出
- **WHEN** 用户在首页按 Esc
- **THEN** 系统退出浏览器程序

#### Scenario: 状态保持
- **WHEN** 用户从 A 页进入 B 页后按 Esc 返回
- **THEN** A 页的光标位置、列表内容等状态 SHALL 与离开时完全一致

### Requirement: PolyMarket API 客户端扩展
系统 SHALL 扩展 PolyMarket 客户端，新增 `ListEvents`、`ListSeries`、`Search` 方法。`Event` 和 `Market` 结构体 SHALL 扩展必要字段（volume, liquidity, volume24hr, bestBid, bestAsk, outcomePrices 等），新增 `Series` 结构体。所有新增字段 SHALL 使用 `omitempty` 标签以保持向后兼容。

#### Scenario: 列出事件
- **WHEN** 调用 `ListEvents` 方法
- **THEN** 系统向 Gamma API `GET /events` 发送请求，返回 Event 列表，支持 limit、offset、order、active、featured 等过滤参数

#### Scenario: 列出系列
- **WHEN** 调用 `ListSeries` 方法
- **THEN** 系统向 Gamma API `GET /series` 发送请求，返回 Series 列表，支持 limit、offset、order、closed 等过滤参数

#### Scenario: 搜索
- **WHEN** 调用 `Search` 方法
- **THEN** 系统向 Gamma API `GET /public-search` 发送请求，返回包含 events、tags、profiles 的搜索结果

### Requirement: 国际化
浏览器 UI 中所有用户可见文本 SHALL 通过 i18n 机制支持中英文。

#### Scenario: 中文界面
- **WHEN** 系统语言设置为中文
- **THEN** 浏览器所有界面文本 SHALL 以中文显示

#### Scenario: 英文界面
- **WHEN** 系统语言设置为英文
- **THEN** 浏览器所有界面文本 SHALL 以英文显示

