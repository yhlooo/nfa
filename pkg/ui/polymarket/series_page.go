package polymarket

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	pm "github.com/yhlooo/nfa/pkg/polymarket"
)

// SeriesPage Series 详情页
type SeriesPage struct {
	ctx    context.Context
	client pm.GammaAPIClient

	series         *pm.Series
	filteredEvents []pm.Event
	nearestIdx     int // 最近活跃事件在 filteredEvents 中的索引，-1 表示无
	cursor         int
	width          int
	loading        bool
	loadErr        error
}

// NewSeriesPage 创建 Series 详情页
func NewSeriesPage(client pm.GammaAPIClient, series *pm.Series) *SeriesPage {
	return &SeriesPage{
		client: client,
		series: series,
		width:  80,
	}
}

// SetContext 设置 context
func (p *SeriesPage) SetContext(ctx context.Context) {
	p.ctx = ctx
}

func (p *SeriesPage) Type() string { return "series" }
func (p *SeriesPage) OnPush()      {}
func (p *SeriesPage) OnPop()       {}

// Init 初始化
func (p *SeriesPage) Init() tea.Cmd {
	// 如果 series 已有完整数据（从首页进入），在后台过滤事件
	if p.series != nil && len(p.series.Events) > 0 {
		p.loading = true
		return p.filterEventsAsync
	}
	// 如果 series 只有 slug（从搜索进入），需要加载完整数据
	if p.series != nil && p.series.Slug != "" && len(p.series.Events) == 0 {
		p.loading = true
		return p.loadSeriesAsync
	}
	return nil
}

// filterEventsAsync 后台过滤事件（不阻塞 UI）
func (p *SeriesPage) filterEventsAsync() tea.Msg {
	events, nearestIdx := filterSeriesEvents(p.series.Events)
	return seriesDataMsg{events: events, nearestIdx: nearestIdx}
}

// loadSeriesAsync 后台加载 series 完整数据
func (p *SeriesPage) loadSeriesAsync() tea.Msg {
	result, err := p.client.ListSeries(p.ctx, &pm.ListSeriesRequest{
		Slug:  []string{p.series.Slug},
		Limit: 1,
	})
	if err != nil {
		return seriesDataMsg{err: err}
	}
	if len(result) == 0 {
		return seriesDataMsg{err: fmt.Errorf("series not found")}
	}
	series := result[0]
	events, nearestIdx := filterSeriesEvents(series.Events)
	return seriesDataMsg{series: &series, events: events, nearestIdx: nearestIdx}
}

// seriesDataMsg series 数据加载完成消息
type seriesDataMsg struct {
	series     *pm.Series
	events     []pm.Event
	nearestIdx int
	err        error
}

// Update 处理更新
func (p *SeriesPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = m.Width
		return p, nil

	case seriesDataMsg:
		p.loading = false
		if m.err != nil {
			p.loadErr = m.err
		} else {
			if m.series != nil {
				p.series = m.series
			}
			p.filteredEvents = m.events
			p.nearestIdx = m.nearestIdx
		}
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
			if p.cursor < len(p.filteredEvents)-1 {
				p.cursor++
			}
		case tea.KeyEnter:
			if p.cursor >= 0 && p.cursor < len(p.filteredEvents) {
				event := p.filteredEvents[p.cursor]
				return p, PushPage(NewEventPage(p.client, &event))
			}
		}
	}

	return p, nil
}

// View 渲染
func (p *SeriesPage) View() string {
	var b strings.Builder

	if p.series == nil {
		b.WriteString(StyleEmpty.Render("  No series data"))
		return b.String()
	}

	// 标题
	b.WriteString(StyleSeriesTag.Render("[Series] "))
	b.WriteString(StyleTitle.Render(p.series.Title))
	b.WriteString("\n")
	if p.series.Slug != "" {
		b.WriteString(StyleDim.Render(p.series.Slug))
		b.WriteString("\n")
	}

	// 副标题
	if p.series.SubTitle != "" {
		b.WriteString(StyleSubtitle.Render(p.series.SubTitle))
		b.WriteString("\n")
	}

	// 描述
	if p.series.Description != "" {
		desc := truncate(p.series.Description, p.width)
		b.WriteString(StyleDescription.Render(desc))
		b.WriteString("\n")
	}

	// 交易量
	vol := formatVolume(p.series.Volume)
	vol24hr := formatVolume(p.series.Volume24hr)
	b.WriteString("\n")
	b.WriteString(StyleLabel.Render("Volume: "))
	b.WriteString(StyleValue.Render("$" + vol))
	if vol24hr != "0" {
		b.WriteString("  ")
		b.WriteString(StyleLabel.Render("24h: "))
		b.WriteString(StyleValue.Render("$" + vol24hr))
	}
	b.WriteString("\n")

	// 分隔线
	b.WriteString(StyleSeparator.Render(strings.Repeat("─", p.width)))
	b.WriteString("\n")

	// 事件列表
	totalEvents := len(p.series.Events)
	if totalEvents == 0 && p.loading {
		totalEvents = -1 // 显示为 loading
	}
	b.WriteString(StyleLabel.Render(fmt.Sprintf("Events (%d/%d)", len(p.filteredEvents), totalEvents)))
	b.WriteString("\n\n")

	if p.loading {
		b.WriteString(StyleLoading.Render("  Loading..."))
	} else if p.loadErr != nil {
		b.WriteString(StyleDisconnected.Render(fmt.Sprintf("  Error: %v", p.loadErr)))
	} else if len(p.filteredEvents) == 0 {
		b.WriteString(StyleEmpty.Render("  No events in this series"))
	} else {
		for i, event := range p.filteredEvents {
			selected := i == p.cursor
			line := p.renderEventItem(&event, i, selected)
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// 底部提示
	b.WriteString(StyleSeparator.Render(strings.Repeat("─", p.width)))
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("↑↓ Navigate  Enter Open  Esc Back"))

	return b.String()
}

func (p *SeriesPage) renderEventItem(event *pm.Event, idx int, selected bool) string {
	title := event.Title
	if title == "" {
		title = event.Slug
	}
	if len(title) > 55 {
		title = title[:52] + "..."
	}

	vol24hr := formatVolume(event.Volume24hr)
	line := "  " + title
	if vol24hr != "0" {
		line += "  " + StyleDim.Render("$"+vol24hr)
	}

	if selected {
		return StyleCursor.Render(">") + line[1:]
	}
	if idx == p.nearestIdx {
		return StyleNearestEvent.Render(line)
	}
	if event.Active && !event.Closed {
		return StyleActiveEvent.Render(line)
	}
	return StyleDim.Render(line)
}

// filterSeriesEvents 过滤 series 事件
// O(n) 遍历，不做全量排序。用固定大小数组维护 top-N，只对小数组排序。
// 返回过滤后的事件列表和最近活跃事件在结果中的索引（-1 表示无）
func filterSeriesEvents(events []pm.Event) ([]pm.Event, int) {
	const maxTotal = 6
	now := time.Now()

	type indexed struct {
		event pm.Event
		time  time.Time
		dist  time.Duration
	}

	var buf [maxTotal]indexed
	bufN := 0
	maxDist := time.Duration(1<<63 - 1)

	for i := range events {
		e := &events[i]
		t := parseTime(e.EndDate)
		if t.IsZero() {
			continue
		}

		dist := t.Sub(now)
		if dist < 0 {
			dist = -dist
		}

		if bufN < maxTotal {
			buf[bufN] = indexed{event: *e, time: t, dist: dist}
			bufN++
			if dist > maxDist {
				maxDist = dist
			}
		} else if dist < maxDist {
			for j := 0; j < maxTotal; j++ {
				if buf[j].dist == maxDist {
					buf[j] = indexed{event: *e, time: t, dist: dist}
					break
				}
			}
			maxDist = 0
			for j := 0; j < maxTotal; j++ {
				if buf[j].dist > maxDist {
					maxDist = buf[j].dist
				}
			}
		}
	}

	collected := buf[:bufN]
	sort.Slice(collected, func(i, j int) bool {
		return collected[i].time.Before(collected[j].time)
	})

	result := make([]pm.Event, bufN)
	for i, c := range collected {
		result[i] = c.event
	}

	nearestIdx := -1
	minDist := time.Duration(1<<63 - 1)
	for i, e := range result {
		if e.Active && !e.Closed {
			endTime := parseTime(e.EndDate)
			if !endTime.IsZero() {
				dist := endTime.Sub(now)
				if dist < 0 {
					dist = -dist
				}
				if dist < minDist {
					minDist = dist
					nearestIdx = i
				}
			}
		}
	}

	return result, nearestIdx
}

// parseTime 解析事件时间
func parseTime(s string) time.Time {
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

var _ Page = (*SeriesPage)(nil)
var _ tea.Model = (*SeriesPage)(nil)
