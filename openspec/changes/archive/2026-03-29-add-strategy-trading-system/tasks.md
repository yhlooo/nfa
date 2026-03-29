## 1. Core Types and Interfaces

- [x] 1.1 Create `pkg/trading/types.go` with basic types (OrderSide, Outcome, OrderType, OrderStatus)
- [x] 1.2 Create `pkg/trading/position.go` with Position struct and TotalValue method
- [x] 1.3 Create `pkg/trading/order.go` with Order struct and related methods
- [x] 1.4 Create `pkg/trading/strategy.go` with Strategy interface, Input, Result structs
- [x] 1.5 Create `pkg/trading/market_info.go` with MarketInfo struct

## 2. Price History Management

- [x] 2.1 Create `pkg/trading/price_history.go` with PricePoint struct
- [x] 2.2 Implement PriceHistory struct with maxSize limit (1000)
- [x] 2.3 Implement Add method for each price type (yesBid, yesAsk, noBid, noAsk)
- [x] 2.4 Implement getter methods for price history

## 3. Strategy Executor

- [x] 3.1 Create `pkg/trading/executor.go` with Executor struct
- [x] 3.2 Implement NewExecutor constructor
- [x] 3.3 Implement Run method to start executor (create Watcher, start event loop)
- [x] 3.4 Implement Stop method to stop executor
- [x] 3.5 Implement event handling from Watcher
- [x] 3.6 Implement strategy execution trigger (on price change and timer)
- [x] 3.7 Implement order execution (dry-run mode)
- [x] 3.8 Implement order execution (real mode, TODO placeholder)
- [x] 3.9 Implement position update logic
- [x] 3.10 Implement trade multiplier logic

## 4. Example Strategy

- [x] 4.1 Create `pkg/trading/simple_strategy.go` with SimpleStrategy
- [x] 4.2 Implement Name method returning "simple"
- [x] 4.3 Implement Execute method with buy/sell logic:
  - Buy when price < 0.2
  - Sell when position profit > 0.4

## 5. PolyMarket Order API (Placeholder)

- [x] 5.1 Create `pkg/polymarket/order.go`
- [x] 5.2 Define CreateOrderRequest struct
- [x] 5.3 Define CancelOrderRequest struct
- [x] 5.4 Implement CreateOrder method with TODO comment
- [x] 5.5 Implement CancelOrder method with TODO comment

## 6. Trading UI

- [x] 6.1 Create `pkg/ui/trading/page.go` with Page struct
- [x] 6.2 Create `pkg/ui/trading/styles.go` with style definitions
- [x] 6.3 Create `pkg/ui/trading/i18n.go` with i18n messages
- [x] 6.4 Implement Init, Update, View methods for Page
- [x] 6.5 Implement market info section
- [x] 6.6 Implement underlying asset price section
- [x] 6.7 Implement price cards (Yes/No bid/ask)
- [x] 6.8 Implement portfolio section (cash, positions, total value)
- [x] 6.9 Implement trade history section
- [x] 6.10 Implement status bar

## 7. CLI Command

- [x] 7.1 Add TradeOptions struct in `pkg/commands/tools_polymarket.go`
- [x] 7.2 Add NewTradeOptions constructor
- [x] 7.3 Add AddPFlags method for trade command flags
- [x] 7.4 Create newTradeCommand function
- [x] 7.5 Implement command logic: parse flags, get market, create executor, run UI
- [x] 7.6 Register trade command under polymarket command

## 8. Internationalization

- [x] 8.1 Add i18n messages for trading UI in `pkg/ui/trading/i18n.go`
- [x] 8.2 Run i18n-translate skill to generate translations

## 9. Testing

- [x] 9.1 Write unit tests for Position (TotalValue calculation)
- [x] 9.2 Write unit tests for PriceHistory (add, get, size limit)
- [x] 9.3 Write unit tests for SimpleStrategy
- [x] 9.4 Write unit tests for Executor (dry-run mode)

## 10. Code Quality

- [x] 10.1 Run `go fmt ./...`
- [x] 10.2 Run `go vet ./...`
- [x] 10.3 Run `go test ./...`
