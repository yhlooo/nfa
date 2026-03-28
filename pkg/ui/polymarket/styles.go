package polymarket

import "github.com/charmbracelet/lipgloss"

// 共享样式定义
var (
	// StyleTitle 标题样式
	StyleTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("36"))

	// StyleSubtitle 副标题样式
	StyleSubtitle = lipgloss.NewStyle().Faint(true)

	// StyleDescription 描述样式
	StyleDescription = lipgloss.NewStyle().Faint(true)

	// StyleSeparator 分隔线样式
	StyleSeparator = lipgloss.NewStyle().Faint(true)

	// StyleTag 类型标签样式
	StyleTag = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))

	// StyleSeriesTag Series 标签样式
	StyleSeriesTag = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))

	// StyleEventTag Event 标签样式
	StyleEventTag = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75"))

	// StyleListItem 列表项样式
	StyleListItem = lipgloss.NewStyle().PaddingLeft(2)

	// StyleCursor 光标样式
	StyleCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)

	// StyleBid 买入价样式（绿色）
	StyleBid = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))

	// StyleAsk 卖出价样式（红色）
	StyleAsk = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	// StyleConnected 连接状态样式（绿色）
	StyleConnected = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))

	// StyleDisconnected 断连状态样式（橙色）
	StyleDisconnected = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

	// StyleLoading 加载中样式
	StyleLoading = lipgloss.NewStyle().Faint(true).Italic(true)

	// StyleEmpty 空状态样式
	StyleEmpty = lipgloss.NewStyle().Faint(true)

	// StyleValue 数值样式
	StyleValue = lipgloss.NewStyle()

	// StyleLabel 标签样式
	StyleLabel = lipgloss.NewStyle().Bold(true)

	// StyleDim 暗淡样式
	StyleDim = lipgloss.NewStyle().Faint(true)

	// StyleActiveEvent 活跃事件样式（蓝色）
	StyleActiveEvent = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

	// StyleNearestEvent 最近活跃事件样式（高亮蓝色）
	StyleNearestEvent = lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
)
