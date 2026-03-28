package polymarket

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	i18nutil "github.com/yhlooo/nfa/pkg/i18n"
	pm "github.com/yhlooo/nfa/pkg/polymarket"
)

// MarketPage Market 实时监听页
type MarketPage struct {
	ctx    context.Context
	client pm.GammaAPIClient

	market  *pm.Market
	watcher *pm.Watcher

	bestBids   map[string]string
	bestAsks   map[string]string
	lastUpdate time.Time
	connected  bool
	width      int

	started bool // Watcher 是否已启动
}

// NewMarketPage 创建 Market 实时监听页
func NewMarketPage(client pm.GammaAPIClient, market *pm.Market) *MarketPage {
	return &MarketPage{
		client:    client,
		market:    market,
		bestBids:  make(map[string]string),
		bestAsks:  make(map[string]string),
		connected: true,
		width:     80,
	}
}

// SetContext 设置 context（在 push 前由 Browser 调用）
func (p *MarketPage) SetContext(ctx context.Context) {
	p.ctx = ctx
}

func (p *MarketPage) Type() string { return "market" }

func (p *MarketPage) OnPush() {
	// 解析 asset IDs
	assetIDs := parseJSONStringArray(p.market.ClobTokenIDs)
	if len(assetIDs) == 0 {
		return
	}

	// 创建并启动 Watcher
	// 需要 *pm.Client（实现 Watcher 所需接口）
	if c, ok := p.client.(*pm.Client); ok {
		p.watcher = pm.NewWatcher(c, assetIDs)
		if err := p.watcher.Start(p.ctx); err != nil {
			p.connected = false
			return
		}
		p.started = true
	}
}

func (p *MarketPage) OnPop() {
	// 停止 Watcher
	if p.watcher != nil && p.started {
		_ = p.watcher.Stop()
		p.watcher = nil
		p.started = false
	}
}

// Init 初始化
func (p *MarketPage) Init() tea.Cmd {
	return p.waitForEvent
}

// Update 处理更新
func (p *MarketPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = m.Width

	case tea.KeyMsg:
		// 所有按键都交给 Browser 处理（Esc 返回等）

	case marketEventMsg:
		p.handleMarketEvent(m.event)

	case connectionStateMsg:
		p.connected = m.state.Connected
		if m.state.LastUpdate.After(p.lastUpdate) {
			p.lastUpdate = m.state.LastUpdate
		}
	}

	return p, p.waitForEvent
}

// View 渲染
func (p *MarketPage) View() string {
	var b strings.Builder

	// 标题
	b.WriteString(StyleTitle.Render(p.market.Question))
	b.WriteString("\n")
	if p.market.Slug != "" {
		b.WriteString(StyleDim.Render(p.market.Slug))
		b.WriteString("\n")
	}

	// 描述
	if p.market.Description != "" {
		desc := truncate(p.market.Description, p.width)
		b.WriteString(StyleDescription.Render(desc))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// 价格卡片
	outcomes := parseJSONStringArray(p.market.Outcomes)
	assetIDs := parseJSONStringArray(p.market.ClobTokenIDs)

	cardWidth := (p.width - 4) / 2
	if cardWidth < 20 {
		cardWidth = 20
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(cardWidth)

	var cards []string
	for i, assetID := range assetIDs {
		if i >= len(outcomes) {
			break
		}
		name := outcomes[i]
		bid := p.bestBids[assetID]
		ask := p.bestAsks[assetID]

		if bid == "" {
			bid = "-"
		}
		if ask == "" {
			ask = "-"
		}

		cardContent := fmt.Sprintf("%s\n\n%s %s\n%s %s",
			StyleLabel.Render(name),
			StyleBid.Render("Bid:"), StyleValue.Render(bid),
			StyleAsk.Render("Ask:"), StyleValue.Render(ask),
		)
		cards = append(cards, cardStyle.Render(cardContent))
	}

	if len(cards) > 0 {
		row := lipgloss.JoinHorizontal(lipgloss.Top, cards...)
		b.WriteString(row)
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// 底部状态栏
	b.WriteString(StyleSeparator.Render(strings.Repeat("─", p.width)))
	b.WriteString("\n")

	var statusText string
	if p.connected {
		statusText = "● " + i18nutil.T(MsgMarketConnected)
		statusText = StyleConnected.Render(statusText)
	} else {
		statusText = "⚠ " + i18nutil.T(MsgMarketDisconnected)
		statusText = StyleDisconnected.Render(statusText)
	}

	lastUpdateText := ""
	if !p.lastUpdate.IsZero() {
		lastUpdateText = "Last update: " + p.lastUpdate.Format("15:04:05")
	}

	footer := fmt.Sprintf("%s | %s | Esc Back", statusText, lastUpdateText)
	b.WriteString(footer)

	return b.String()
}

func (p *MarketPage) handleMarketEvent(event pm.MarketEvent) {
	p.lastUpdate = event.Timestamp

	switch event.Type {
	case "book":
		if book, ok := event.Data.(*pm.BookEvent); ok {
			if len(book.Bids) > 0 {
				p.bestBids[book.AssetID] = book.Bids[0].Price
			}
			if len(book.Asks) > 0 {
				p.bestAsks[book.AssetID] = book.Asks[0].Price
			}
		}

	case "price_change":
		if pc, ok := event.Data.(*pm.PriceChange); ok {
			if pc.BestBid != "" {
				p.bestBids[pc.AssetID] = pc.BestBid
			}
			if pc.BestAsk != "" {
				p.bestAsks[pc.AssetID] = pc.BestAsk
			}
		}
	}
}

func (p *MarketPage) waitForEvent() tea.Msg {
	if p.watcher == nil {
		return nil
	}

	select {
	case event, ok := <-p.watcher.Events():
		if !ok {
			return nil
		}
		return marketEventMsg{event: event}
	case state, ok := <-p.watcher.ConnectionState():
		if !ok {
			return nil
		}
		return connectionStateMsg{state: state}
	}
}

// marketEventMsg 市场事件消息
type marketEventMsg struct {
	event pm.MarketEvent
}

// connectionStateMsg 连接状态消息
type connectionStateMsg struct {
	state pm.ConnectionState
}

var _ Page = (*MarketPage)(nil)
var _ tea.Model = (*MarketPage)(nil)
