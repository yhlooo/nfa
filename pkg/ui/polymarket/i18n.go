package polymarket

import "github.com/nicksnyder/go-i18n/v2/i18n"

// 国际化消息定义
var (
	// 浏览器标题
	MsgBrowserTitle = &i18n.Message{
		ID: "PolyMarketBrowser.Title", Other: "PolyMarket Browser",
	}

	// 首页
	MsgHomeSearchPlaceholder = &i18n.Message{
		ID: "PolyMarketBrowser.Home.SearchPlaceholder", Other: "Search events and series...",
	}
	MsgHomeLoading = &i18n.Message{
		ID: "PolyMarketBrowser.Home.Loading", Other: "Loading...",
	}
	MsgHomeNoResults = &i18n.Message{
		ID: "PolyMarketBrowser.Home.NoResults", Other: "No results found",
	}
	MsgHomeSeriesTag = &i18n.Message{
		ID: "PolyMarketBrowser.Home.SeriesTag", Other: "Series",
	}
	MsgHomeEventTag = &i18n.Message{
		ID: "PolyMarketBrowser.Home.EventTag", Other: "Event",
	}
	MsgHomeHint = &i18n.Message{
		ID: "PolyMarketBrowser.Home.Hint", Other: "↑↓ Navigate  Tab Search  Enter Open  Esc Quit",
	}

	// Series 页
	MsgSeriesVolume = &i18n.Message{
		ID: "PolyMarketBrowser.Series.Volume", Other: "Volume: {{.Volume}}",
	}
	MsgSeriesEvents = &i18n.Message{
		ID: "PolyMarketBrowser.Series.Events", Other: "Events",
	}
	MsgSeriesNoEvents = &i18n.Message{
		ID: "PolyMarketBrowser.Series.NoEvents", Other: "No events in this series",
	}
	MsgSeriesHint = &i18n.Message{
		ID: "PolyMarketBrowser.Series.Hint", Other: "↑↓ Navigate  Enter Open  Esc Back",
	}

	// Event 页
	MsgEventVolume24h = &i18n.Message{
		ID: "PolyMarketBrowser.Event.Volume24h", Other: "Volume 24h: {{.Volume}}",
	}
	MsgEventLiquidity = &i18n.Message{
		ID: "PolyMarketBrowser.Event.Liquidity", Other: "Liquidity: {{.Liquidity}}",
	}
	MsgEventMarkets = &i18n.Message{
		ID: "PolyMarketBrowser.Event.Markets", Other: "Markets",
	}
	MsgEventNoMarkets = &i18n.Message{
		ID: "PolyMarketBrowser.Event.NoMarkets", Other: "No markets in this event",
	}
	MsgEventHint = &i18n.Message{
		ID: "PolyMarketBrowser.Event.Hint", Other: "↑↓ Navigate  Enter Watch  Esc Back",
	}

	// Market 页
	MsgMarketBid = &i18n.Message{
		ID: "PolyMarketBrowser.Market.Bid", Other: "Bid:",
	}
	MsgMarketAsk = &i18n.Message{
		ID: "PolyMarketBrowser.Market.Ask", Other: "Ask:",
	}
	MsgMarketConnected = &i18n.Message{
		ID: "PolyMarketBrowser.Market.Connected", Other: "Connected",
	}
	MsgMarketDisconnected = &i18n.Message{
		ID: "PolyMarketBrowser.Market.Disconnected", Other: "Disconnected (reconnecting...)",
	}
	MsgMarketLastUpdate = &i18n.Message{
		ID: "PolyMarketBrowser.Market.LastUpdate", Other: "Last update: {{.Time}}",
	}
	MsgMarketHint = &i18n.Message{
		ID: "PolyMarketBrowser.Market.Hint", Other: "Esc Back",
	}

	// 通用
	MsgConnected = &i18n.Message{
		ID: "PolyMarketBrowser.Connected", Other: "Connected",
	}
	MsgDisconnected = &i18n.Message{
		ID: "PolyMarketBrowser.Disconnected", Other: "Disconnected (reconnecting...)",
	}
	MsgPressEscToBack = &i18n.Message{
		ID: "PolyMarketBrowser.PressEscToBack", Other: "Esc Back",
	}
	MsgPressCtrlCToExit = &i18n.Message{
		ID: "PolyMarketBrowser.PressCtrlCToExit", Other: "Ctrl+C Quit",
	}
)
