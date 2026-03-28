## Why

当前 `nfa tools polymarket` 仅支持通过 slug 监听单个市场的实时价格（`watch` 子命令），用户无法浏览 PolyMarket 平台上的热门事件和系列、搜索感兴趣的主题、或通过多级导航发现新市场。需要一个交互式 TUI 浏览器，让用户在终端内完整浏览和探索 PolyMarket 内容。

## What Changes

- **新增交互式 PolyMarket 浏览器**：`nfa tools polymarket` 默认进入浏览模式，采用页面栈导航（首页 → Series 页 → Event 页 → Market 实时监听页）
- **新增首页**：自动加载热门 Series 和 Event 混合列表，支持搜索（Enter 触发），按 volume24hr 降序排列
- **新增 Series 详情页**：展示系列描述和关联事件列表
- **新增 Event 详情页**：展示事件描述、volume/liquidity 摘要和关联市场列表
- **新增 Market 实时监听页**：基于 WebSocket 实时展示各 outcome 的 Bid/Ask 价格，复用现有 Watcher
- **扩展 PolyMarket API 客户端**：新增 `ListEvents`、`ListSeries`、`Search` 方法，扩展 `Event`、`Market` 结构体并新增 `Series` 结构体
- **命令入口调整**：现有 `watch` 功能改为 `nfa tools polymarket watch [MARKET_SLUG]` 子命令
- **新增 `pkg/ui/polymarket/` 包**：独立于现有 `polymarketwatcher`，使用页面栈模式管理多页面导航

## Capabilities

### New Capabilities
- `polymarket-browser`: 交互式 PolyMarket 浏览器 TUI，包含多页面导航、搜索、实时价格监听功能

### Modified Capabilities
<!-- 无需修改现有 capability 的 spec 级别行为 -->

## Impact

- **新增包**：`pkg/ui/polymarket/`（完整新 TUI 包）
- **修改包**：
  - `pkg/polymarket/`（扩展 API 客户端和数据模型）
  - `pkg/commands/`（调整 polymarket 命令入口结构）
- **新增依赖**：可能需要 `github.com/charmbracelet/bubbles`（bubbletea 生态的 textinput 组件）
- **外部 API 依赖**：PolyMarket Gamma API（`/events`、`/series`、`/public-search`）均为公开端点，无需认证
