package polymarketwatcher

import "github.com/nicksnyder/go-i18n/v2/i18n"

// 国际化消息定义
var (
	MsgConnected        = &i18n.Message{ID: "PolyMarketWatcher.Connected", Other: "Connected"}
	MsgDisconnected     = &i18n.Message{ID: "PolyMarketWatcher.Disconnected", Other: "Disconnected (reconnecting...)"}
	MsgLastUpdate       = &i18n.Message{ID: "PolyMarketWatcher.LastUpdate", Other: "Last update: {{.Time}}"}
	MsgPressCtrlCToExit = &i18n.Message{ID: "PolyMarketWatcher.PressCtrlCToExit", Other: "Press Ctrl+C to exit"}
	MsgUnderlyingPrice  = &i18n.Message{ID: "PolyMarketWatcher.UnderlyingPrice", Other: "{{.Symbol}} Price: {{.Value}}"}
	MsgPriceToBeat      = &i18n.Message{ID: "PolyMarketWatcher.PriceToBeat", Other: "Price to Beat: {{.Value}}"}
)
