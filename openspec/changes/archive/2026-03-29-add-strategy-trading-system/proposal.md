## Why

当前 nfa 项目已有 PolyMarket 市场浏览和价格监听功能，但缺少自动化的策略交易能力。用户需要手动监控价格并做出交易决策。策略交易系统将允许用户在 PolyMarket 市场上运行预定义的交易策略，实现自动化交易决策。

## What Changes

- **策略接口**: 定义无状态、无副作用的策略接口 `Strategy`，接收市场和持仓信息，输出交易决策
- **策略执行器**: 在指定市场上应用策略进行交易，支持 dry-run 模拟交易和真实交易（预留）
- **交易 UI**: 实时展示策略执行情况，包括持仓、现金、交易记录等
- **CLI 命令**: 新增 `nfa tools polymarket trade` 子命令，支持指定策略、市场、交易倍数等参数

## Capabilities

### New Capabilities

- `strategy-trading`: 策略交易系统核心能力，包含策略接口、执行器、持仓管理

### Modified Capabilities

- `polymarket-tools`: 新增 `trade` 子命令

## Impact

- **New Package**: `pkg/trading/` - 策略接口、执行器、持仓管理
- **New Package**: `pkg/ui/trading/` - 交易 UI 页面
- **Code Changes**: `pkg/commands/tools_polymarket.go` - 新增 trade 子命令
- **Code Changes**: `pkg/polymarket/order.go` - 新增下单/撤单 API（空实现，TODO）
- **Backwards Compatible**: 不影响现有功能，新增独立命令入口
