# Strategy Trading System

策略交易系统允许用户在 PolyMarket 市场上运行预定义的交易策略。

## 策略接口

策略是一个无状态、无副作用的逻辑单元，基于输入的市场和持仓信息输出交易决策。

### 接口定义

```go
// Strategy 策略接口
type Strategy interface {
    // Name 返回策略名称
    Name() string
    // Execute 执行策略，返回交易决策
    Execute(ctx context.Context, in Input) (*Result, error)
}

// Input 策略输入
type Input struct {
    // Market 市场元信息
    Market MarketInfo
    // Position 当前持仓
    Position *Position
    // Orders 当前未完成订单
    Orders []*Order
    // ServerTime PolyMarket 服务器时间（如有）
    ServerTime time.Time
    // YesBidPrices Yes 买价趋势（最多 1000 个点）
    YesBidPrices []PricePoint
    // YesAskPrices Yes 卖价趋势
    YesAskPrices []PricePoint
    // NoBidPrices No 买价趋势
    NoBidPrices []PricePoint
    // NoAskPrices No 卖价趋势
    NoAskPrices []PricePoint
    // TrackingPrice 跟踪底层资产价格（如有）
    TrackingPrice *float64
    // TrackingTarget 跟踪底层资产目标价格（如有）
    TrackingTarget *float64
}

// Result 策略输出
type Result struct {
    // Orders 新订单请求
    Orders []*OrderRequest
    // CancelIDs 要取消的订单 ID
    CancelIDs []string
}

// OrderRequest 订单请求
type OrderRequest struct {
    Side      OrderSide   // BUY / SELL
    Outcome   Outcome     // YES / NO
    Size      float64     // 数量（基础单位）
    Price     float64     // 价格
    OrderType OrderType   // LIMIT / MARKET
}
```

### 策略约束

- 策略必须是无状态且无副作用的，可以在任何时候调用
- 策略输出的订单买入数量最少应价值 1 美元
- 卖出数量没有限制

## 持仓管理

### Position

```go
// Position 持仓信息
type Position struct {
    // Cash 现金（可负）
    Cash float64
    // YesShares Yes 持仓数量
    YesShares float64
    // NoShares No 持仓数量
    NoShares float64
    // YesAvgCost Yes 平均成本
    YesAvgCost float64
    // NoAvgCost No 平均成本
    NoAvgCost float64
}
```

### 资产总价值计算

```
总价值 = 现金 + Yes持仓 * Yes当前买价 + No持仓 * No当前买价
```

- 策略开始时总价值初始为 0
- 正数表示盈利，负数表示亏损

### 交易规则

- 现金初始值为 0，可以是负数
- 买入 Yes/No 需要消耗现金并增加持仓
- 卖出 Yes/No 需减少持仓并增加现金
- Yes/No 持仓数量不可以是负数

## 订单管理

### Order

```go
// Order 订单记录
type Order struct {
    // ID 订单 ID
    ID string
    // CreatedAt 创建时间
    CreatedAt time.Time
    // Side 买卖方向
    Side OrderSide
    // Outcome YES / NO
    Outcome Outcome
    // Price 价格（限价单）
    Price float64
    // Size 数量（乘以交易倍数后的实际数量）
    Size float64
    // OrderType 订单类型
    OrderType OrderType
    // Status 订单状态
    Status OrderStatus
    // FilledAt 成交时间
    FilledAt *time.Time
    // FilledPrice 成交价格
    FilledPrice float64
}
```

### 订单状态

- `PENDING`: 等待成交
- `FILLED`: 已成交
- `CANCELLED`: 已取消

## 策略执行器

### 功能

- 在指定市场上应用策略进行交易
- 监听市场价格变化，触发策略执行
- 支持定时触发策略执行（可配置间隔）
- 支持 dry-run 模拟交易和真实交易（预留）

### 触发条件

执行器在以下情形执行策略：

1. 监听的市场价格发生变化
2. 监听的底层资产价格发生变化
3. 每间隔指定时间（默认 5 秒）

### 交易倍数

- 执行器根据配置对策略输出的订单数量乘以固定倍数
- 输入给策略的订单数量、持仓数量等都是翻倍前的原始数量
- 翻倍操作对策略透明

### dry-run 模式

- 订单仅在执行器内部记录
- 当市场价格触及限价单价格时模拟成交
- 记录盈亏等信息

## CLI 命令

### 命令格式

```bash
nfa tools polymarket trade [flags]
```

### 参数

| 参数 | 简写 | 说明 | 默认值 |
|------|------|------|--------|
| `--dry-run` | `-d` | 模拟交易模式 | true |
| `--market` | `-m` | 市场 slug（必需） | - |
| `--strategy` | `-s` | 策略名（必需） | - |
| `--multiplier` | `-x` | 交易倍数 | 1 |
| `--interval` | `-i` | 策略执行间隔 | 5s |

### 示例

```bash
# 模拟交易
nfa tools polymarket trade --dry-run -m btc-updown-15m-1774707300 -s simple -x 10

# 真实交易（预留，当前 TODO）
nfa tools polymarket trade -m btc-updown-15m-1774707300 -s simple -x 10
```

## UI 展示

UI 实时显示以下信息：

- 市场标题、slug、描述等元信息
- 当前 Yes/No 的最新 bid/ask 价格
- 跟踪的底层资产价格（如有）
- 当前持仓：Yes/No 数量、平均成本、持仓价值
- 现金、总资产价值
- 交易记录列表：时间、方向、资产类型、报价、成交价、数量

### UI 布局示例

```
┌─────────────────────────────────────────────────────────────────────┐
│  Will BTC be above $88k at 3pm ET today?                           │
│  btc-updown-15m-1774707300                                          │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  BTC Price: $87,542.00    Target: $88,000.00                        │
│                                                                     │
│  ┌──────────────────────┐    ┌──────────────────────┐              │
│  │         YES          │    │          NO          │              │
│  │   Bid: 0.45          │    │   Bid: 0.52          │              │
│  │   Ask: 0.47          │    │   Ask: 0.54          │              │
│  └──────────────────────┘    └──────────────────────┘              │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│  Portfolio                                                          │
│  Cash: $-125.50    YES: 150 shares @ $0.42    NO: 0 shares          │
│  Position Value: $67.50    Total Value: $-58.00                     │
├─────────────────────────────────────────────────────────────────────┤
│  Trade History                                                      │
│  15:04:32  BUY   YES  100 @ 0.42  (filled 0.42)                     │
│  15:03:21  SELL  YES   50 @ 0.48  (filled 0.48)                     │
│  15:02:15  BUY   YES   50 @ 0.41  (filled 0.41)                     │
├─────────────────────────────────────────────────────────────────────┤
│  ● Connected | Strategy: simple | Mode: DRY-RUN | Multiplier: 10x  │
└─────────────────────────────────────────────────────────────────────┘
```

## 示例策略

### SimpleStrategy

简单示例策略：

- 价格低于 0.2 时买入
- 持仓收益高于 0.4 时卖出
