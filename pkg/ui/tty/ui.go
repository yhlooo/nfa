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

	input       textarea.Model
	inPrompting bool
	buffer      string

	acputil.NopFS
	acputil.NopTerminal

	conn      *acp.ClientSideConnection
	sessionID acp.SessionId
}

var _ tea.Model = (*ChatUI)(nil)
var _ acp.Client = (*ChatUI)(nil)

// Run 运行 UI
func (ui *ChatUI) Run(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx)
	ui.ctx = ctx
	ui.logger = logger

	ui.input = textarea.New()
	ui.input.Prompt = "> "
	ui.input.CharLimit = 1024
	ui.input.ShowLineNumbers = false
	ui.input.SetWidth(30)
	ui.input.SetHeight(1)
	ui.input.Focus()

	p := tea.NewProgram(ui, tea.WithContext(ctx))
	ui.p = p
	_, err := p.Run()
	return err
}

// Init 开始运行 UI 的第一个操作
func (ui *ChatUI) Init() tea.Cmd {
	return tea.Sequence(
		ui.initAgent,
		tea.Println(`
╭─────────────────────────────────╮
│         _   __ ______ ___       │
│        / | / // ____//   |      │
│       /  |/ // /_   / /| |      │
│      / /|  // __/  / ___ |      │
│     /_/ |_//_/    /_/  |_|      │
│                                 │
╰─────────────────────────────────╯

`),
		textarea.Blink,
	)
}

// Update 处理更新事件
func (ui *ChatUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logger := ui.logger

	var inputCmd tea.Cmd
	ui.input, inputCmd = ui.input.Update(msg)

	cmds := []tea.Cmd{inputCmd}

	switch typedMsg := msg.(type) {
	case tea.WindowSizeMsg:
		ui.width = typedMsg.Width
		ui.input.SetWidth(typedMsg.Width)
	case tea.KeyMsg:
		switch typedMsg.Type {
		case tea.KeyEnter:
			if !ui.inPrompting {
				if !ui.input.KeyMap.InsertNewline.Enabled() {
					content := strings.TrimRight(ui.input.Value(), "\n")
					ui.input.Reset()
					if content != "" {
						cmds = append(cmds, tea.Sequence(
							tea.Printf("> %s", content),
							ui.newPrompt(content),
						))
						ui.inPrompting = true
					}
				}
			}

		case tea.KeyTab:
			ui.input.KeyMap.InsertNewline.SetEnabled(true)
			ui.input.SetHeight(5)
		case tea.KeyEsc:
			content := strings.Trim(ui.input.Value(), "\n")

			if content != "" && !ui.inPrompting {
				cmds = append(cmds, tea.Sequence(
					tea.Printf("> %s", content),
					ui.newPrompt(content),
				))
				ui.inPrompting = true
			}

			if content == "" || !ui.inPrompting {
				ui.input.KeyMap.InsertNewline.SetEnabled(false)
				ui.input.SetHeight(1)
				ui.input.Reset()
			}

		case tea.KeyCtrlC:
			if !ui.inPrompting {
				return ui, tea.Quit
			}
			cmds = append(cmds, ui.cancelPrompt)
		}

	case acp.SessionUpdate:
		switch {
		case typedMsg.AgentMessageChunk != nil:
			content := typedMsg.AgentMessageChunk.Content
			switch {
			case content.Text != nil:
				ui.buffer += content.Text.Text
			}
		}

	case PromptResult:
		logger.Error(typedMsg.Error, "prompt error")
		ui.inPrompting = false
		if typedMsg.Error != nil {
			cmds = append(cmds, tea.Println(ui.buffer+"\n"+typedMsg.Error.Error()))
		} else {
			cmds = append(cmds, tea.Println(ui.buffer))
		}
		ui.buffer = ""

	case QuitError:
		logger.Error(typedMsg, "error")
		cmds = append(cmds,
			tea.Println(typedMsg.Error()),
			tea.Quit,
		)

	case error:
		logger.Error(typedMsg, "error")
		cmds = append(cmds, tea.Println(typedMsg.Error()))
	}

	return ui, tea.Batch(cmds...)
}

// View 渲染显示内容
func (ui *ChatUI) View() string {
	return fmt.Sprintf(`%s

%s
%s
%s`,
		ui.buffer,
		strings.Repeat("─", ui.width),
		ui.input.View(),
		strings.Repeat("─", ui.width))
}
