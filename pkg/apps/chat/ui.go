package chat

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coder/acp-go-sdk"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/yhlooo/nfa/pkg/agents"
	"github.com/yhlooo/nfa/pkg/configs"
	"github.com/yhlooo/nfa/pkg/history"
	i18nutil "github.com/yhlooo/nfa/pkg/i18n"
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

var _ tea.Model = (*Chat)(nil)
var _ acp.Client = (*Chat)(nil)

// Init 开始运行 UI 的第一个操作
func (chat *Chat) Init() tea.Cmd {
	var sessionCmd tea.Cmd
	if chat.resumeSessionID != "" {
		sessionCmd = chat.loadSession
	} else {
		sessionCmd = chat.newSession
	}

	cmds := []tea.Cmd{
		chat.printHello(),
		sessionCmd,
		textarea.Blink,
	}
	if chat.initialPrompt != "" {
		// 在 session 创建后发送初始 prompt
		cmds = append(cmds, chat.newPrompt(chat.initialPrompt))
	}
	return tea.Sequence(cmds...)
}

// Update 处理更新事件
func (chat *Chat) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// 根据 viewState 路由消息
	switch chat.viewState {
	case viewStateInput:
		return chat.updateInInputState(msg)
	case viewStateModelSelect:
		return chat.updateInModelSelectState(msg)
	}
	return chat, nil
}

// updateInInputState 处理输入状态
func (chat *Chat) updateInInputState(msg tea.Msg) (tea.Model, tea.Cmd) {
	logger := chat.logger

	var inputCmd tea.Cmd
	chat.input, inputCmd = chat.input.Update(msg)
	var vpCmd tea.Cmd
	chat.vp, vpCmd = chat.vp.Update(msg)

	cmds := []tea.Cmd{inputCmd, vpCmd}

	switch typedMsg := msg.(type) {
	case tea.WindowSizeMsg:
		chat.width = typedMsg.Width
		chat.modelUsageStyle = chat.modelUsageStyle.Width(typedMsg.Width)
		chat.logger.Info(fmt.Sprintf("resize message: width: %d, height: %d", typedMsg.Width, typedMsg.Height))

	case tea.KeyMsg:
		chat.logger.Info(fmt.Sprintf("key message: %q", typedMsg.String()))
		switch typedMsg.Type {
		case tea.KeyEnter:
			if !chat.vp.AgentProcessing() && !chat.input.MultiLineMode() {
				content := strings.TrimRight(chat.input.Value(), "\n")
				chat.input.Reset()
				if content != "" {
					// 保存到历史记录
					chat.history.Add(content)
					if err := history.SaveHistory(chat.historyPath, chat.history); err != nil {
						chat.logger.Error(err, "failed to save history")
					}

					switch content {
					case "/exit":
						return chat, tea.Quit
					case "/model", "/model :primary":
						return chat, chat.enterModelSelectMode(ModelTypePrimary)
					case "/model :light":
						return chat, chat.enterModelSelectMode(ModelTypeLight)
					case "/model :vision":
						return chat, chat.enterModelSelectMode(ModelTypeVision)
					case "/skills":
						cmds = append(cmds, chat.printSkillsList())
					default:
						// 检查是否是 /model 开头的直接设置命令
						if modelType, modelName, ok := chat.handleDirectModelSet(content); ok {
							cmds = append(cmds, tea.Printf(
								"\033[34m✓ %s %s\033[0m",
								i18nutil.LocalizeContext(chat.ctx, &i18n.LocalizeConfig{
									DefaultMessage: MsgSetModel,
									TemplateData:   map[string]any{"Type": modelType},
								}),
								modelName,
							))
						} else {
							cmds = append(cmds, chat.newPrompt(content))
						}
					}
				}
			}

		case tea.KeyEsc:
			if chat.vp.AgentProcessing() {
				cmds = append(cmds, chat.cancelPrompt)
			}

		case tea.KeyCtrlC:
			if !chat.vp.AgentProcessing() {
				return chat, tea.Quit
			}
			cmds = append(cmds, chat.cancelPrompt)
		}

	case acp.PromptResponse:
		chat.modelUsage = agents.GetMetaCurrentModelUsageValue(typedMsg.Meta)
		cmds = append(cmds, chat.vp.Flush())
		if chat.autoExitAfterResponse {
			cmds = append(cmds, tea.Quit)
		}
	case acp.SessionNotification:
		chat.modelUsage = agents.GetMetaCurrentModelUsageValue(typedMsg.Meta)
		cmds = append(cmds, chat.vp.Flush())
	case acp.PromptRequest:
		cmds = append(cmds, chat.vp.Flush())

	case QuitError:
		logger.Error(typedMsg.Error, "error")
		cmds = append(cmds, tea.Quit)

	case error:
		logger.Error(typedMsg, "error")
	}

	cmds = append(cmds, chat.updateComponents()...)

	return chat, tea.Batch(cmds...)
}

// updateInModelSelectState 处理模型选择状态
func (chat *Chat) updateInModelSelectState(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var vpCmd tea.Cmd
	chat.vp, vpCmd = chat.vp.Update(msg)
	cmds = append(cmds, vpCmd)

	var selectorCmd tea.Cmd
	chat.modelSelector, selectorCmd = chat.modelSelector.Update(msg)
	cmds = append(cmds, selectorCmd)

	switch typedMsg := msg.(type) {
	case tea.KeyMsg:
		switch typedMsg.Type {
		case tea.KeyEnter:
			var modelType ModelType
			modelName := ""
			modelType, modelName, chat.curModels = chat.modelSelector.GetSelectedModels()
			if err := configs.SaveDefaultModels(chat.cfgPath, chat.curModels); err != nil {
				cmds = append(cmds, func() tea.Msg {
					return fmt.Errorf("failed to save model: %w", err)
				})
			} else {
				cmds = append(cmds, tea.Printf(
					"\033[34m✓ %s %s\033[0m",
					i18nutil.LocalizeContext(chat.ctx, &i18n.LocalizeConfig{
						DefaultMessage: MsgSetModel,
						TemplateData:   map[string]any{"Type": modelType},
					}),
					modelName,
				))
			}
			cmds = append(cmds, chat.exitModelSelectMode())
		case tea.KeyEsc:
			// 取消选择
			cmds = append(cmds, chat.exitModelSelectMode())
		}
	}

	cmds = append(cmds, chat.updateComponents()...)

	return chat, tea.Batch(cmds...)
}

// updateComponents 根据状态更新组件
func (chat *Chat) updateComponents() []tea.Cmd {
	var cmds []tea.Cmd

	// 设置输入状态
	if chat.vp.AgentProcessing() {
		chat.input.Blur()
	} else {
		if !chat.input.Focused() {
			cmds = append(cmds, chat.input.Focus())
		}
	}

	return cmds
}

// enterModelSelectMode 进入模型选择模式
func (chat *Chat) enterModelSelectMode(t ModelType) tea.Cmd {
	chat.modelSelector.SetModelType(t)
	chat.viewState = viewStateModelSelect
	return nil
}

// exitModelSelectMode 退出模型选择模式
func (chat *Chat) exitModelSelectMode() tea.Cmd {
	chat.viewState = viewStateInput
	return nil
}

// View 渲染显示内容
func (chat *Chat) View() string {
	vpView := chat.vp.View()
	if vpView != "" {
		vpView += "\n"
	}

	var bottomView string
	switch chat.viewState {
	case viewStateInput:
		if chat.input.Focused() {
			bottomView = chat.input.View()
		}
	case viewStateModelSelect:
		bottomView = chat.modelSelector.View()
	}

	modelUsageView := ""
	if chat.modelUsage.InputTokens != 0 {
		modelUsageView += fmt.Sprintf("↑ %s", intWithSeparator(chat.modelUsage.InputTokens))
	}
	if out := chat.modelUsage.ThoughtsTokens + chat.modelUsage.ThoughtsTokens; out != 0 {
		modelUsageView += fmt.Sprintf(" | ↓ %s", intWithSeparator(out))
	}
	if modelUsageView != "" {
		modelUsageView = i18nutil.TContext(chat.ctx, MsgTokenUsage) + " " + strings.TrimPrefix(modelUsageView, " | ")
	}

	return fmt.Sprintf(
		`%s
%s%s

`,
		vpView,
		bottomView,
		chat.modelUsageStyle.Render(modelUsageView),
	)
}

// printHello 输出欢迎信息
func (chat *Chat) printHello() tea.Cmd {
	bannerLines := strings.Split(otter.MustOtter(true, false, 1), "\n")
	if len(bannerLines) < 5 {
		return nil
	}

	i := len(bannerLines) - 6
	bannerLines[i] += fmt.Sprintf("\r\033[36C\033[1mNFA\033[2m v%s\033[0m", version.Version)
	i++
	bannerLines[i] += fmt.Sprintf("\r\033[36C\033[2m%s\033[0m", chat.curModels.GetPrimary())
	i++
	if chat.curModels.GetLight() != chat.curModels.GetPrimary() {
		bannerLines[i] += fmt.Sprintf("\r\033[36C\033[2m%s (light)\033[0m", chat.curModels.GetLight())
		i++
	}
	if visionModel := chat.curModels.GetVision(); visionModel != "" && visionModel != chat.curModels.GetPrimary() {
		bannerLines[i] += fmt.Sprintf("\r\033[36C\033[2m%s (vision)\033[0m", chat.curModels.GetVision())
		i++
	}
	bannerLines[i] += fmt.Sprintf("\r\033[36C\033[1;33m%s\033[0m", i18nutil.TContext(chat.ctx, MsgNFANote))

	return tea.Println("\n" + strings.Join(bannerLines, "\n") + "\n")
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
func (chat *Chat) printSkillsList() tea.Cmd {
	var buf strings.Builder

	// 标题和总数
	buf.WriteString("\033[2m" + strings.Repeat("─", chat.width) + "\033[0m\n")
	buf.WriteString("\033[36m" + i18nutil.TContext(chat.ctx, MsgSkills) + "\033[0m\n")
	buf.WriteString("\033[30m" + i18nutil.LocalizeContext(chat.ctx, &i18n.LocalizeConfig{
		DefaultMessage: MsgSkillsCount,
		PluralCount:    len(chat.skills),
		TemplateData:   map[string]any{"Count": len(chat.skills)},
	}) + "\033[0m\n\n")

	// 按 Source 分组
	var builtins, locals []skills.SkillMeta
	for _, s := range chat.skills {
		if s.Source == skills.SkillSourceBuiltin {
			builtins = append(builtins, s)
		} else {
			locals = append(locals, s)
		}
	}

	// Builtin skills
	if len(builtins) > 0 {
		buf.WriteString("\033[30m" + i18nutil.TContext(chat.ctx, MsgBuiltinSkills) + "\033[0m\n")
		for _, s := range builtins {
			buf.WriteString(fmt.Sprintf("\033[1m%s\033[0m - %s\n", s.Name, s.Description))
		}
		if len(locals) > 0 {
			buf.WriteString("\n")
		}
	}

	// Local skills
	if len(locals) > 0 {
		buf.WriteString("\033[30m" + i18nutil.TContext(chat.ctx, MsgLocalSkills) + "\033[0m\n")
		for _, s := range locals {
			buf.WriteString(fmt.Sprintf("\033[1m%s\0330m - %s\n", s.Name, s.Description))
		}
	}
	buf.WriteString("\033[2m" + strings.Repeat("─", chat.width) + "\033[0m\n")

	return tea.Println(buf.String())
}

// handleDirectModelSet 处理直接设置模型的命令
func (chat *Chat) handleDirectModelSet(content string) (modelType ModelType, modelName string, ok bool) {
	parts := strings.Fields(content)
	if len(parts) == 0 || parts[0] != "/model" {
		return "", "", false
	}

	modelType = ModelTypePrimary

	switch len(parts) {
	case 2:
		// /model <provider>/<name>
		modelName = parts[1]
	case 3:
		// /model :target <provider>/<name>
		if !strings.HasPrefix(parts[1], ":") {
			// 无效格式，不处理，让用户重新输入
			return "", "", false
		}
		modelType = ModelType(strings.TrimPrefix(parts[1], ":"))
		modelName = parts[2]
	default:
		// 无效格式，不处理
		return "", "", false
	}

	if modelName == "" {
		return "", "", false
	}

	// 应用模型并保存
	switch modelType {
	case ModelTypePrimary:
		chat.curModels.Primary = modelName
	case ModelTypeLight:
		chat.curModels.Light = modelName
	case ModelTypeVision:
		chat.curModels.Vision = modelName
	default:
		return "", "", false
	}

	chat.modelSelector.SetCurrentModels(chat.curModels)

	// 只保存 defaultModels 字段到文件
	if err := configs.SaveDefaultModels(chat.cfgPath, chat.curModels); err != nil {
		// 保存失败，忽略错误
		chat.logger.Error(err, "save default models error")
		return modelType, modelName, true
	}

	return modelType, modelName, true
}
