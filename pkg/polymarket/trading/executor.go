package trading

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"

	pm "github.com/yhlooo/nfa/pkg/polymarket"
)

// ExecutorEvent 执行器事件
type ExecutorEvent struct {
	Type      string      // "order_filled", "order_created", "order_cancelled", "position_update", "error"
	Data      interface{} // 事件数据
	Timestamp time.Time
}

// Executor 策略执行器
type Executor struct {
	client     *pm.Client
	market     *pm.Market
	strategy   Strategy
	dryRun     bool
	multiplier float64
	interval   time.Duration

	// 内部状态
	position     *Position
	orders       []*Order
	priceHistory *PriceHistory

	// 市场信息
	marketInfo MarketInfo
	assetIDs   []string

	// 底层资产价格
	underlyingPrice *float64
	priceToBeat     *float64

	// Watcher
	watcher *pm.Watcher

	// 事件通道
	eventCh chan ExecutorEvent

	// 控制
	stopCh chan struct{}
	wg     sync.WaitGroup

	// 订单 ID 计数器（用于 dry-run）
	orderIDCounter int
}

// NewExecutor 创建策略执行器
func NewExecutor(client *pm.Client, market *pm.Market, strategy Strategy, opts ExecutorOptions) *Executor {
	// 解析 asset IDs
	var assetIDs []string
	if market.ClobTokenIDs != "" {
		_ = json.Unmarshal([]byte(market.ClobTokenIDs), &assetIDs)
	}

	// 创建市场信息
	marketInfo := MarketInfo{
		ID:          market.ID,
		Slug:        market.Slug,
		Question:    market.Question,
		Description: market.Description,
	}
	if len(assetIDs) >= 2 {
		marketInfo.YesAssetID = assetIDs[0]
		marketInfo.NoAssetID = assetIDs[1]
	}

	return &Executor{
		client:       client,
		market:       market,
		strategy:     strategy,
		dryRun:       opts.DryRun,
		multiplier:   opts.Multiplier,
		interval:     opts.Interval,
		position:     NewPosition(),
		orders:       make([]*Order, 0),
		priceHistory: NewPriceHistory(1000),
		marketInfo:   marketInfo,
		assetIDs:     assetIDs,
		eventCh:      make(chan ExecutorEvent, 100),
		stopCh:       make(chan struct{}),
	}
}

// ExecutorOptions 执行器选项
type ExecutorOptions struct {
	DryRun     bool
	Multiplier float64
	Interval   time.Duration
}

// Events 返回事件通道
func (e *Executor) Events() <-chan ExecutorEvent {
	return e.eventCh
}

// Position 返回当前持仓
func (e *Executor) Position() *Position {
	return e.position.Clone()
}

// Orders 返回所有订单
func (e *Executor) Orders() []*Order {
	return e.orders
}

// PendingOrders 返回待成交订单
func (e *Executor) PendingOrders() []*Order {
	var pending []*Order
	for _, o := range e.orders {
		if o.IsPending() {
			pending = append(pending, o)
		}
	}
	return pending
}

// PriceHistory 返回价格历史
func (e *Executor) PriceHistory() *PriceHistory {
	return e.priceHistory
}

// MarketInfo 返回市场信息
func (e *Executor) MarketInfo() MarketInfo {
	return e.marketInfo
}

// UnderlyingPrice 返回底层资产价格
func (e *Executor) UnderlyingPrice() *float64 {
	return e.underlyingPrice
}

// PriceToBeat 返回目标价格
func (e *Executor) PriceToBeat() *float64 {
	return e.priceToBeat
}

// Run 启动执行器
func (e *Executor) Run(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx)

	// 创建并启动 Watcher
	e.watcher = pm.NewWatcher(e.client, e.market, e.assetIDs)
	if err := e.watcher.Start(ctx); err != nil {
		return fmt.Errorf("start watcher error: %w", err)
	}

	// 启动事件处理循环
	e.wg.Add(2)
	go e.eventLoop(ctx)
	go e.timerLoop(ctx)

	logger.V(1).Info("executor started", "market", e.marketInfo.Slug, "strategy", e.strategy.Name())
	return nil
}

// Stop 停止执行器
func (e *Executor) Stop() error {
	close(e.stopCh)
	e.wg.Wait()

	if e.watcher != nil {
		_ = e.watcher.Stop()
	}

	close(e.eventCh)
	return nil
}

// eventLoop 事件处理循环
func (e *Executor) eventLoop(ctx context.Context) {
	defer e.wg.Done()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ctx.Done():
			return
		case event, ok := <-e.watcher.Events():
			if !ok {
				return
			}
			e.handleMarketEvent(ctx, event)
		}
	}
}

// timerLoop 定时触发循环
func (e *Executor) timerLoop(ctx context.Context) {
	defer e.wg.Done()

	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.executeStrategy(ctx)
		}
	}
}

// handleMarketEvent 处理市场事件
func (e *Executor) handleMarketEvent(ctx context.Context, event pm.MarketEvent) {
	logger := logr.FromContextOrDiscard(ctx)
	now := time.Now()

	switch event.Type {
	case "book":
		if book, ok := event.Data.(*pm.BookEvent); ok {
			// 更新价格历史
			if len(book.Bids) > 0 {
				if bid, err := parseFloat(book.Bids[0].Price); err == nil {
					e.priceHistory.AddYesBid(PricePoint{Time: now, Value: bid})
				}
			}
			if len(book.Asks) > 0 {
				if ask, err := parseFloat(book.Asks[0].Price); err == nil {
					e.priceHistory.AddYesAsk(PricePoint{Time: now, Value: ask})
				}
			}

			// 检查待成交订单
			e.checkPendingOrders()

			// 触发策略执行
			e.executeStrategy(ctx)
		}

	case "price_change":
		if pc, ok := event.Data.(*pm.PriceChange); ok {
			// 根据资产 ID 更新对应价格历史
			if pc.BestBid != "" {
				if bid, err := parseFloat(pc.BestBid); err == nil {
					if pc.AssetID == e.marketInfo.YesAssetID {
						e.priceHistory.AddYesBid(PricePoint{Time: now, Value: bid})
					} else if pc.AssetID == e.marketInfo.NoAssetID {
						e.priceHistory.AddNoBid(PricePoint{Time: now, Value: bid})
					}
				}
			}
			if pc.BestAsk != "" {
				if ask, err := parseFloat(pc.BestAsk); err == nil {
					if pc.AssetID == e.marketInfo.YesAssetID {
						e.priceHistory.AddYesAsk(PricePoint{Time: now, Value: ask})
					} else if pc.AssetID == e.marketInfo.NoAssetID {
						e.priceHistory.AddNoAsk(PricePoint{Time: now, Value: ask})
					}
				}
			}

			// 检查待成交订单
			e.checkPendingOrders()

			// 触发策略执行
			e.executeStrategy(ctx)
		}

	case "underlying_price":
		if up, ok := event.Data.(*pm.UnderlyingPriceEvent); ok {
			e.underlyingPrice = &up.Value
			logger.V(1).Info("underlying price updated", "value", up.Value)

			// 触发策略执行
			e.executeStrategy(ctx)
		}

	case "price_to_beat":
		if ptb, ok := event.Data.(*pm.PriceToBeatEvent); ok {
			e.priceToBeat = &ptb.PriceToBeat
			logger.V(1).Info("price to beat updated", "value", ptb.PriceToBeat)
		}
	}
}

// executeStrategy 执行策略
func (e *Executor) executeStrategy(ctx context.Context) {
	logger := logr.FromContextOrDiscard(ctx)

	// 准备输入
	input := Input{
		Market:         e.marketInfo,
		Position:       e.position.Clone(),
		Orders:         e.PendingOrders(),
		YesBidPrices:   e.priceHistory.YesBids(),
		YesAskPrices:   e.priceHistory.YesAsks(),
		NoBidPrices:    e.priceHistory.NoBids(),
		NoAskPrices:    e.priceHistory.NoAsks(),
		TrackingPrice:  e.underlyingPrice,
		TrackingTarget: e.priceToBeat,
	}

	// 执行策略
	result, err := e.strategy.Execute(ctx, input)
	if err != nil {
		logger.Error(err, "strategy execute error")
		return
	}

	if result == nil || result.IsEmpty() {
		return
	}

	// 处理取消订单
	for _, id := range result.CancelIDs {
		e.cancelOrder(ctx, id)
	}

	// 处理新订单
	for _, req := range result.Orders {
		e.createOrder(ctx, req)
	}
}

// createOrder 创建订单
func (e *Executor) createOrder(ctx context.Context, req *OrderRequest) {
	logger := logr.FromContextOrDiscard(ctx)

	// 计算实际数量（乘以倍数）
	actualSize := req.Size * e.multiplier

	// 生成订单 ID
	e.orderIDCounter++
	orderID := fmt.Sprintf("dry-run-%s-%d", uuid.New().String()[:8], e.orderIDCounter)

	// 创建订单
	order := NewOrder(orderID, req.Side, req.Outcome, req.Price, actualSize, req.Size, req.OrderType)
	e.orders = append(e.orders, order)

	logger.V(1).Info("order created",
		"id", orderID,
		"side", req.Side,
		"outcome", req.Outcome,
		"price", req.Price,
		"size", actualSize,
		"dry-run", e.dryRun,
	)

	// 发送事件
	e.eventCh <- ExecutorEvent{
		Type:      "order_created",
		Data:      order,
		Timestamp: time.Now(),
	}

	// 如果是市价单，立即成交
	if req.OrderType == OrderTypeMarket {
		e.fillMarketOrder(order)
		return
	}

	// 如果是真实交易模式
	if !e.dryRun {
		// TODO: 实现真实下单
		logger.Info("real order not implemented yet", "id", orderID)
	}
}

// cancelOrder 取消订单
func (e *Executor) cancelOrder(ctx context.Context, orderID string) {
	logger := logr.FromContextOrDiscard(ctx)

	for _, order := range e.orders {
		if order.ID == orderID && order.IsPending() {
			order.MarkCancelled()
			logger.V(1).Info("order cancelled", "id", orderID)

			// 发送事件
			e.eventCh <- ExecutorEvent{
				Type:      "order_cancelled",
				Data:      order,
				Timestamp: time.Now(),
			}

			// 如果是真实交易模式
			if !e.dryRun {
				// TODO: 实现真实撤单
				logger.Info("real cancel not implemented yet", "id", orderID)
			}
			return
		}
	}
}

// checkPendingOrders 检查待成交订单
func (e *Executor) checkPendingOrders() {
	yesBid := e.priceHistory.LatestYesBid()
	yesAsk := e.priceHistory.LatestYesAsk()
	noBid := e.priceHistory.LatestNoBid()
	noAsk := e.priceHistory.LatestNoAsk()

	for _, order := range e.orders {
		if !order.IsPending() {
			continue
		}

		// 检查是否可以成交
		if order.Outcome == OutcomeYes {
			if order.Side == OrderSideBuy && yesAsk != nil {
				// 买单：价格 >= 当前卖价时成交
				if order.Price >= *yesAsk {
					e.fillOrder(order, *yesAsk)
				}
			} else if order.Side == OrderSideSell && yesBid != nil {
				// 卖单：价格 <= 当前买价时成交
				if order.Price <= *yesBid {
					e.fillOrder(order, *yesBid)
				}
			}
		} else if order.Outcome == OutcomeNo {
			if order.Side == OrderSideBuy && noAsk != nil {
				if order.Price >= *noAsk {
					e.fillOrder(order, *noAsk)
				}
			} else if order.Side == OrderSideSell && noBid != nil {
				if order.Price <= *noBid {
					e.fillOrder(order, *noBid)
				}
			}
		}
	}
}

// fillOrder 成交订单
func (e *Executor) fillOrder(order *Order, price float64) {
	logger := logr.FromContextOrDiscard(context.Background())

	order.MarkFilled(price)

	// 更新持仓
	if order.Outcome == OutcomeYes {
		if order.Side == OrderSideBuy {
			e.position.BuyYes(order.Size, price)
		} else {
			e.position.SellYes(order.Size, price)
		}
	} else {
		if order.Side == OrderSideBuy {
			e.position.BuyNo(order.Size, price)
		} else {
			e.position.SellNo(order.Size, price)
		}
	}

	logger.V(1).Info("order filled",
		"id", order.ID,
		"side", order.Side,
		"outcome", order.Outcome,
		"price", price,
		"size", order.Size,
	)

	// 发送事件
	e.eventCh <- ExecutorEvent{
		Type:      "order_filled",
		Data:      order,
		Timestamp: time.Now(),
	}
	e.eventCh <- ExecutorEvent{
		Type:      "position_update",
		Data:      e.position.Clone(),
		Timestamp: time.Now(),
	}
}

// fillMarketOrder 市价单成交
func (e *Executor) fillMarketOrder(order *Order) {
	var price *float64

	if order.Outcome == OutcomeYes {
		price = e.priceHistory.LatestYesAsk()
		if order.Side == OrderSideSell {
			price = e.priceHistory.LatestYesBid()
		}
	} else {
		price = e.priceHistory.LatestNoAsk()
		if order.Side == OrderSideSell {
			price = e.priceHistory.LatestNoBid()
		}
	}

	if price != nil {
		e.fillOrder(order, *price)
	} else {
		// 没有价格数据，无法成交
		order.MarkCancelled()
	}
}

// parseFloat 解析浮点数
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
