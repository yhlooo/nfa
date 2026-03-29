package polymarkettrading

import (
	"github.com/charmbracelet/lipgloss"
)

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

	// StyleLabel 标签样式
	StyleLabel = lipgloss.NewStyle().Bold(true)

	// StyleValue 数值样式
	StyleValue = lipgloss.NewStyle()

	// StyleDim 暗淡样式
	StyleDim = lipgloss.NewStyle().Faint(true)

	// StyleBid 买入价样式（绿色）
	StyleBid = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))

	// StyleAsk 卖出价样式（红色）
	StyleAsk = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	// StyleConnected 连接状态样式（绿色）
	StyleConnected = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))

	// StyleDisconnected 断连状态样式（橙色）
	StyleDisconnected = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

	// StyleBuy 买入样式（绿色）
	StyleBuy = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))

	// StyleSell 卖出样式（红色）
	StyleSell = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	// StyleYes Yes 样式
	StyleYes = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

	// StyleNo No 样式
	StyleNo = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

	// StyleProfit 盈利样式（绿色）
	StyleProfit = lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Bold(true)

	// StyleLoss 亏损样式（红色）
	StyleLoss = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)

	// StylePrice 价格样式
	StylePrice = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
)
