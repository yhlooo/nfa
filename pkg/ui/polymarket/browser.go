package polymarket

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yhlooo/nfa/pkg/polymarket"
)

// Options 浏览器选项
type Options struct {
	Client *polymarket.Client
}

// Browser PolyMarket 浏览器主模型
type Browser struct {
	ctx    context.Context
	client *polymarket.Client

	stack []Page
	width int
}

// NewBrowser 创建浏览器
func NewBrowser(opts Options) *Browser {
	return &Browser{
		client: opts.Client,
		width:  80,
	}
}

// Run 运行浏览器
func (b *Browser) Run(ctx context.Context) error {
	b.ctx = ctx

	// push 首页
	home := NewHomePage(b.client)
	// 首页也需要 context 来加载数据
	home.SetContext(ctx)
	home.OnPush()
	b.stack = []Page{home}

	p := tea.NewProgram(b, tea.WithContext(ctx))
	_, err := p.Run()
	return err
}

var _ tea.Model = (*Browser)(nil)

// Init 初始化
func (b *Browser) Init() tea.Cmd {
	if len(b.stack) > 0 {
		return b.stack[0].Init()
	}
	return nil
}

// Update 处理更新
func (b *Browser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		b.width = m.Width
		// 传递给所有页面
		for _, page := range b.stack {
			page.Update(m)
		}
		return b, nil

	case tea.KeyMsg:
		// 全局按键
		switch m.Type {
		case tea.KeyCtrlC:
			return b, tea.Quit
		case tea.KeyEsc:
			if len(b.stack) <= 1 {
				return b, tea.Quit
			}
			// Pop 当前页面
			current := b.stack[len(b.stack)-1]
			current.OnPop()
			b.stack = b.stack[:len(b.stack)-1]
			return b, nil
		}

		// 委托给当前页面
		if len(b.stack) > 0 {
			top := b.stack[len(b.stack)-1]
			model, cmd := top.Update(msg)
			if page, ok := model.(Page); ok {
				b.stack[len(b.stack)-1] = page
			}
			return b, cmd
		}

	case pushPageMsg:
		// 在 push 前设置 context
		if cs, ok := m.page.(ContextSetter); ok {
			cs.SetContext(b.ctx)
		}
		b.stack = append(b.stack, m.page)
		m.page.OnPush()
		// 触发新页面的 Init（返回异步加载 Cmd）
		initCmd := m.page.Init()
		if initCmd != nil {
			return b, initCmd
		}
		return b, nil
	}

	// 委托给当前页面
	if len(b.stack) > 0 {
		top := b.stack[len(b.stack)-1]
		model, cmd := top.Update(msg)
		if page, ok := model.(Page); ok {
			b.stack[len(b.stack)-1] = page
		}
		return b, cmd
	}

	return b, nil
}

// View 渲染
func (b *Browser) View() string {
	if len(b.stack) == 0 {
		return ""
	}
	return b.stack[len(b.stack)-1].View()
}

// PushPage 推入页面（作为 Cmd 使用）
func PushPage(page Page) tea.Cmd {
	return func() tea.Msg {
		return pushPageMsg{page: page}
	}
}

// pushPageMsg 推入页面消息
type pushPageMsg struct {
	page Page
}

// --- 辅助函数 ---

// formatVolume 格式化交易量
func formatVolume(v float64) string {
	switch {
	case v >= 1_000_000:
		s := fmt.Sprintf("%.1f", v/1_000_000)
		s = strings.TrimSuffix(strings.TrimSuffix(s, ".0"), "0")
		return s + "M"
	case v >= 1_000:
		s := fmt.Sprintf("%.1f", v/1_000)
		s = strings.TrimSuffix(strings.TrimSuffix(s, ".0"), "0")
		return s + "K"
	default:
		return fmt.Sprintf("%.0f", v)
	}
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
