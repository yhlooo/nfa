package polymarket

import (
	"context"
	"net/http"
	"net/url"
)

const (
	GammaAPIEndpoint = "https://gamma-api.polymarket.com"
	DataAPIEndpoint  = "https://data-api.polymarket.com"
	CLOBEndpoint     = "https://clob.polymarket.com"
)

// NewClient 创建 PolyMarket 客户端
func NewClient(authInfo AuthInfo) *Client {
	return &Client{
		CommonClient: NewCommonClient(authInfo),
	}
}

// Client PolyMarket 客户端
type Client struct {
	*CommonClient
}

var _ FullClient = (*Client)(nil)

// GetEventBySlug 通过 slug 获取事件
func (c *Client) GetEventBySlug(ctx context.Context, req *GetEventBySlugRequest) (*Event, error) {
	query := url.Values{}
	if req.IncludeChat {
		query.Set("include_chat", "true")
	}
	if req.IncludeTemplate {
		query.Set("include_template", "true")
	}

	event := &Event{}
	err := c.Do(ctx, &RawRequest{
		Method:   http.MethodGet,
		Endpoint: GammaAPIEndpoint,
		URI:      "/events/slug/" + req.Slug,
		Query:    query,
	}, event)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// SendHeartbeat 发送心跳
func (c *Client) SendHeartbeat(ctx context.Context) (*HeartbeatStatus, error) {
	status := &HeartbeatStatus{}
	err := c.Do(ctx, &RawRequest{
		Method:     http.MethodPost,
		Endpoint:   CLOBEndpoint,
		URI:        "/heartbeats",
		WithL2Auth: true,
	}, status)
	if err != nil {
		return nil, err
	}
	return status, nil
}

// GetUserOrders 获取用户订单
func (c *Client) GetUserOrders(ctx context.Context, req *GetUserOrdersRequest) (*OrdersList, error) {
	query := url.Values{}
	if req.ID != "" {
		query.Set("id", req.ID)
	}
	if req.Maker != "" {
		query.Set("maker", req.Maker)
	}
	if req.AssetID != "" {
		query.Set("asset_id", req.AssetID)
	}
	if req.NextCursor != "" {
		query.Set("next_cursor", req.NextCursor)
	}

	orders := &OrdersList{}
	err := c.Do(ctx, &RawRequest{
		Method:     http.MethodGet,
		Endpoint:   CLOBEndpoint,
		URI:        "/orders",
		Query:      query,
		WithL2Auth: true,
	}, orders)
	if err != nil {
		return nil, err
	}
	return orders, nil
}
