## 1. PolyMarket API 客户端扩展

- [x] 1.1 在 `pkg/polymarket/types.go` 中新增 `Series` 结构体（id, slug, title, subtitle, description, volume, volume24hr, liquidity, active, closed, events 等字段）
- [x] 1.2 在 `pkg/polymarket/types.go` 中扩展 `Event` 结构体，添加 volume, volume24hr, liquidity, series ([]Series), markets ([]Market), startDate, endDate, active, closed, featured 等字段，所有新字段使用 `omitempty`
- [x] 1.3 在 `pkg/polymarket/types.go` 中扩展 `Market` 结构体，添加 outcomePrices, volume, liquidity, volume24hr, bestBid, bestAsk, outcomes 等字段
- [x] 1.4 在 `pkg/polymarket/types.go` 中新增请求/响应类型：`ListEventsRequest`（limit, offset, order, ascending, active, featured 等）、`ListSeriesRequest`（limit, offset, order, ascending, closed 等）、`SearchResult`（events, tags）、`SearchRequest`（query, limit）
- [x] 1.5 在 `pkg/polymarket/real_client.go` 中实现 `ListEvents` 方法（GET /events）
- [x] 1.6 在 `pkg/polymarket/real_client.go` 中实现 `ListSeries` 方法（GET /series）
- [x] 1.7 在 `pkg/polymarket/real_client.go` 中实现 `Search` 方法（GET /public-search）
- [x] 1.8 扩展 `GammaAPIClient` 接口，添加 `ListEvents`、`ListSeries`、`Search` 方法签名

## 2. TUI 框架搭建

- [x] 2.1 创建 `pkg/ui/polymarket/` 包目录结构
- [x] 2.2 实现 `page.go`：定义 `Page` 接口（包含 `tea.Model` + `Type() string` 方法）
- [x] 2.3 实现 `browser.go`：主 Browser 模型，管理页面栈（push/pop）、委托 Update/View 给栈顶 Page、处理全局 Esc/CtrlC
- [x] 2.4 实现 `styles.go`：共享 lipgloss 样式定义（标题、描述、列表项、状态栏等）
- [x] 2.5 实现 `i18n.go`：定义所有浏览器相关的 i18n Message 结构体

## 3. 首页实现

- [x] 3.1 实现 `home_page.go`：首页模型，包含搜索输入框（bubbles/textinput）和列表
- [x] 3.2 实现首页数据加载：启动时并行调用 ListSeries + ListEvents，合并去重，按 volume24hr 降序排列
- [x] 3.3 实现首页交互：Tab 切换搜索框焦点、上下键选择列表项、Enter 进入详情、Esc 退出
- [x] 3.4 实现首页搜索：Enter 触发 Search API，结果替换列表，清空恢复热门列表
- [x] 3.5 实现首页渲染：列表项类型标记 [Series]/[Event]、标题、交易量摘要

## 4. Series 详情页实现

- [x] 4.1 实现 `series_page.go`：Series 详情页模型，展示标题、描述、交易量信息
- [x] 4.2 实现事件列表渲染和导航（上下键选择、Enter 进入 Event 页）
- [x] 4.3 处理 Series 下 events 为空的边界情况

## 5. Event 详情页实现

- [x] 5.1 实现 `event_page.go`：Event 详情页模型，展示标题、描述、volume24h、liquidity
- [x] 5.2 实现市场列表渲染（含 outcomePrices 摘要）和导航（上下键选择、Enter 进入 Market 页）
- [x] 5.3 处理 Event 下 markets 为空的边界情况

## 6. Market 实时监听页实现

- [x] 6.1 实现 `market_page.go`：Market 实时监听页模型，展示市场问题和描述
- [x] 6.2 集成 Watcher：进入时创建并启动 Watcher，实时展示各 outcome 的 Bid/Ask 价格
- [x] 6.3 实现 Watcher 生命周期管理：离开 Market 页时调用 Watcher.Stop() 关闭 WebSocket 连接
- [x] 6.4 显示连接状态（Connected/Disconnected）和最后更新时间

## 7. 命令入口调整

- [x] 7.1 修改 `pkg/commands/tools_polymarket.go`：`nfa tools polymarket` 默认进入浏览器模式，`watch` 作为子命令
- [x] 7.2 确保现有 `nfa tools polymarket watch [MARKET_SLUG]` 功能不受影响

## 8. 国际化和收尾

- [x] 8.1 运行 `i18n-translate` skill 生成翻译文件
- [x] 8.2 运行 `go fmt ./...`、`go vet ./...`、`go test ./...` 确保代码质量
