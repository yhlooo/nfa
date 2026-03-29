package polymarket

import (
	"context"
)

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	// AssetID 资产 ID
	AssetID string `json:"asset_id"`
	// Side 买卖方向 BUY/SELL
	Side string `json:"side"`
	// Size 数量
	Size string `json:"size"`
	// Price 价格（限价单）
	Price string `json:"price,omitempty"`
	// OrderType 订单类型 GT (Good Till Canceled) / GTC
	OrderType string `json:"order_type"`
	// Expiration 过期时间戳
	Expiration int64 `json:"expiration,omitempty"`
}

// CreateOrderResponse 创建订单响应
type CreateOrderResponse struct {
	// OrderID 订单 ID
	OrderID string `json:"order_id"`
	// Status 状态
	Status string `json:"status"`
}

// CancelOrderRequest 取消订单请求
type CancelOrderRequest struct {
	// OrderID 订单 ID
	OrderID string `json:"order_id"`
}

// CancelOrderResponse 取消订单响应
type CancelOrderResponse struct {
	// Status 状态
	Status string `json:"status"`
}

// CreateOrder 创建订单
//
// TODO: 实现真实下单逻辑
// 需要实现：
// 1. 构造订单签名
// 2. 调用 CLOB API POST /order
// 3. 处理响应
func (c *Client) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	// TODO: 实现真实下单
	// 参考: https://docs.polymarket.com/#create-order
	return nil, ErrNotImplemented
}

// CancelOrder 取消订单
//
// TODO: 实现真实撤单逻辑
// 需要实现：
// 1. 构造取消签名
// 2. 调用 CLOB API DELETE /order/{order_id}
// 3. 处理响应
func (c *Client) CancelOrder(ctx context.Context, req *CancelOrderRequest) (*CancelOrderResponse, error) {
	// TODO: 实现真实撤单
	// 参考: https://docs.polymarket.com/#cancel-order
	return nil, ErrNotImplemented
}
