package chat

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coder/acp-go-sdk"

	"github.com/yhlooo/nfa/pkg/agents"
	"github.com/yhlooo/nfa/pkg/version"
)

// ACPAgent ACP Agent 接口
type ACPAgent interface {
	acp.Agent
	acp.AgentLoader
}

// initAgent 初始化 Agent
func (chat *Chat) initAgent(ctx context.Context) error {
	_, err := chat.agent.Initialize(ctx, acp.InitializeRequest{
		ClientCapabilities: acp.ClientCapabilities{},
		ClientInfo: &acp.Implementation{
			Name:    "NFA",
			Title:   acp.Ptr("NFA (Not Financial Advice)"),
			Version: version.Version,
		},
	})
	if err != nil {
		return fmt.Errorf("initialize agent error: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("new session error: get current working directory error: %w", err)
	}
	chat.cwd = cwd

	return nil
}

// newSession 创建会话
func (chat *Chat) newSession() tea.Msg {
	resp, err := chat.agent.NewSession(chat.ctx, acp.NewSessionRequest{
		Cwd:        chat.cwd,
		McpServers: []acp.McpServer{},
	})
	if err != nil {
		return QuitError{Error: fmt.Errorf("new session error: %w", err)}
	}
	chat.sessionID = resp.SessionId
	if resp.Models != nil {
		chat.curPrimaryModel = string(resp.Models.CurrentModelId)
	}

	return nil
}

// loadSession 加载会话
func (chat *Chat) loadSession() tea.Msg {
	resp, err := chat.agent.LoadSession(chat.ctx, acp.LoadSessionRequest{
		SessionId:  acp.SessionId(chat.resumeSessionID),
		Cwd:        chat.cwd,
		McpServers: []acp.McpServer{},
	})
	if err != nil {
		return QuitError{Error: fmt.Errorf("load session error: %w", err)}
	}
	chat.sessionID = acp.SessionId(chat.resumeSessionID)
	if resp.Models != nil {
		chat.curPrimaryModel = string(resp.Models.CurrentModelId)
	}
	return nil
}

// newPrompt 开始一轮对话
func (chat *Chat) newPrompt(prompt string) tea.Cmd {
	return func() tea.Msg {
		req := acp.PromptRequest{
			SessionId: chat.sessionID,
			Prompt: []acp.ContentBlock{
				acp.TextBlock(prompt),
			},
		}
		chat.p.Send(req)
		resp, err := chat.agent.Prompt(chat.ctx, req)
		if err != nil {
			chat.p.Send(fmt.Errorf("new prompt error: %w", err))
		}
		return resp
	}
}

// cancelPrompt 取消当轮对话
func (chat *Chat) cancelPrompt() tea.Msg {
	err := chat.agent.Cancel(chat.ctx, acp.CancelNotification{
		SessionId: chat.sessionID,
	})
	if err != nil {
		return fmt.Errorf("cancel prompt error: %w", err)
	}
	return nil
}

// RequestPermission 请求授权
func (chat *Chat) RequestPermission(
	_ context.Context,
	params acp.RequestPermissionRequest,
) (acp.RequestPermissionResponse, error) {
	if len(params.Options) == 0 {
		return acp.RequestPermissionResponse{Outcome: acp.NewRequestPermissionOutcomeCancelled()}, nil
	}
	// TODO: 暂不支持，总是选第一个
	return acp.RequestPermissionResponse{
		Outcome: acp.NewRequestPermissionOutcomeSelected(params.Options[0].OptionId),
	}, nil
}

// SessionUpdate 更新会话
func (chat *Chat) SessionUpdate(ctx context.Context, params acp.SessionNotification) error {
	chat.p.Send(params)
	if channelID := agents.GetMetaIntValue(params.Meta, channelIDMetaKey); channelID > 0 &&
		channelID <= len(chat.channels) {
		if err := chat.channels[channelID-1].Send(ctx, params.Meta, &params, false); err != nil {
			chat.logger.Error(err, "send notification to channel error")
		}
	}
	return nil
}
