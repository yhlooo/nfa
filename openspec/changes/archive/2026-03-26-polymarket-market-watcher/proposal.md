# 提案：PolyMarket 市场实时监听功能

## 概述

添加 `nfa tools polymarket watch [MARKET_SLUG]` 命令，通过 WebSocket 实时监听指定 PolyMarket 市场并显示市场数据。

## 动机

用户希望能够实时监控 PolyMarket 市场的买卖价格变化，以便做出交易决策。目前只能通过网页手动刷新查看，缺乏命令行工具支持。

## 范围

### 包含
- CLI 子命令 `nfa tools polymarket watch [MARKET_SLUG]`
- 通过市场 slug 获取市场信息（标题、描述、token IDs）
- WebSocket 连接到 PolyMarket Market Channel
- 实时显示 Yes/No 的买价（bid）和卖价（ask）
- 断线自动重连，UI 显示连接状态

### 不包含
- 多市场同时监听
- 订单簿深度显示
- 历史价格图表
- 交易功能

## 设计概要

```
┌─────────────────────────────────────────────────────────────────┐
│                          数据流                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  用户输入 MARKET_SLUG                                           │
│         │                                                       │
│         ▼                                                       │
│  ┌─────────────┐    GET /markets/slug/{slug}                   │
│  │   Client    │ ───────────────────────────> Gamma API        │
│  │             │ <──────────────────────────                    │
│  │             │       Market{question, description,            │
│  │             │              clobTokenIds, outcomes}          │
│  └──────┬──────┘                                                │
│         │                                                       │
│         │ ConnectMarketWebSocket(clobTokenIds)                  │
│         ▼                                                       │
│  ┌─────────────────┐    wss://ws-subscriptions-clob...         │
│  │ PolyMarketWatcher│ ─────────────────────────> WebSocket     │
│  │                 │ <────────────────────────                  │
│  │                 │       Market Events (book, price_change)  │
│  └──────┬──────────┘                                            │
│         │ chan MarketEvent                                      │
│         ▼                                                       │
│  ┌─────────────────┐                                            │
│  │       UI        │    Bubbletea TUI                           │
│  │ PolyMarketWatcher│                                          │
│  └─────────────────┘                                            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 组件职责

| 组件 | 位置 | 职责 |
|------|------|------|
| CLI Command | `pkg/commands/polymarket.go` | 解析命令参数，初始化组件 |
| Client | `pkg/polymarket/real_client.go` | HTTP API 请求、WebSocket 连接建立 |
| PolyMarketWatcher | `pkg/polymarket/watcher.go` | WebSocket 消息处理、心跳、重连 |
| UI | `pkg/ui/polymarketwatcher/` | Bubbletea 界面渲染 |

## UI 设计

```
┌─────────────────────────────────────────────────────────────────┐
│ Will Tesla win the 2026 presidential election?                  │
│ This market will resolve to "Yes" if Tesla...                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   ┌─────────────────────┐   ┌─────────────────────┐           │
│   │        Yes          │   │         No          │           │
│   ├─────────────────────┤   ├─────────────────────┤           │
│   │  Bid: 0.45          │   │  Bid: 0.52          │           │
│   │  Ask: 0.47          │   │  Ask: 0.54          │           │
│   └─────────────────────┘   └─────────────────────┘           │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│ ● Connected | Press Ctrl+C to exit                              │
└─────────────────────────────────────────────────────────────────┘
```

断连状态：
```
│ ⚠ Disconnected (reconnecting...) | Last update: 10:30:45       │
```

## 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| WebSocket 连接不稳定 | 数据中断 | 自动重连机制 + UI 状态提示 |
| API 响应格式变化 | 解析失败 | 宽松解析，缺失字段显示占位符 |
| 网络延迟 | 数据滞后 | 显示最后更新时间 |

## 依赖

- 新增依赖: `github.com/gorilla/websocket` (WebSocket 客户端)
- 现有依赖: `github.com/charmbracelet/bubbletea` (TUI)

## 相关文档

- PolyMarket Market Channel API: `.temp/polymarket_market_channel_api.md`
- PolyMarket Get Market by Slug API: `.temp/polymarket_get_market_by_slug.md`
