package polymarket

// SubscriptionRequest WebSocket 订阅请求
type SubscriptionRequest struct {
	AssetIDs    []string `json:"assets_ids"`
	Type        string   `json:"type"`         // "market"
	InitialDump bool     `json:"initial_dump"` // true
	Level       int      `json:"level"`        // 2
}

// BookEvent 订单簿快照事件
type BookEvent struct {
	EventType string       `json:"event_type"` // "book"
	AssetID   string       `json:"asset_id"`
	Market    string       `json:"market"`
	Bids      []PriceLevel `json:"bids"`
	Asks      []PriceLevel `json:"asks"`
	Timestamp string       `json:"timestamp"`
	Hash      string       `json:"hash"`
}

// PriceLevel 价格档位
type PriceLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// PriceChangeEvent 价格变化事件
type PriceChangeEvent struct {
	EventType    string        `json:"event_type"` // "price_change"
	Market       string        `json:"market"`
	PriceChanges []PriceChange `json:"price_changes"`
	Timestamp    string        `json:"timestamp"`
}

// PriceChange 价格变化
type PriceChange struct {
	AssetID string `json:"asset_id"`
	Price   string `json:"price"`
	Size    string `json:"size"`
	Side    string `json:"side"` // "BUY" / "SELL"
	Hash    string `json:"hash"`
	BestBid string `json:"best_bid"`
	BestAsk string `json:"best_ask"`
}

// LastTradePriceEvent 最近成交价事件
type LastTradePriceEvent struct {
	EventType       string `json:"event_type"` // "last_trade_price"
	AssetID         string `json:"asset_id"`
	Market          string `json:"market"`
	Price           string `json:"price"`
	Size            string `json:"size"`
	FeeRateBps      string `json:"fee_rate_bps"`
	Side            string `json:"side"` // "BUY" / "SELL"
	Timestamp       string `json:"timestamp"`
	TransactionHash string `json:"transaction_hash"`
}
