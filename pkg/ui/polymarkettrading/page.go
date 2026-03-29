package polymarkettrading

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	i18nutil "github.com/yhlooo/nfa/pkg/i18n"
	trading2 "github.com/yhlooo/nfa/pkg/polymarket/trading"
)

// Page 交易页面
type Page struct {
	ctx        context.Context
	executor   *trading2.Executor
	strategy   trading2.Strategy
	dryRun     bool
	multiplier float64
	marketSlug string

	width  int
	height int

	// 状态
	connected bool
	lastEvent time.Time
}

// NewPage 创建交易页面
func NewPage(executor *trading2.Executor, strategy trading2.Strategy, dryRun bool, multiplier float64, marketSlug string) *Page {
	return &Page{
		executor:   executor,
		strategy:   strategy,
		dryRun:     dryRun,
		multiplier: multiplier,
		marketSlug: marketSlug,
		connected:  true,
	}
}

// SetContext 设置 context
func (p *Page) SetContext(ctx context.Context) {
	p.ctx = ctx
}

// Init 初始化
func (p *Page) Init() tea.Cmd {
	return p.waitForEvent
}

// Update 处理更新
func (p *Page) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = m.Width
		p.height = m.Height

	case tea.KeyMsg:
		// 处理退出按键
		switch m.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return p, tea.Quit
		}

	case executorEventMsg:
		p.handleExecutorEvent(m.event)

	case tickMsg:
		// 定时刷新，继续等待下一个事件
		return p, p.waitForEvent
	}

	return p, p.waitForEvent
}

// View 渲染
func (p *Page) View() string {
	var b strings.Builder

	// 市场信息
	marketInfo := p.executor.MarketInfo()
	b.WriteString(StyleTitle.Render(truncate(marketInfo.Question, p.width)))
	b.WriteString("\n")
	if marketInfo.Slug != "" {
		b.WriteString(StyleSubtitle.Render(marketInfo.Slug))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// 底层资产价格
	if underlying := p.executor.UnderlyingPrice(); underlying != nil {
		b.WriteString(StylePrice.Render(fmt.Sprintf("%s: $%.2f", i18nutil.T(MsgUnderlyingPrice), *underlying)))
		if target := p.executor.PriceToBeat(); target != nil {
			b.WriteString("  ")
			b.WriteString(StyleValue.Render(fmt.Sprintf("%s: $%.2f", i18nutil.T(MsgTargetPrice), *target)))
		}
		b.WriteString("\n\n")
	}

	// 价格卡片
	priceHistory := p.executor.PriceHistory()
	yesBid := priceHistory.LatestYesBid()
	yesAsk := priceHistory.LatestYesAsk()
	noBid := priceHistory.LatestNoBid()
	noAsk := priceHistory.LatestNoAsk()

	cardWidth := (p.width - 4) / 2
	if cardWidth < 20 {
		cardWidth = 20
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(cardWidth)

	// Yes 卡片
	yesCard := p.renderPriceCard("YES", yesBid, yesAsk)
	b.WriteString(cardStyle.Render(yesCard))

	b.WriteString("  ")

	// No 卡片
	noCard := p.renderPriceCard("NO", noBid, noAsk)
	b.WriteString(cardStyle.Render(noCard))

	b.WriteString("\n\n")

	// 持仓信息
	b.WriteString(StyleLabel.Render(i18nutil.T(MsgPortfolio)))
	b.WriteString("\n")
	b.WriteString(StyleSeparator.Render(strings.Repeat("─", p.width)))
	b.WriteString("\n")

	position := p.executor.Position()
	cashStr := fmt.Sprintf("%s: $%.2f", i18nutil.T(MsgCash), position.Cash)
	yesStr := fmt.Sprintf("%s: %.0f %s @ $%.2f", i18nutil.T(MsgYesShares), position.YesShares, i18nutil.T(MsgShares), position.YesAvgCost)
	noStr := fmt.Sprintf("%s: %.0f %s", i18nutil.T(MsgNoShares), position.NoShares, i18nutil.T(MsgShares))

	b.WriteString(fmt.Sprintf("%s    %s    %s", cashStr, yesStr, noStr))
	b.WriteString("\n")

	// 计算总价值
	if yesBid != nil && noBid != nil {
		totalValue := position.TotalValue(*yesBid, *noBid)
		posValue := position.YesShares**yesBid + position.NoShares**noBid

		var totalStyle lipgloss.Style
		if totalValue >= 0 {
			totalStyle = StyleProfit
		} else {
			totalStyle = StyleLoss
		}

		b.WriteString(fmt.Sprintf("%s: $%.2f    %s: %s",
			i18nutil.T(MsgPositionValue), posValue,
			i18nutil.T(MsgTotalValue), totalStyle.Render(fmt.Sprintf("$%.2f", totalValue))))
	}

	b.WriteString("\n\n")

	// 交易记录
	b.WriteString(StyleLabel.Render(i18nutil.T(MsgTradeHistory)))
	b.WriteString("\n")
	b.WriteString(StyleSeparator.Render(strings.Repeat("─", p.width)))
	b.WriteString("\n")

	orders := p.executor.Orders()
	if len(orders) == 0 {
		b.WriteString(StyleDim.Render(i18nutil.T(MsgNoTrades)))
	} else {
		// 只显示最近 10 条
		start := 0
		if len(orders) > 10 {
			start = len(orders) - 10
		}
		for i := start; i < len(orders); i++ {
			order := orders[i]
			b.WriteString(p.renderOrder(order))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// 状态栏
	b.WriteString(StyleSeparator.Render(strings.Repeat("─", p.width)))
	b.WriteString("\n")

	var statusText string
	if p.connected {
		statusText = "● " + i18nutil.T(MsgConnected)
		statusText = StyleConnected.Render(statusText)
	} else {
		statusText = "⚠ " + i18nutil.T(MsgDisconnected)
		statusText = StyleDisconnected.Render(statusText)
	}

	modeText := i18nutil.T(MsgDryRun)
	if !p.dryRun {
		modeText = i18nutil.T(MsgRealMode)
	}

	statusBar := fmt.Sprintf("%s | %s: %s | %s: %s | %s: %.0fx | %s",
		statusText,
		i18nutil.T(MsgStrategy), p.strategy.Name(),
		i18nutil.T(MsgMode), modeText,
		i18nutil.T(MsgMultiplier), p.multiplier,
		i18nutil.T(MsgPressEscToExit),
	)
	b.WriteString(statusBar)

	return b.String()
}

// renderPriceCard 渲染价格卡片
func (p *Page) renderPriceCard(outcome string, bid, ask *float64) string {
	var b strings.Builder

	b.WriteString(StyleLabel.Render(outcome))
	b.WriteString("\n\n")

	if bid != nil {
		b.WriteString(StyleBid.Render(i18nutil.T(MsgBid) + ":"))
		b.WriteString(" ")
		b.WriteString(StyleValue.Render(fmt.Sprintf("%.4f", *bid)))
	} else {
		b.WriteString(StyleBid.Render(i18nutil.T(MsgBid) + ":"))
		b.WriteString(" ")
		b.WriteString(StyleDim.Render("-"))
	}
	b.WriteString("\n")

	if ask != nil {
		b.WriteString(StyleAsk.Render(i18nutil.T(MsgAsk) + ":"))
		b.WriteString(" ")
		b.WriteString(StyleValue.Render(fmt.Sprintf("%.4f", *ask)))
	} else {
		b.WriteString(StyleAsk.Render(i18nutil.T(MsgAsk) + ":"))
		b.WriteString(" ")
		b.WriteString(StyleDim.Render("-"))
	}

	return b.String()
}

// renderOrder 渲染订单
func (p *Page) renderOrder(order *trading2.Order) string {
	timeStr := order.CreatedAt.Format("15:04:05")

	var sideStyle lipgloss.Style
	var sideText string
	if order.Side == trading2.OrderSideBuy {
		sideStyle = StyleBuy
		sideText = i18nutil.T(MsgBuy)
	} else {
		sideStyle = StyleSell
		sideText = i18nutil.T(MsgSell)
	}

	var outcomeStyle lipgloss.Style
	if order.Outcome == trading2.OutcomeYes {
		outcomeStyle = StyleYes
	} else {
		outcomeStyle = StyleNo
	}

	var statusText string
	if order.IsFilled() {
		statusText = fmt.Sprintf("(%s %.4f)", i18nutil.T(MsgFilled), order.FilledPrice)
	} else if order.IsPending() {
		statusText = fmt.Sprintf("(%s)", i18nutil.T(MsgPending))
	} else {
		statusText = fmt.Sprintf("(%s)", i18nutil.T(MsgCancelled))
	}

	return fmt.Sprintf("%s  %s  %s  %.0f %s %.4f  %s",
		StyleDim.Render(timeStr),
		sideStyle.Render(sideText),
		outcomeStyle.Render(string(order.Outcome)),
		order.Size,
		i18nutil.T(MsgAt),
		order.Price,
		StyleDim.Render(statusText),
	)
}

// handleExecutorEvent 处理执行器事件
func (p *Page) handleExecutorEvent(event trading2.ExecutorEvent) {
	p.lastEvent = event.Timestamp
}

// waitForEvent 等待事件
func (p *Page) waitForEvent() tea.Msg {
	if p.executor == nil {
		return nil
	}

	select {
	case event, ok := <-p.executor.Events():
		if !ok {
			return nil
		}
		return executorEventMsg{event: event}
	case <-time.After(200 * time.Millisecond):
		// 定时触发刷新，即使没有事件也让 UI 更新
		return tickMsg{}
	}
}

// executorEventMsg 执行器事件消息
type executorEventMsg struct {
	event trading2.ExecutorEvent
}

// tickMsg 定时刷新消息
type tickMsg struct{}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
