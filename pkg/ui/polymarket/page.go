package polymarket

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

// Page 页面接口
//
// 每个 Page 实现自己的 Update 和 View，由 Browser 管理页面栈
type Page interface {
	tea.Model

	// Type 返回页面类型标识
	Type() string

	// OnPush 当页面被 push 到栈时调用
	OnPush()

	// OnPop 当页面从栈中 pop 时调用（用于清理资源）
	OnPop()
}

// ContextSetter 页面可能需要 context 来启动异步操作
type ContextSetter interface {
	SetContext(ctx context.Context)
}
