# 设计：PolyMarket 市场实时监听功能

## 架构

### 分层架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                        │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              PolyMarketWatcherUI                     │   │
│  │  (pkg/ui/polymarketwatcher/ui.go)                   │   │
│  │  - Bubbletea Model                                   │   │
│  │  - 渲染市场信息                                       │   │
│  │  - 处理用户输入                                       │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ chan MarketEvent
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                     Business Layer                           │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              PolyMarketWatcher                       │   │
│  │  (pkg/polymarket/watcher.go)                        │   │
│  │  - WebSocket 消息读取                                │   │
│  │  - 心跳保活 (PING/PONG)                              │   │
│  │  - 断线重连                                          │   │
│  │  - 事件解析与分发                                    │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ *websocket.Conn
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                     Transport Layer                          │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                  Client                              │   │
│  │  (pkg/polymarket/real_client.go)                    │   │
│  │  - GetMarketBySlug() HTTP 请求                      │   │
│  │  - ConnectMarketWebSocket() WebSocket 连接          │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## 数据结构

### Market (API 响应)

```go
// Market 市场信息
type Market struct {
    ID            string `json:"id"`
    Question      string `json:"question"`
    Description   string `json:"description"`
    ConditionID   string `json:"conditionId"`
    Slug          string `json:"slug"`
    ClobTokenIDs  string `json:"clobTokenIds"`  // JSON 字符串数组
    Outcomes      string `json:"outcomes"`      // JSON 字符串数组
    Active        bool   `json:"active"`
    Closed        bool   `json:"closed"`
}
```

### WebSocket 消息类型

```go
// SubscriptionRequest 订阅请求
type SubscriptionRequest struct {
    AssetIDs  []string `json:"assets_ids"`
    Type      string   `json:"type"`           // "market"
    InitialDump bool   `json:"initial_dump"`   // true
    Level      int     `json:"level"`          // 2
}

// BookEvent 订单簿快照
type BookEvent struct {
    EventType string        `json:"event_type"`  // "book"
    AssetID   string        `json:"asset_id"`
    Market    string        `json:"market"`
    Bids      []PriceLevel  `json:"bids"`
    Asks      []PriceLevel  `json:"asks"`
    Timestamp string        `json:"timestamp"`
    Hash      string        `json:"hash"`
}

// PriceLevel 价格档位
type PriceLevel struct {
    Price string `json:"price"`
    Size  string `json:"size"`
}

// PriceChangeEvent 价格变化事件
type PriceChangeEvent struct {
    EventType    string         `json:"event_type"`  // "price_change"
    Market       string         `json:"market"`
    PriceChanges []PriceChange  `json:"price_changes"`
    Timestamp    string         `json:"timestamp"`
}

// PriceChange 价格变化
type PriceChange struct {
    AssetID  string `json:"asset_id"`
    Price    string `json:"price"`
    Size     string `json:"size"`
    Side     string `json:"side"`      // "BUY" / "SELL"
    Hash     string `json:"hash"`
    BestBid  string `json:"best_bid"`
    BestAsk  string `json:"best_ask"`
}
```

### UI 状态

```go
// UIState UI 状态
type UIState struct {
    // 市场信息（固定）
    MarketQuestion    string
    MarketDescription string
    OutcomeNames      []string  // ["Yes", "No"]
    AssetIDs          []string  // 对应的 token IDs

    // 实时数据
    BestBids  map[string]string  // assetID -> price
    BestAsks  map[string]string  // assetID -> price

    // 连接状态
    Connected     bool
    LastUpdate    time.Time
    Reconnecting  bool
}
```

## 关键流程

### 1. 启动流程

```
┌─────────────────────────────────────────────────────────────┐
│                       启动时序                               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  CLI Command                                                │
│       │                                                     │
│       │ 1. 解析 MARKET_SLUG                                 │
│       ▼                                                     │
│  ┌─────────┐                                                │
│  │ Client  │                                                │
│  └────┬────┘                                                │
│       │ 2. GetMarketBySlug(slug)                            │
│       │    GET /markets/slug/{slug}                         │
│       ▼                                                     │
│  ┌─────────┐                                                │
│  │ Market  │ question, description, clobTokenIds, outcomes  │
│  └────┬────┘                                                │
│       │ 3. 解析 clobTokenIds, outcomes (JSON strings)       │
│       ▼                                                     │
│  ┌─────────────────┐                                        │
│  │ Client          │ 4. ConnectMarketWebSocket(assetIDs)    │
│  └────────┬────────┘    返回 *websocket.Conn               │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────┐                                        │
│  │ PolyMarketWatcher│ 5. 发送订阅请求                       │
│  └────────┬────────┘    启动读取 goroutine                  │
│           │                                                 │
│           │ 6. chan MarketEvent                             │
│           ▼                                                 │
│  ┌─────────────────┐                                        │
│  │      UI         │ 7. 渲染初始状态                        │
│  └─────────────────┘    等待事件更新                        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 2. 心跳保活

```
┌─────────────────────────────────────────────────────────────┐
│                       心跳流程                               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  PolyMarketWatcher                                          │
│       │                                                     │
│       │ 每 10 秒                                            │
│       │                                                     │
│       ▼                                                     │
│  ┌─────────────┐                                            │
│  │ 发送 "PING" │ ────────────────────> WebSocket Server     │
│  └─────────────┘                                            │
│       │                                                     │
│       │ <──────────────────── 接收 "PONG"                   │
│       │                                                     │
│       ▼                                                     │
│  ┌─────────────────┐                                        │
│  │ 重置心跳计时器  │                                        │
│  └─────────────────┘                                        │
│                                                             │
│  如果超时未收到 PONG:                                        │
│       │                                                     │
│       ▼                                                     │
│  ┌─────────────────┐                                        │
│  │ 触发重连流程    │                                        │
│  └─────────────────┘                                        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 3. 断线重连

```
┌─────────────────────────────────────────────────────────────┐
│                       重连流程                               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  检测到连接断开                                              │
│       │                                                     │
│       ▼                                                     │
│  ┌─────────────────────┐                                    │
│  │ 发送 Disconnected   │                                    │
│  │ 事件到 UI           │                                    │
│  └──────────┬──────────┘                                    │
│             │                                               │
│             ▼                                               │
│  ┌─────────────────────┐                                    │
│  │ 等待指数退避        │ 1s, 2s, 4s, 8s... 最大 30s        │
│  └──────────┬──────────┘                                    │
│             │                                               │
│             ▼                                               │
│  ┌─────────────────────┐                                    │
│  │ 重新建立 WebSocket  │                                    │
│  │ 连接                │                                    │
│  └──────────┬──────────┘                                    │
│             │                                               │
│      ┌──────┴──────┐                                        │
│      │             │                                        │
│   成功           失败                                       │
│      │             │                                        │
│      ▼             ▼                                        │
│  ┌────────┐   ┌────────┐                                    │
│  │发送订阅│   │继续重试│                                    │
│  │请求    │   │        │                                    │
│  └───┬────┘   └────────┘                                    │
│      │                                                     │
│      ▼                                                     │
│  ┌─────────────────────┐                                    │
│  │ 发送 Connected      │                                    │
│  │ 事件到 UI           │                                    │
│  └─────────────────────┘                                    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 接口设计

### Client 接口扩展

```go
// GammaAPIClient Gamma API 客户端
type GammaAPIClient interface {
    // GetEventBySlug 通过 slug 获取事件
    GetEventBySlug(ctx context.Context, req *GetEventBySlugRequest) (*Event, error)
    // GetMarketBySlug 通过 slug 获取市场
    GetMarketBySlug(ctx context.Context, slug string) (*Market, error)
}

// CLOBReaderClient CLOB 读 API 客户端
type CLOBReaderClient interface {
    // ConnectMarketWebSocket 连接市场 WebSocket
    ConnectMarketWebSocket(ctx context.Context, assetIDs []string) (*websocket.Conn, error)
}
```

### PolyMarketWatcher 接口

```go
// Watcher 市场监听器接口
type Watcher interface {
    // Start 启动监听
    Start(ctx context.Context) error
    // Stop 停止监听
    Stop() error
    // Events 返回事件通道
    Events() <-chan MarketEvent
    // ConnectionState 返回连接状态通道
    ConnectionState() <-chan ConnectionState
}

// MarketEvent 市场事件
type MarketEvent struct {
    Type      string      // "book", "price_change"
    AssetID   string
    Data      interface{} // BookEvent 或 PriceChangeEvent
    Timestamp time.Time
}

// ConnectionState 连接状态
type ConnectionState struct {
    Connected    bool
    LastUpdate   time.Time
    Reconnecting bool
}
```

## 错误处理

| 错误场景 | 处理方式 |
|----------|----------|
| 市场 slug 不存在 | 命令退出，打印错误信息 |
| WebSocket 连接失败 | 自动重试，UI 显示断连状态 |
| 消息解析失败 | 记录日志，忽略该消息 |
| 心跳超时 | 触发重连 |

## 配置

无需额外配置，所有参数通过命令行传入。
