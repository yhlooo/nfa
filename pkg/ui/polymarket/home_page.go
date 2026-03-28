package polymarket

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-logr/logr"

	pm "github.com/yhlooo/nfa/pkg/polymarket"
)

// homeItem 首页列表项
type homeItem struct {
	Type       string // "series" 或 "event"
	Series     *pm.Series
	Event      *pm.Event
	SeriesSlug string // 搜索结果中的 series slug（用于按 series 聚合）
}

func (h *homeItem) title() string {
	if h.Series != nil {
		return h.Series.Title
	}
	if h.Event != nil {
		return h.Event.Title
	}
	if h.SeriesSlug != "" {
		// slug 转为可读标题：btc-up-or-down-5m -> BTC Up or Down 5M
		return slugToTitle(h.SeriesSlug)
	}
	return ""
}

// slugToTitle 将 slug 转换为可读标题
func slugToTitle(slug string) string {
	parts := strings.Split(slug, "-")
	var words []string
	for _, p := range parts {
		if p == "" {
			continue
		}
		words = append(words, strings.Title(p))
	}
	return strings.Join(words, " ")
}

func (h *homeItem) volume24hr() float64 {
	if h.Series != nil {
		return h.Series.Volume24hr
	}
	if h.Event != nil {
		return h.Event.Volume24hr
	}
	return 0
}

// HomePage 首页
type HomePage struct {
	ctx    context.Context
	client pm.GammaAPIClient
	logger logr.Logger

	searchInput textinput.Model
	searchFocus bool // 焦点在搜索框
	items       []homeItem
	cursor      int
	loading     bool
	err         error
	width       int

	mu sync.Mutex
}

// SetContext 设置 context
func (p *HomePage) SetContext(ctx context.Context) {
	p.ctx = ctx
}

// NewHomePage 创建首页
func NewHomePage(client pm.GammaAPIClient) *HomePage {
	si := textinput.New()
	si.Prompt = "> "
	si.Placeholder = "Search events and series..."
	si.CharLimit = 100
	si.Width = 60

	return &HomePage{
		client:      client,
		searchInput: si,
		width:       80,
	}
}

func (p *HomePage) Type() string { return "home" }
func (p *HomePage) OnPush() {
	p.loading = true
	p.err = nil
	p.searchFocus = true
	p.searchInput.Focus()
}
func (p *HomePage) OnPop() {}

// Init 初始化
func (p *HomePage) Init() tea.Cmd {
	return p.loadHomeItems
}

// Update 处理更新
func (p *HomePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = m.Width
		p.searchInput.Width = p.width - 4
		return p, nil

	case homeItemsLoadedMsg:
		p.mu.Lock()
		p.items = m.items
		p.loading = false
		p.err = m.err
		p.cursor = 0
		p.mu.Unlock()
		return p, nil

	case tea.KeyMsg:
		return p.handleKey(m)
	}

	return p, nil
}

func (p *HomePage) handleKey(m tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 搜索框焦点模式
	if p.searchFocus {
		switch m.Type {
		case tea.KeyEnter:
			query := strings.TrimSpace(p.searchInput.Value())
			p.searchInput.Blur()
			p.searchFocus = false
			if query == "" {
				// 清空搜索，恢复热门列表
				p.loading = true
				p.items = nil
				return p, p.loadHomeItems
			}
			// 执行搜索
			p.loading = true
			p.items = nil
			return p, p.doSearch(query)
		case tea.KeyEsc:
			p.searchInput.Blur()
			p.searchInput.SetValue("")
			p.searchFocus = false
			return p, nil
		case tea.KeyCtrlC:
			return p, tea.Quit
		case tea.KeyDown:
			// 按下键切换到列表
			p.searchInput.Blur()
			p.searchFocus = false
			if len(p.items) > 0 && p.cursor < 0 {
				p.cursor = 0
			}
			return p, nil
		}

		var cmd tea.Cmd
		p.searchInput, cmd = p.searchInput.Update(m)
		return p, cmd
	}

	// 列表焦点模式
	switch m.Type {
	case tea.KeyCtrlC:
		return p, tea.Quit
	case tea.KeyTab:
		p.searchFocus = true
		p.searchInput.Focus()
		return p, nil
	case tea.KeyUp:
		if p.cursor > 0 {
			p.cursor--
		} else {
			// 移到最上面时聚焦搜索框
			p.searchFocus = true
			p.searchInput.Focus()
		}
	case tea.KeyDown:
		if p.cursor < len(p.items)-1 {
			p.cursor++
		}
	case tea.KeyEnter:
		if p.cursor >= 0 && p.cursor < len(p.items) {
			item := p.items[p.cursor]
			switch item.Type {
			case "series":
				if item.Series != nil {
					return p, PushPage(NewSeriesPage(p.client, item.Series))
				}
				if item.SeriesSlug != "" {
					// 搜索结果中的 series，创建只有 slug 的 SeriesPage，加载在页面内完成
					return p, PushPage(NewSeriesPage(p.client, &pm.Series{Slug: item.SeriesSlug}))
				}
			case "event":
				return p, PushPage(NewEventPage(p.client, item.Event))
			}
		}
	}

	return p, nil
}

// View 渲染
func (p *HomePage) View() string {
	var b strings.Builder

	// 标题
	title := StyleTitle.Render("PolyMarket Browser")
	b.WriteString(title)
	b.WriteString("\n\n")

	// 搜索框
	if p.searchFocus {
		b.WriteString(p.searchInput.View())
	} else {
		b.WriteString(p.searchInput.View())
	}
	b.WriteString("\n")

	// 分隔线
	b.WriteString(StyleSeparator.Render(strings.Repeat("─", p.width)))
	b.WriteString("\n")

	// 列表
	p.mu.Lock()
	items := p.items
	loading := p.loading
	err := p.err
	cursor := p.cursor
	p.mu.Unlock()

	if loading {
		b.WriteString("\n")
		b.WriteString(StyleLoading.Render("  Loading..."))
	} else if err != nil {
		b.WriteString("\n")
		b.WriteString(StyleDisconnected.Render(fmt.Sprintf("  Error: %v", err)))
	} else if len(items) == 0 {
		b.WriteString("\n")
		b.WriteString(StyleEmpty.Render("  No results found"))
	} else {
		for i, item := range items {
			line := p.renderItem(item, i == cursor)
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// 底部提示
	b.WriteString(StyleSeparator.Render(strings.Repeat("─", p.width)))
	b.WriteString("\n")
	hint := "↑↓ Navigate  Tab Search  Enter Open  Esc Quit"
	b.WriteString(StyleDim.Render(hint))

	return b.String()
}

func (p *HomePage) renderItem(item homeItem, selected bool) string {
	var parts []string

	// 类型标签
	switch item.Type {
	case "series":
		parts = append(parts, StyleSeriesTag.Render("[Series]"))
	case "event":
		parts = append(parts, StyleEventTag.Render("[Event]"))
	}

	// 标题
	title := item.title()
	if len(title) > 50 {
		title = title[:47] + "..."
	}
	parts = append(parts, title)

	// 交易量
	vol := formatVolume(item.volume24hr())
	if vol != "0" {
		parts = append(parts, StyleDim.Render("$"+vol))
	}

	line := "  " + strings.Join(parts, " ")
	if selected {
		return StyleCursor.Render(">") + line[1:]
	}
	return line
}

// loadHomeItems 加载首页数据
func (p *HomePage) loadHomeItems() tea.Msg {
	p.logger = logr.FromContextOrDiscard(p.ctx)

	var series []pm.Series
	var events []pm.Event
	var seriesErr, eventsErr error
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		series, seriesErr = p.client.ListSeries(p.ctx, &pm.ListSeriesRequest{
			Limit:  20,
			Order:  "volume24hr",
			Closed: &boolFalse,
		})
	}()
	go func() {
		defer wg.Done()
		events, eventsErr = p.client.ListEvents(p.ctx, &pm.ListEventsRequest{
			Limit:  20,
			Order:  "volume24hr",
			Active: &boolTrue,
		})
	}()
	wg.Wait()

	if seriesErr != nil {
		return homeItemsLoadedMsg{err: fmt.Errorf("list series: %w", seriesErr)}
	}
	if eventsErr != nil {
		return homeItemsLoadedMsg{err: fmt.Errorf("list events: %w", eventsErr)}
	}

	// 合并去重
	items := make([]homeItem, 0, len(series)+len(events))
	seriesSlugSet := make(map[string]struct{})

	for i := range series {
		items = append(items, homeItem{Type: "series", Series: &series[i]})
		if series[i].Slug != "" {
			seriesSlugSet[series[i].Slug] = struct{}{}
		}
	}

	for i := range events {
		// 跳过已属于已展示 series 的事件
		skip := false
		for _, s := range events[i].Series {
			if _, ok := seriesSlugSet[s.Slug]; ok {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		// 跳过没有独立标题的事件（可能是 series 的子事件）
		if events[i].Title == "" {
			continue
		}
		items = append(items, homeItem{Type: "event", Event: &events[i]})
	}

	return homeItemsLoadedMsg{items: items}
}

// doSearch 执行搜索
func (p *HomePage) doSearch(query string) tea.Cmd {
	return func() tea.Msg {
		result, err := p.client.Search(p.ctx, &pm.SearchRequest{
			Query: query,
			Limit: 20,
		})
		if err != nil {
			return homeItemsLoadedMsg{err: fmt.Errorf("search: %w", err)}
		}

		items := make([]homeItem, 0, len(result.Events))
		seenSeries := make(map[string]struct{})

		for i := range result.Events {
			event := &result.Events[i]
			// 搜索 API 返回的事件没有 Series 数组，但有 SeriesSlug 字段
			if event.SeriesSlug != "" {
				if _, ok := seenSeries[event.SeriesSlug]; !ok {
					seenSeries[event.SeriesSlug] = struct{}{}
					// 用 series slug 作为标识，点击时通过 API 加载 series 详情
					items = append(items, homeItem{Type: "series", SeriesSlug: event.SeriesSlug})
				}
			} else if len(event.Series) > 0 {
				// 如果有 Series 数组（其他 API 返回的），用 series slug 去重
				for _, s := range event.Series {
					if s.Slug != "" {
						if _, ok := seenSeries[s.Slug]; !ok {
							seenSeries[s.Slug] = struct{}{}
							seriesCopy := s
							items = append(items, homeItem{Type: "series", Series: &seriesCopy})
						}
					}
				}
			} else {
				// 独立事件（没有 series）
				items = append(items, homeItem{Type: "event", Event: event})
			}
		}

		return homeItemsLoadedMsg{items: items}
	}
}

// homeItemsLoadedMsg 首页数据加载完成消息
type homeItemsLoadedMsg struct {
	items []homeItem
	err   error
}

var boolFalse = false
var boolTrue = true

// 确保 HomePage 实现了 Page 接口
var _ Page = (*HomePage)(nil)
var _ tea.Model = (*HomePage)(nil)
