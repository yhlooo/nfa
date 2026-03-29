package trading

import (
	"context"
	"time"
)

// Strategy 策略接口
//
// 策略必须是无状态且无副作用的，可以在任何时候调用
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

// NewResult 创建策略输出
func NewResult() *Result {
	return &Result{}
}

// AddOrder 添加订单请求
func (r *Result) AddOrder(order *OrderRequest) {
	r.Orders = append(r.Orders, order)
}

// AddCancelID 添加取消订单 ID
func (r *Result) AddCancelID(id string) {
	r.CancelIDs = append(r.CancelIDs, id)
}

// IsEmpty 是否为空（不采取行动）
func (r *Result) IsEmpty() bool {
	return len(r.Orders) == 0 && len(r.CancelIDs) == 0
}
