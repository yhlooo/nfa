package polymarketwatcher

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-logr/logr"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	i18nutil "github.com/yhlooo/nfa/pkg/i18n"
	"github.com/yhlooo/nfa/pkg/polymarket"
)

// Options UI 选项
type Options struct {
	Market       *polymarket.Market
	AssetIDs     []string
	OutcomeNames []string
	Watcher      *polymarket.Watcher
}

// NewUI 创建 UI
func NewUI(opts Options) *UI {
	return &UI{
		market:       opts.Market,
		assetIDs:     opts.AssetIDs,
		outcomeNames: opts.OutcomeNames,
		watcher:      opts.Watcher,
		bestBids:     make(map[string]string),
		bestAsks:     make(map[string]string),
		connected:    true,
		width:        80,
	}
}

// hasUnderlyingAsset 是否有底层资产
func (ui *UI) hasUnderlyingAsset() bool {
	return polymarket.ParseResolutionSource(ui.market) != nil
}

// UI PolyMarket 市场监听 UI
type UI struct {
	ctx    context.Context
	logger logr.Logger

	market       *polymarket.Market
	assetIDs     []string
	outcomeNames []string
	watcher      *polymarket.Watcher

	bestBids        map[string]string
	bestAsks        map[string]string
	lastUpdate      time.Time
	connected       bool
	underlyingPrice *float64 // 底层资产当前价格
	priceToBeat     *float64 // 起始价格
	underlyingSym   string   // 底层资产符号

	width int
}

var _ tea.Model = (*UI)(nil)

// Run 运行 UI
func (ui *UI) Run(ctx context.Context) error {
	ui.ctx = ctx
	ui.logger = logr.FromContextOrDiscard(ctx)

	p := tea.NewProgram(ui, tea.WithContext(ctx))
	_, err := p.Run()
	return err
}

// Init 初始化
func (ui *UI) Init() tea.Cmd {
	return ui.waitForEvent
}

// Update 处理更新
func (ui *UI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		ui.width = m.Width

	case tea.KeyMsg:
		switch m.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return ui, tea.Quit
		}

	case marketEventMsg:
		ui.handleMarketEvent(m.event)

	case connectionStateMsg:
		ui.connected = m.state.Connected
		if m.state.LastUpdate.After(ui.lastUpdate) {
			ui.lastUpdate = m.state.LastUpdate
		}
	}

	return ui, ui.waitForEvent
}

// View 渲染
func (ui *UI) View() string {
	var b strings.Builder

	// 标题样式
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("36"))
	descStyle := lipgloss.NewStyle().Faint(true)
	separatorStyle := lipgloss.NewStyle().Faint(true)

	// 市场信息
	b.WriteString(titleStyle.Render(ui.market.Question))
	b.WriteString("\n")
	if ui.market.Description != "" {
		// 限制描述长度
		desc := ui.market.Description
		if len(desc) > ui.width {
			desc = desc[:ui.width-3] + "..."
		}
		b.WriteString(descStyle.Render(desc))
		b.WriteString("\n")
	}
	b.WriteString(separatorStyle.Render(strings.Repeat("─", ui.width)))
	b.WriteString("\n\n")

	// 底层资产价格区域
	if ui.hasUnderlyingAsset() {
		underlyingStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")) // 黄色
		priceToBeatStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))           // 橙色

		if ui.underlyingPrice != nil {
			priceText := i18nutil.LocalizeContext(ui.ctx, &i18n.LocalizeConfig{
				DefaultMessage: MsgUnderlyingPrice,
				TemplateData: map[string]any{
					"Symbol": strings.ToUpper(ui.underlyingSym),
					"Value":  fmt.Sprintf("$%.2f", *ui.underlyingPrice),
				},
			})
			b.WriteString(underlyingStyle.Render(priceText))
			b.WriteString("\n")
		}

		if ui.priceToBeat != nil {
			ptbText := i18nutil.LocalizeContext(ui.ctx, &i18n.LocalizeConfig{
				DefaultMessage: MsgPriceToBeat,
				TemplateData: map[string]any{
					"Value": fmt.Sprintf("$%.2f", *ui.priceToBeat),
				},
			})
			b.WriteString(priceToBeatStyle.Render(ptbText))
			b.WriteString("\n")
		}

		b.WriteString("\n")
	}

	// 价格显示
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width((ui.width - 4) / 2)

	labelStyle := lipgloss.NewStyle().Bold(true)
	bidStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("34"))  // 绿色
	askStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // 红色
	valueStyle := lipgloss.NewStyle()

	// 渲染每个结果的价格卡片
	var cards []string
	for i, assetID := range ui.assetIDs {
		if i >= len(ui.outcomeNames) {
			break
		}
		name := ui.outcomeNames[i]
		bid := ui.bestBids[assetID]
		ask := ui.bestAsks[assetID]

		if bid == "" {
			bid = "-"
		}
		if ask == "" {
			ask = "-"
		}

		cardContent := fmt.Sprintf("%s\n\n%s %s\n%s %s",
			labelStyle.Render(name),
			bidStyle.Render("Bid:"), valueStyle.Render(bid),
			askStyle.Render("Ask:"), valueStyle.Render(ask),
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
	b.WriteString(separatorStyle.Render(strings.Repeat("─", ui.width)))
	b.WriteString("\n")

	statusStyle := lipgloss.NewStyle()
	var statusText string
	if ui.connected {
		statusStyle = statusStyle.Foreground(lipgloss.Color("34")) // 绿色
		statusText = "● " + i18nutil.TContext(ui.ctx, MsgConnected)
	} else {
		statusStyle = statusStyle.Foreground(lipgloss.Color("214")) // 橙色
		statusText = "⚠ " + i18nutil.TContext(ui.ctx, MsgDisconnected)
	}

	lastUpdateText := ""
	if !ui.lastUpdate.IsZero() {
		lastUpdateText = i18nutil.LocalizeContext(ui.ctx, &i18n.LocalizeConfig{
			DefaultMessage: MsgLastUpdate,
			TemplateData:   map[string]any{"Time": ui.lastUpdate.Format("15:04:05")},
		})
	}

	exitHint := i18nutil.TContext(ui.ctx, MsgPressCtrlCToExit)

	footer := fmt.Sprintf("%s | %s | %s", statusText, lastUpdateText, exitHint)
	b.WriteString(statusStyle.Render(footer))

	return b.String()
}

// handleMarketEvent 处理市场事件
func (ui *UI) handleMarketEvent(event polymarket.MarketEvent) {
	ui.lastUpdate = event.Timestamp

	switch event.Type {
	case "book":
		if book, ok := event.Data.(*polymarket.BookEvent); ok {
			// 从订单簿获取最佳买卖价
			if len(book.Bids) > 0 {
				ui.bestBids[book.AssetID] = book.Bids[0].Price
			}
			if len(book.Asks) > 0 {
				ui.bestAsks[book.AssetID] = book.Asks[0].Price
			}
		}

	case "price_change":
		if pc, ok := event.Data.(*polymarket.PriceChange); ok {
			// 更新最佳买卖价
			if pc.BestBid != "" {
				ui.bestBids[pc.AssetID] = pc.BestBid
			}
			if pc.BestAsk != "" {
				ui.bestAsks[pc.AssetID] = pc.BestAsk
			}
		}

	case "underlying_price":
		if up, ok := event.Data.(*polymarket.UnderlyingPriceEvent); ok {
			ui.underlyingPrice = &up.Value
			if ui.underlyingSym == "" {
				ui.underlyingSym = up.Symbol
			}
		}

	case "price_to_beat":
		if ptb, ok := event.Data.(*polymarket.PriceToBeatEvent); ok {
			ui.priceToBeat = &ptb.PriceToBeat
		}
	}
}

// marketEventMsg 市场事件消息
type marketEventMsg struct {
	event polymarket.MarketEvent
}

// connectionStateMsg 连接状态消息
type connectionStateMsg struct {
	state polymarket.ConnectionState
}

// waitForEvent 等待事件
func (ui *UI) waitForEvent() tea.Msg {
	select {
	case <-ui.ctx.Done():
		return tea.Quit
	case event, ok := <-ui.watcher.Events():
		if !ok {
			return tea.Quit
		}
		return marketEventMsg{event: event}
	case state, ok := <-ui.watcher.ConnectionState():
		if !ok {
			return tea.Quit
		}
		return connectionStateMsg{state: state}
	}
}
