package chat

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coder/acp-go-sdk"
	"github.com/firebase/genkit/go/ai"
	"github.com/go-logr/logr"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/yhlooo/nfa/pkg/acputil"
	"github.com/yhlooo/nfa/pkg/channels"
	"github.com/yhlooo/nfa/pkg/configs"
	"github.com/yhlooo/nfa/pkg/history"
	i18nutil "github.com/yhlooo/nfa/pkg/i18n"
	"github.com/yhlooo/nfa/pkg/skills"
)

// Options UI 运行选项
type Options struct {
	AgentIn               io.Writer
	AgentOut              io.Reader
	Agent                 ACPAgent
	InitialPrompt         string
	AutoExitAfterResponse bool
	ResumeSessionID       string
	Channels              []channels.Channel
}

// NewChat 创建对话应用
func NewChat(opts Options) *Chat {
	ui := &Chat{
		channels:              opts.Channels,
		modelUsageStyle:       lipgloss.NewStyle().Faint(true).Align(lipgloss.Right).PaddingRight(2),
		initialPrompt:         opts.InitialPrompt,
		autoExitAfterResponse: opts.AutoExitAfterResponse,
		resumeSessionID:       opts.ResumeSessionID,
		width:                 80,
	}

	if opts.Agent != nil {
		ui.agent = opts.Agent
	} else {
		ui.agent = acp.NewClientSideConnection(ui, opts.AgentIn, opts.AgentOut)
	}

	return ui
}

// Chat 对话应用
type Chat struct {
	ctx    context.Context
	logger logr.Logger
	p      *tea.Program

	vp    MessageViewport
	input *InputBox
	width int

	acputil.NopFS
	acputil.NopTerminal
	modelUsageStyle lipgloss.Style

	agent    ACPAgent
	channels []channels.Channel

	cwd                   string
	sessionID             acp.SessionId
	curPrimaryModel       string
	modelUsage            ai.GenerationUsage
	initialPrompt         string
	autoExitAfterResponse bool
	resumeSessionID       string
	skills                []skills.SkillMeta

	// Model selection
	cfgPath string

	// History
	history     *history.History
	historyPath string
}

// Run 运行
func (chat *Chat) Run(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx)

	chat.ctx = ctx
	chat.logger = logger

	var err error
	chat.vp, err = NewMessageViewport(ctx)
	if err != nil {
		return err
	}

	chat.cfgPath = configs.ConfigPathFromContext(ctx)

	// 确定历史文件路径（与配置文件同目录）
	chat.historyPath = filepath.Join(filepath.Dir(chat.cfgPath), "history.json")

	// 加载历史记录
	chat.history, err = history.LoadHistory(chat.historyPath)
	if err != nil {
		logger.Error(err, "failed to load history, starting with empty history")
		chat.history = history.NewHistory(100)
	}

	// 初始化 agent
	if err := chat.initAgent(ctx); err != nil {
		return fmt.Errorf("initialize agent error: %w", err)
	}

	chat.input = NewInputBox(ctx, []SelectorOption{
		{Name: "clear", Description: i18nutil.TContext(ctx, MsgCmdDescClear)},
		{Name: "model", Description: i18nutil.TContext(ctx, MsgCmdDescModel)},
		{Name: "skills", Description: i18nutil.TContext(ctx, MsgCmdDescSkills)},
		{Name: "exit", Description: i18nutil.TContext(ctx, MsgCmdDescExit)},
	}, chat.history, chat.historyPath)

	p := tea.NewProgram(chat, tea.WithContext(ctx))
	chat.p = p

	// 监听信道
	for i, ch := range chat.channels {
		go chat.handleChannel(ctx, i+1, ch)
	}

	_, err = p.Run()

	// 打印会话恢复提示
	if chat.sessionID != "" {
		fmt.Printf("\n%s\n%s\n",
			i18nutil.TContext(ctx, MsgResumeSession),
			i18nutil.LocalizeContext(ctx, &i18n.LocalizeConfig{
				DefaultMessage: MsgResumeCommand,
				TemplateData:   map[string]any{"SessionID": chat.sessionID},
			}),
		)
	}

	return err
}
