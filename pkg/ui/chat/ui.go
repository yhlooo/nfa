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
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/yhlooo/nfa/pkg/acputil"
	"github.com/yhlooo/nfa/pkg/agents"
	"github.com/yhlooo/nfa/pkg/configs"
	i18nutil "github.com/yhlooo/nfa/pkg/i18n"
	"github.com/yhlooo/nfa/pkg/models"
	"github.com/yhlooo/nfa/pkg/otter"
	"github.com/yhlooo/nfa/pkg/skills"
	"github.com/yhlooo/nfa/pkg/version"
)

type viewState string

const (
	viewStateInput       viewState = "input"
	viewStateModelSelect viewState = "model_select"
)

type ModelType string

const (
	ModelTypePrimary ModelType = "primary"
	ModelTypeLight   ModelType = "light"
	ModelTypeVision  ModelType = "vision"
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
		width:                 80,
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
	width int

	acputil.NopFS
	acputil.NopTerminal
	modelUsageStyle lipgloss.Style

	conn                  ACPClientSideConnection
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

	// Skills
	skills []skills.SkillMeta
}

var _ tea.Model = (*ChatUI)(nil)
var _ acp.Client = (*ChatUI)(nil)

// Run 运行 UI
func (ui *ChatUI) Run(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx)

	ui.ctx = ctx
	ui.logger = logger

	var err error
	ui.vp, err = NewMessageViewport(ctx)
	if err != nil {
		return err
	}

	ui.modelSelector = NewModelSelector()

	ui.cfgPath = configs.ConfigPathFromContext(ctx)

	// 初始化 agent
	if err := ui.initAgent(ctx); err != nil {
		return fmt.Errorf("initialize agent error: %w", err)
	}

	ui.input = NewInputBox(ctx, []SelectorOption{
		{Name: "clear", Description: i18nutil.TContext(ctx, MsgCmdDescClear)},
		{Name: "model", Description: i18nutil.TContext(ctx, MsgCmdDescModel)},
		{Name: "skills", Description: i18nutil.TContext(ctx, MsgCmdDescSkills)},
		{Name: "exit", Description: i18nutil.TContext(ctx, MsgCmdDescExit)},
	})

	p := tea.NewProgram(ui, tea.WithContext(ctx))
	ui.p = p
	_, err = p.Run()
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
		ui.width = typedMsg.Width
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
					case "/model", "/model :primary":
						return ui, ui.enterModelSelectMode(ModelTypePrimary)
					case "/model :light":
						return ui, ui.enterModelSelectMode(ModelTypeLight)
					case "/model :vision":
						return ui, ui.enterModelSelectMode(ModelTypeVision)
					case "/skills":
						cmds = append(cmds, ui.printSkillsList())
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
	bannerLines := strings.Split(otter.MustOtter(true, false, 1), "\n")
	if len(bannerLines) < 5 {
		return nil
	}

	i := len(bannerLines) - 6
	bannerLines[i] += fmt.Sprintf("\r\033[36C\033[1mNFA\033[2m v%s\033[0m", version.Version)
	i++
	bannerLines[i] += fmt.Sprintf("\r\033[36C\033[2m%s\033[0m", ui.curModels.GetPrimary())
	i++
	if ui.curModels.GetLight() != ui.curModels.GetPrimary() {
		bannerLines[i] += fmt.Sprintf("\r\033[36C\033[2m%s (light)\033[0m", ui.curModels.GetLight())
		i++
	}
	if ui.curModels.GetVision() != ui.curModels.GetPrimary() {
		bannerLines[i] += fmt.Sprintf("\r\033[36C\033[2m%s (vision)\033[0m", ui.curModels.GetVision())
		i++
	}
	bannerLines[i] += fmt.Sprintf("\r\033[36C\033[1;33m%s\033[0m", i18nutil.TContext(ui.ctx, MsgNFANote))

	return func() tea.Msg {
		return tea.Printf("\n" + strings.Join(bannerLines, "\n"))()
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

// printSkillsList 输出技能列表
func (ui *ChatUI) printSkillsList() tea.Cmd {
	var buf strings.Builder

	// 标题和总数
	buf.WriteString("\033[2m" + strings.Repeat("─", ui.width) + "\033[0m\n")
	buf.WriteString("\033[36mSkills\033[0m\n")
	buf.WriteString("\033[30m" + i18nutil.LocalizeContext(ui.ctx, &i18n.LocalizeConfig{
		DefaultMessage: MsgSkillsCount,
		PluralCount:    len(ui.skills),
		TemplateData:   map[string]any{"Count": len(ui.skills)},
	}) + "\033[0m\n\n")

	// 按 Source 分组
	var builtins, locals []skills.SkillMeta
	for _, s := range ui.skills {
		if s.Source == skills.SkillSourceBuiltin {
			builtins = append(builtins, s)
		} else {
			locals = append(locals, s)
		}
	}

	// Builtin skills
	if len(builtins) > 0 {
		buf.WriteString("\033[30mBuiltin skills\033[0m\n")
		for _, s := range builtins {
			buf.WriteString(fmt.Sprintf("\033[1m%s\033[0m - %s\n", s.Name, s.Description))
		}
		if len(locals) > 0 {
			buf.WriteString("\n")
		}
	}

	// Local skills
	if len(locals) > 0 {
		buf.WriteString("\033[30mLocal skills\033[0m\n")
		for _, s := range locals {
			buf.WriteString(fmt.Sprintf("\033[1m%s\0330m - %s\n", s.Name, s.Description))
		}
	}
	buf.WriteString("\033[2m" + strings.Repeat("─", ui.width) + "\033[0m\n")

	return tea.Println(buf.String())
}
