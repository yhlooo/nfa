package trading

import (
	"time"
)

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
	// OriginalSize 原始数量（乘以交易倍数前）
	OriginalSize float64
	// OrderType 订单类型
	OrderType OrderType
	// Status 订单状态
	Status OrderStatus
	// FilledAt 成交时间
	FilledAt *time.Time
	// FilledPrice 成交价格
	FilledPrice float64
}

// NewOrder 创建新订单
func NewOrder(id string, side OrderSide, outcome Outcome, price, size, originalSize float64, orderType OrderType) *Order {
	return &Order{
		ID:           id,
		CreatedAt:    time.Now(),
		Side:         side,
		Outcome:      outcome,
		Price:        price,
		Size:         size,
		OriginalSize: originalSize,
		OrderType:    orderType,
		Status:       OrderStatusPending,
	}
}

// MarkFilled 标记订单已成交
func (o *Order) MarkFilled(filledPrice float64) {
	now := time.Now()
	o.Status = OrderStatusFilled
	o.FilledAt = &now
	o.FilledPrice = filledPrice
}

// MarkCancelled 标记订单已取消
func (o *Order) MarkCancelled() {
	o.Status = OrderStatusCancelled
}

// IsPending 是否待成交
func (o *Order) IsPending() bool {
	return o.Status == OrderStatusPending
}

// IsFilled 是否已成交
func (o *Order) IsFilled() bool {
	return o.Status == OrderStatusFilled
}

// IsCancelled 是否已取消
func (o *Order) IsCancelled() bool {
	return o.Status == OrderStatusCancelled
}

// OrderRequest 订单请求（策略输出）
type OrderRequest struct {
	// Side 买卖方向
	Side OrderSide
	// Outcome YES / NO
	Outcome Outcome
	// Size 数量（基础单位，执行器会乘以倍数）
	Size float64
	// Price 价格
	Price float64
	// OrderType 订单类型
	OrderType OrderType
}

// NewOrderRequest 创建订单请求
func NewOrderRequest(side OrderSide, outcome Outcome, size, price float64, orderType OrderType) *OrderRequest {
	return &OrderRequest{
		Side:      side,
		Outcome:   outcome,
		Size:      size,
		Price:     price,
		OrderType: orderType,
	}
}
