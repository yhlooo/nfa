package polymarket

import (
	"context"

	"github.com/gorilla/websocket"
)

// ClientInterface PolyMarket 客户端
type ClientInterface interface {
	CommonClientInterface
	GammaAPIClient
	DataAPIClient
	CLOBReaderClient
	CLOBWriterClient
}

// GammaAPIClient Gamma API 客户端
type GammaAPIClient interface {
	// GetEventBySlug 通过 slug 获取事件
	GetEventBySlug(ctx context.Context, req *GetEventBySlugRequest) (*Event, error)
	// GetMarketBySlug 通过 slug 获取市场
	GetMarketBySlug(ctx context.Context, slug string) (*Market, error)
}

// DataAPIClient Data API 客户端
type DataAPIClient interface{}

// CLOBReaderClient CLOB 读 API 客户端
type CLOBReaderClient interface {
	// ConnectMarketChannel WebSocket 连接市场信道
	ConnectMarketChannel(ctx context.Context) (*websocket.Conn, error)
}

// CLOBWriterClient CLOB 写 API 客户端
type CLOBWriterClient interface {
	// SendHeartbeat 发送心跳
	SendHeartbeat(ctx context.Context) (*HeartbeatStatus, error)
	// GetUserOrders 获取用户订单
	GetUserOrders(ctx context.Context, req *GetUserOrdersRequest) (*OrdersList, error)
}

// CommonClientInterface 通用客户端
type CommonClientInterface interface {
	// AuthInfo 获取认证信息
	AuthInfo() AuthInfo
	// SetAuthInfo 设置认证信息
	SetAuthInfo(authInfo AuthInfo)
}

// APIKeyInfo API 密钥信息
type APIKeyInfo struct {
	APIKey     string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// ListMeta 列表元信息
type ListMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor"`
	Count      int    `json:"count"`
}

// GetEventBySlugRequest 通过 slug 获取事件请求
type GetEventBySlugRequest struct {
	Slug            string
	IncludeChat     bool
	IncludeTemplate bool
}

// Event 事件
type Event struct {
	ID          string `json:"id"`
	Ticker      string `json:"ticker,omitempty"`
	Slug        string `json:"slug,omitempty"`
	Title       string `json:"title,omitempty"`
	SubTitle    string `json:"subtitle,omitempty"`
	Description string `json:"description,omitempty"`
	// ...
}

// GetUserOrdersRequest 获取用户订单请求
type GetUserOrdersRequest struct {
	ID         string
	Maker      string
	AssetID    string
	NextCursor string
}

// OrdersList 订单列表
type OrdersList struct {
	ListMeta
	Data []Order `json:"data"`
}

// Order 订单
type Order struct {
	ID              string        `json:"id"`
	Status          string        `json:"status"`
	Owner           string        `json:"owner"`
	MakerAddress    string        `json:"maker_address"`
	Market          string        `json:"market"`
	AssetID         string        `json:"asset_id"`
	Side            string        `json:"side"`
	OriginalSize    string        `json:"original_size"`
	SizeMatched     string        `json:"size_matched"`
	Price           string        `json:"price"`
	Outcome         string        `json:"outcome"`
	Expiration      string        `json:"expiration"`
	OrderType       string        `json:"order_type"`
	CreatedAt       int           `json:"created_at"`
	AssociateTrades []interface{} `json:"associate_trades,omitempty"`
}

// HeartbeatStatus 心跳状态
type HeartbeatStatus struct {
	Status string `json:"status"`
}

// Market 市场信息
type Market struct {
	ID           string `json:"id"`
	Question     string `json:"question"`
	Description  string `json:"description"`
	ConditionID  string `json:"conditionId"`
	Slug         string `json:"slug"`
	ClobTokenIDs string `json:"clobTokenIds"` // JSON 字符串数组
	Outcomes     string `json:"outcomes"`     // JSON 字符串数组
	Active       bool   `json:"active"`
	Closed       bool   `json:"closed"`
}
