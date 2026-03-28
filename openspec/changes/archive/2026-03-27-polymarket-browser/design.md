## Context

当前 `pkg/polymarket/` 已实现基础的 REST API 客户端（`GetEventBySlug`、`GetMarketBySlug`）和 WebSocket 实时价格监听（`Watcher`）。`pkg/ui/polymarketwatcher/` 提供了单市场实时价格展示的 TUI。`pkg/commands/tools_polymarket.go` 注册了 `nfa tools polymarket watch [MARKET_SLUG]` 命令。

PolyMarket Gamma API 提供了公开的列表端点（`GET /events`、`GET /series`、`GET /public-search`），无需认证即可访问。

## Goals / Non-Goals

**Goals:**
- 提供完整的 PolyMarket 浏览体验：首页浏览 → Series/Event 详情 → Market 实时监听
- 页面栈导航，Esc 返回上一级且保持状态
- 首页混合展示热门 Series 和 Event，支持搜索
- Market 页实时 WebSocket 价格更新

**Non-Goals:**
- 不实现交易功能（下单、撤单等）
- 不实现用户认证相关功能
- 不实现分页加载更多（首页使用固定 limit，后续可扩展）
- 不实现收藏、历史记录等个性化功能

## Decisions

### 1. 页面栈导航模式

**选择**：自定义页面栈（Browser 模型管理 `[]Page` 切片）

**替代方案**：
- 状态机（单状态切换）：无法保持多层页面状态
- bubbletea 的 viewport/list 组合：不适用于多级页面导航

**理由**：页面栈天然适合层级导航，push/pop 语义清晰，每层页面状态独立。Browser 作为 `tea.Model` 的唯一实现，将 Update/View 委托给栈顶 Page。

```
type Page interface {
    tea.Model
    Type() string  // "home", "series", "event", "market"
}
```

### 2. 首页数据源策略

**选择**：并行调用 `GET /series` + `GET /events`，客户端合并去重

**理由**：PolyMarket 没有单一的混合列表端点。`GET /events` 返回事件粒度数据，series 下的多个事件会重复出现；`GET /series` 只返回系列。并行调用后去重（移除已属于已展示 series 的事件），按 `volume24hr` 降序排列。

**替代方案**：仅用 `GET /events` 从事件的 `series` 字段提取系列信息 — 但无法获取系列完整事件列表，且一个系列下多个事件会导致重复。

### 3. 搜索交互模式

**选择**：Tab 切换焦点到搜索框，Enter 触发搜索

**理由**：默认光标在列表上，可以直接浏览，无需每次 focus 搜索框。`GET /public-search?q=xxx` 返回 events 数组，从每个 event 的 `series` 字段提取 series 信息进行展示。搜索结果替换当前列表，清空搜索框后恢复默认热门列表。

### 4. UI 包结构

**选择**：新建 `pkg/ui/polymarket/` 独立包

**理由**：现有 `pkg/ui/polymarketwatcher/` 只处理单市场监听，功能简单。新的浏览器 UI 涉及页面栈、搜索、多级导航等复杂逻辑，放在独立包中职责更清晰。Market 页会复用 `pkg/polymarket/watcher.go` 的 `Watcher`，但不复用 `polymarketwatcher` 的 UI 代码。

### 5. API 客户端扩展方式

**选择**：在现有 `real_client.go` 和 `types.go` 中扩展

**理由**：保持统一的客户端入口。新增 `ListEvents`、`ListSeries`、`Search` 方法到 `Client`，扩展 `GammaAPIClient` 接口。`Event`、`Market` 结构体添加必要字段（使用 `omitempty` 避免破坏现有序列化），新增 `Series` 结构体。

### 6. 依赖选择

**选择**：引入 `github.com/charmbracelet/bubbles` 提供 `textinput` 组件

**理由**：bubbletea 官方组件库，项目已使用 bubbletea，引入 bubbles 是标准做法。仅需使用其中的 `textinput` 组件实现搜索输入框。

## Risks / Trade-offs

- **[API 稳定性]** PolyMarket Gamma API 是非官方公开接口，可能随时变更 → 在 `types.go` 中使用 `omitempty` 标签，对未知字段保持宽容
- **[首页数据量]** 并行两次 API 调用可能造成首页加载延迟 → 使用 `limit=20` 控制数据量，loading 状态显示占位信息
- **[搜索结果去重]** `GET /public-search` 返回 events 数组，同一 series 下多个事件需要客户端去重 → 按 series slug 分组，仅展示第一个事件
- **[Market 页 Watcher 生命周期]** 进入 Market 页创建 Watcher，Esc 返回时需确保停止 → 利用 `Watcher.Stop()` 和 `sync.WaitGroup` 保证 goroutine 清理，在 Page 的 `OnUnmount` 生命周期方法中处理
