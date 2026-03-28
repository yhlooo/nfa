package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	pm "github.com/yhlooo/nfa/pkg/polymarket"
)

// EventPage Event 详情页
type EventPage struct {
	client pm.GammaAPIClient
	ctx    context.Context

	event   *pm.Event
	cursor  int
	width   int
	loading bool
	loadErr error
}

// NewEventPage 创建 Event 详情页
func NewEventPage(client pm.GammaAPIClient, event *pm.Event) *EventPage {
	return &EventPage{
		client: client,
		event:  event,
		width:  80,
	}
}

// SetContext 设置 context
func (p *EventPage) SetContext(ctx context.Context) {
	p.ctx = ctx
}

func (p *EventPage) Type() string { return "event" }
func (p *EventPage) OnPush() {
	// 如果事件没有 markets 数据，通过 API 加载完整事件
	if len(p.event.Markets) == 0 && p.event.Slug != "" {
		// 异步加载，Init 会返回 Cmd
	}
}
func (p *EventPage) OnPop() {}

// Init 初始化
func (p *EventPage) Init() tea.Cmd {
	if len(p.event.Markets) == 0 && p.event.Slug != "" {
		p.loading = true
		return p.loadEvent
	}
	return nil
}

// loadEvent 加载完整事件数据
func (p *EventPage) loadEvent() tea.Msg {
	result, err := p.client.GetEventBySlug(p.ctx, &pm.GetEventBySlugRequest{
		Slug: p.event.Slug,
	})
	if err != nil {
		return eventLoadedMsg{err: err}
	}
	return eventLoadedMsg{event: result}
}

// eventLoadedMsg 事件加载完成消息
type eventLoadedMsg struct {
	event *pm.Event
	err   error
}

// Update 处理更新
func (p *EventPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = m.Width
		return p, nil

	case eventLoadedMsg:
		if m.err != nil {
			p.loadErr = m.err
		} else {
			p.event = m.event
		}
		p.loading = false
		return p, nil

	case tea.KeyMsg:
		switch m.Type {
		case tea.KeyCtrlC:
			return p, tea.Quit
		case tea.KeyUp:
			if p.cursor > 0 {
				p.cursor--
			}
		case tea.KeyDown:
			if p.cursor < len(p.event.Markets)-1 {
				p.cursor++
			}
		case tea.KeyEnter:
			if p.cursor >= 0 && p.cursor < len(p.event.Markets) {
				market := p.event.Markets[p.cursor]
				return p, PushPage(NewMarketPage(p.client, &market))
			}
		}
	}

	return p, nil
}

// View 渲染
func (p *EventPage) View() string {
	var b strings.Builder

	// 标题
	b.WriteString(StyleTitle.Render(p.event.Title))
	b.WriteString("\n")
	if p.event.Slug != "" {
		b.WriteString(StyleDim.Render(p.event.Slug))
		b.WriteString("\n")
	}

	// 副标题
	if p.event.SubTitle != "" {
		b.WriteString(StyleSubtitle.Render(p.event.SubTitle))
		b.WriteString("\n")
	}

	// 描述
	if p.event.Description != "" {
		desc := truncate(p.event.Description, p.width)
		b.WriteString(StyleDescription.Render(desc))
		b.WriteString("\n")
	}

	// 交易量信息
	b.WriteString("\n")
	vol24hr := formatVolume(p.event.Volume24hr)
	liq := formatVolume(p.event.Liquidity)
	if vol24hr != "0" {
		b.WriteString(StyleLabel.Render("Volume 24h: "))
		b.WriteString(StyleValue.Render("$" + vol24hr))
		b.WriteString("  ")
	}
	if liq != "0" {
		b.WriteString(StyleLabel.Render("Liquidity: "))
		b.WriteString(StyleValue.Render("$" + liq))
	}
	b.WriteString("\n")

	// 分隔线
	b.WriteString(StyleSeparator.Render(strings.Repeat("─", p.width)))
	b.WriteString("\n")

	// 市场列表
	b.WriteString(StyleLabel.Render(fmt.Sprintf("Markets (%d)", len(p.event.Markets))))
	b.WriteString("\n\n")

	if p.loading {
		b.WriteString(StyleLoading.Render("  Loading..."))
	} else if p.loadErr != nil {
		b.WriteString(StyleDisconnected.Render(fmt.Sprintf("  Error: %v", p.loadErr)))
	} else if len(p.event.Markets) == 0 {
		b.WriteString(StyleEmpty.Render("  No markets in this event"))
	} else {
		for i, market := range p.event.Markets {
			selected := i == p.cursor
			line := p.renderMarketItem(&market, selected)
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// 底部提示
	b.WriteString(StyleSeparator.Render(strings.Repeat("─", p.width)))
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("↑↓ Navigate  Enter Watch  Esc Back"))

	return b.String()
}

func (p *EventPage) renderMarketItem(market *pm.Market, selected bool) string {
	question := market.Question
	if question == "" {
		question = market.Slug
	}
	if len(question) > 45 {
		question = question[:42] + "..."
	}

	// 解析 outcomes 和 prices
	priceStr := ""
	outcomes := parseJSONStringArray(market.Outcomes)
	prices := parseJSONStringArray(market.OutcomePrices)
	if len(outcomes) == len(prices) && len(outcomes) > 0 {
		parts := make([]string, 0, len(outcomes))
		for j, o := range outcomes {
			parts = append(parts, fmt.Sprintf("%s %s", o, prices[j]))
		}
		priceStr = strings.Join(parts, " / ")
	}

	line := "  " + question
	if priceStr != "" {
		remaining := p.width - len(line) - 4
		if remaining < len(priceStr)+2 {
			priceStr = priceStr[:remaining-3] + "..."
		}
		line += "  " + StyleDim.Render(priceStr)
	}

	if selected {
		return StyleCursor.Render(">") + line[1:]
	}
	return line
}

// parseJSONStringArray 解析 JSON 字符串数组
func parseJSONStringArray(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return nil
	}
	return result
}

var _ Page = (*EventPage)(nil)
var _ tea.Model = (*EventPage)(nil)
