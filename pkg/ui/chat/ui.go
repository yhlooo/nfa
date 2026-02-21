package chat

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coder/acp-go-sdk"
	"github.com/firebase/genkit/go/ai"
	"github.com/go-logr/logr"

	"github.com/yhlooo/nfa/pkg/acputil"
	"github.com/yhlooo/nfa/pkg/agents"
	"github.com/yhlooo/nfa/pkg/configs"
	"github.com/yhlooo/nfa/pkg/models"
)

type viewState string

const (
	viewStateInput       viewState = "input"
	viewStateModelSelect viewState = "model_select"
)

type ModelType string

const (
	ModelTypeMain   ModelType = "main"
	ModelTypeFast   ModelType = "fast"
	ModelTypeVision ModelType = "vision"
)

// Options UI 运行选项
type Options struct {
	AgentClientIn         io.Reader
	AgentClientOut        io.Writer
	InitialPrompt         string
	AutoExitAfterResponse bool
}

// NewChatUI 创建对话 UI
func NewChatUI(opts Options) *ChatUI {
	ui := &ChatUI{
		modelUsageStyle:       lipgloss.NewStyle().Faint(true).Align(lipgloss.Right).PaddingRight(2),
		initialPrompt:         opts.InitialPrompt,
		autoExitAfterResponse: opts.AutoExitAfterResponse,
		viewState:             viewStateInput,
	}
	ui.conn = acp.NewClientSideConnection(ui, opts.AgentClientOut, opts.AgentClientIn)
	return ui
}

// ChatUI 对话 UI
type ChatUI struct {
	ctx    context.Context
	logger logr.Logger
	p      *tea.Program

	vp    MessageViewport
	input *InputBox

	acputil.NopFS
	acputil.NopTerminal
	modelUsageStyle lipgloss.Style

	conn                  *acp.ClientSideConnection
	cwd                   string
	sessionID             acp.SessionId
	curModels             models.Models
	modelUsage            ai.GenerationUsage
	initialPrompt         string
	autoExitAfterResponse bool

	// Model selection
	cfgPath       string
	viewState     viewState
	modelSelector *ModelSelector
}

var _ tea.Model = (*ChatUI)(nil)
var _ acp.Client = (*ChatUI)(nil)

// Run 运行 UI
func (ui *ChatUI) Run(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx)

	ui.ctx = ctx
	ui.logger = logger

	ui.vp = NewMessageViewport()
	ui.modelSelector = NewModelSelector()

	ui.cfgPath = configs.ConfigPathFromContext(ctx)

	// 初始化 agent
	if err := ui.initAgent(ctx); err != nil {
		return fmt.Errorf("initialize agent error: %w", err)
	}

	ui.input = NewInputBox(ctx, []SelectorOption{
		//{Name: "mcp", Description: "Manage MCP servers"},
		{Name: "clear", Description: "Start a fresh conversation"},
		{Name: "model", Description: "Set the AI model for NFA"},
		{Name: "exit", Description: "Exit the NFA"},
	})

	p := tea.NewProgram(ui, tea.WithContext(ctx))
	ui.p = p
	_, err := p.Run()
	return err
}

// Init 开始运行 UI 的第一个操作
func (ui *ChatUI) Init() tea.Cmd {
	cmds := []tea.Cmd{
		ui.newSession,
		ui.printHello(),
		textarea.Blink,
	}
	if ui.initialPrompt != "" {
		// 在 session 创建后发送初始 prompt
		cmds = append(cmds, ui.newPrompt(ui.initialPrompt))
	}
	return tea.Sequence(cmds...)
}

// Update 处理更新事件
func (ui *ChatUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// 根据 viewState 路由消息
	switch ui.viewState {
	case viewStateInput:
		return ui.updateInInputState(msg)
	case viewStateModelSelect:
		return ui.updateInModelSelectState(msg)
	}
	return ui, nil
}

// updateInInputState 处理输入状态
func (ui *ChatUI) updateInInputState(msg tea.Msg) (tea.Model, tea.Cmd) {
	logger := ui.logger

	var inputCmd tea.Cmd
	ui.input, inputCmd = ui.input.Update(msg)
	var vpCmd tea.Cmd
	ui.vp, vpCmd = ui.vp.Update(msg)

	cmds := []tea.Cmd{inputCmd, vpCmd}

	switch typedMsg := msg.(type) {
	case tea.WindowSizeMsg:
		ui.modelUsageStyle = ui.modelUsageStyle.Width(typedMsg.Width)
		ui.logger.Info(fmt.Sprintf("resize message: width: %d, height: %d", typedMsg.Width, typedMsg.Height))

	case tea.KeyMsg:
		ui.logger.Info(fmt.Sprintf("key message: %q", typedMsg.String()))
		switch typedMsg.Type {
		case tea.KeyEnter:
			if !ui.vp.AgentProcessing() && !ui.input.MultiLineMode() {
				content := strings.TrimRight(ui.input.Value(), "\n")
				ui.input.Reset()
				if content != "" {
					switch content {
					case "/exit":
						return ui, tea.Quit
					case "/model", "/model :main":
						return ui, ui.enterModelSelectMode(ModelTypeMain)
					case "/model :fast":
						return ui, ui.enterModelSelectMode(ModelTypeFast)
					case "/model :vision":
						return ui, ui.enterModelSelectMode(ModelTypeVision)
					default:
						// 检查是否是 /model 开头的直接设置命令
						if modelType, modelName, ok := ui.handleDirectModelSet(content); ok {
							cmds = append(cmds, tea.Printf(
								"\033[34m✓ set %s model: %s\033[0m",
								modelType, modelName,
							))
						} else {
							cmds = append(cmds, ui.newPrompt(content))
						}
					}
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

	case acp.PromptResponse:
		ui.modelUsage = agents.GetMetaCurrentModelUsageValue(typedMsg.Meta)
		cmds = append(cmds, ui.vp.Flush())
		if ui.autoExitAfterResponse {
			cmds = append(cmds, tea.Quit)
		}
	case acp.SessionNotification:
		ui.modelUsage = agents.GetMetaCurrentModelUsageValue(typedMsg.Meta)
		cmds = append(cmds, ui.vp.Flush())
	case acp.PromptRequest:
		cmds = append(cmds, ui.vp.Flush())

	case QuitError:
		logger.Error(typedMsg.Error, "error")
		cmds = append(cmds, tea.Quit)

	case error:
		logger.Error(typedMsg, "error")
	}

	cmds = append(cmds, ui.updateComponents()...)

	return ui, tea.Batch(cmds...)
}

// updateInModelSelectState 处理模型选择状态
func (ui *ChatUI) updateInModelSelectState(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var vpCmd tea.Cmd
	ui.vp, vpCmd = ui.vp.Update(msg)
	cmds = append(cmds, vpCmd)

	var selectorCmd tea.Cmd
	ui.modelSelector, selectorCmd = ui.modelSelector.Update(msg)
	cmds = append(cmds, selectorCmd)

	switch typedMsg := msg.(type) {
	case tea.KeyMsg:
		switch typedMsg.Type {
		case tea.KeyEnter:
			var modelType ModelType
			modelName := ""
			modelType, modelName, ui.curModels = ui.modelSelector.GetSelectedModels()
			if err := configs.SaveDefaultModels(ui.cfgPath, ui.curModels); err != nil {
				cmds = append(cmds, func() tea.Msg {
					return fmt.Errorf("failed to save model: %w", err)
				})
			} else {
				cmds = append(cmds, tea.Printf("\033[34m✓ set %s model: %s\033[0m", modelType, modelName))
			}
			cmds = append(cmds, ui.exitModelSelectMode())
		case tea.KeyEsc:
			// 取消选择
			cmds = append(cmds, ui.exitModelSelectMode())
		}
	}

	cmds = append(cmds, ui.updateComponents()...)

	return ui, tea.Batch(cmds...)
}

// updateComponents 根据状态更新组件
func (ui *ChatUI) updateComponents() []tea.Cmd {
	var cmds []tea.Cmd

	// 设置输入状态
	if ui.vp.AgentProcessing() {
		ui.input.Blur()
	} else {
		if !ui.input.Focused() {
			cmds = append(cmds, ui.input.Focus())
		}
	}

	return cmds
}

// View 渲染显示内容
func (ui *ChatUI) View() string {
	vpView := ui.vp.View()
	if vpView != "" {
		vpView += "\n"
	}

	var bottomView string
	switch ui.viewState {
	case viewStateInput:
		if ui.input.Focused() {
			bottomView = ui.input.View()
		}
	case viewStateModelSelect:
		bottomView = ui.modelSelector.View()
	}

	modelUsageView := ""
	if ui.modelUsage.InputTokens != 0 {
		modelUsageView += fmt.Sprintf("↑ %s", intWithSeparator(ui.modelUsage.InputTokens))
	}
	if out := ui.modelUsage.ThoughtsTokens + ui.modelUsage.ThoughtsTokens; out != 0 {
		modelUsageView += fmt.Sprintf(" | ↓ %s", intWithSeparator(out))
	}
	if modelUsageView != "" {
		modelUsageView = "Token Usage: " + strings.TrimPrefix(modelUsageView, " | ")
	}

	return fmt.Sprintf(
		`%s
%s%s

`,
		vpView,
		bottomView,
		ui.modelUsageStyle.Render(modelUsageView),
	)
}

// printHello 输出欢迎信息
func (ui *ChatUI) printHello() tea.Cmd {
	return func() tea.Msg {
		return tea.Printf(`
╭─────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                 │ `+"\033[1;32m"+`Tips:`+"\033[0m"+`                                                         │
│                                 │ ...                                                           │
│                                 │ ...                                                           │
│`+"\033[1;34m"+`         _   __ ______ ___       `+"\033[0m"+`│ ...                                                           │
│`+"\033[1;34m"+`        / | / // ____//   |      `+"\033[0m"+`│ ...                                                           │
│`+"\033[1;34m"+`       /  |/ // /_   / /| |      `+"\033[0m"+`│                                                               │
│`+"\033[1;34m"+`      / /|  // __/  / ___ |      `+"\033[0m"+`│ `+"\033[1;33m"+`NOTE: Any output should not be construed as financial advice.`+"\033[0m"+` │
│`+"\033[1;34m"+`     /_/ |_//_/    /_/  |_|      `+"\033[0m"+`│ ───────────────────────────────────────────────────────────── │
│                                 │ `+"\033[1;32m"+`Model`+"\033[0m"+`    %-52s │
│                                 │          %-52s │
│                                 │          %-52s │
│                                 │ `+"\033[1;32m"+`Session`+"\033[0m"+`  %-52s │
╰─────────────────────────────────────────────────────────────────────────────────────────────────╯
`,
			ui.curModels.Main+" (main)",
			ui.curModels.Fast+" (fast)",
			ui.curModels.Vision+" (vision)",
			ui.sessionID,
		)()
	}
}

// intWithSeparator 每 step 位带分隔符 sep 的表示整数的字符串
func intWithSeparator(v int) string {
	vStr := strconv.FormatInt(int64(v), 10)

	// 暂时去除负号
	sign := ""
	if v < 0 {
		sign = "-"
		vStr = vStr[1:]
	}

	divided := make([]string, (len(vStr)+2)/3)

	j := len(divided) - 1
	for i := len(vStr); i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}

		divided[j] = vStr[start:i]
		j--
	}

	return sign + strings.Join(divided, ",")
}
