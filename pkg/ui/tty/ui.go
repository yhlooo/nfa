package tty

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coder/acp-go-sdk"
	"github.com/go-logr/logr"

	"github.com/yhlooo/nfa/pkg/acputil"
)

// Options UI 运行选项
type Options struct {
	AgentClientIn  io.Reader
	AgentClientOut io.Writer
}

// NewChatUI 创建对话 UI
func NewChatUI(opts Options) *ChatUI {
	ui := &ChatUI{}
	ui.conn = acp.NewClientSideConnection(ui, opts.AgentClientOut, opts.AgentClientIn)
	return ui
}

// ChatUI 对话 UI
type ChatUI struct {
	ctx    context.Context
	logger logr.Logger
	p      *tea.Program

	width int

	input textarea.Model
	vp    MessageViewport

	acputil.NopFS
	acputil.NopTerminal

	conn            *acp.ClientSideConnection
	cwd             string
	sessionID       acp.SessionId
	defaultModel    string
	availableModels []string
}

var _ tea.Model = (*ChatUI)(nil)
var _ acp.Client = (*ChatUI)(nil)

// Run 运行 UI
func (ui *ChatUI) Run(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx)

	if err := ui.initAgent(ctx); err != nil {
		return fmt.Errorf("initialize agent error: %w", err)
	}

	ui.ctx = ctx
	ui.logger = logger

	ui.input = textarea.New()
	ui.input.Prompt = "> "
	ui.input.CharLimit = 1024
	ui.input.ShowLineNumbers = false
	ui.input.SetWidth(30)
	ui.input.SetHeight(1)
	ui.input.Focus()

	ui.vp = NewMessageViewport()

	p := tea.NewProgram(ui, tea.WithContext(ctx))
	ui.p = p
	_, err := p.Run()
	return err
}

// Init 开始运行 UI 的第一个操作
func (ui *ChatUI) Init() tea.Cmd {
	return tea.Sequence(
		ui.newSession,
		ui.printHello(),
		textarea.Blink,
	)
}

// Update 处理更新事件
func (ui *ChatUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logger := ui.logger

	var inputCmd tea.Cmd
	ui.input, inputCmd = ui.input.Update(msg)
	var vpCmd tea.Cmd
	ui.vp, vpCmd = ui.vp.Update(msg)

	cmds := []tea.Cmd{
		inputCmd,
		vpCmd,
	}

	switch typedMsg := msg.(type) {
	case tea.WindowSizeMsg:
		ui.width = typedMsg.Width
		ui.input.SetWidth(typedMsg.Width)
		ui.vp.SetWidth(typedMsg.Width)
	case tea.KeyMsg:
		switch typedMsg.Type {
		case tea.KeyEnter:
			if !ui.vp.AgentProcessing() {
				content := strings.TrimRight(ui.input.Value(), "\n")
				ui.input.Reset()
				if content != "" {
					cmds = append(cmds, ui.newPrompt(content))
				}
			}

		case tea.KeyEsc:
			if ui.vp.AgentProcessing() {
				cmds = append(cmds, ui.cancelPrompt)
			}

		case tea.KeyCtrlC:
			if !ui.vp.AgentProcessing() {
				return ui, tea.Quit
			}
			cmds = append(cmds, ui.cancelPrompt)
		}

	case acp.PromptRequest, acp.PromptResponse, acp.SessionNotification:
		cmds = append(cmds, ui.vp.Flush())

	case QuitError:
		logger.Error(typedMsg.Error, "error")
		cmds = append(cmds, tea.Quit)

	case error:
		logger.Error(typedMsg, "error")
	}

	return ui, tea.Batch(cmds...)
}

// View 渲染显示内容
func (ui *ChatUI) View() string {
	vpView := ui.vp.View()
	if vpView != "" {
		vpView += "\n"
	}
	return fmt.Sprintf(`%s
%s
%s
%s`,
		vpView,
		strings.Repeat("─", ui.width),
		ui.input.View(),
		strings.Repeat("─", ui.width))
}

// printHello 输出欢迎信息
func (ui *ChatUI) printHello() tea.Cmd {
	return func() tea.Msg {
		return tea.Printf(`
╭─────────────────────────────────╮
│         _   __ ______ ___       │
│        / | / // ____//   |      │
│       /  |/ // /_   / /| |      │
│      / /|  // __/  / ___ |      │
│     /_/ |_//_/    /_/  |_|      │
│                                 │
╰─────────────────────────────────╯
Session : %s
Model   : %s
`, ui.sessionID, ui.defaultModel)()
	}
}
