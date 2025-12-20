package tty

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coder/acp-go-sdk"

	"github.com/yhlooo/nfa/pkg/version"
)

// initAgent 初始化 Agent
func (ui *ChatUI) initAgent() tea.Msg {
	_, err := ui.conn.Initialize(ui.ctx, acp.InitializeRequest{
		ClientCapabilities: acp.ClientCapabilities{},
		ClientInfo: &acp.Implementation{
			Name:    "NFA",
			Title:   acp.Ptr("NFA (Not Financial Advice)"),
			Version: version.Version,
		},
	})
	if err != nil {
		return QuitError{Error: fmt.Errorf("initialize agent error: %w", err)}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return QuitError{Error: fmt.Errorf("new session error: get current working directory error: %w", err)}
	}

	resp, err := ui.conn.NewSession(ui.ctx, acp.NewSessionRequest{
		Cwd:        cwd,
		McpServers: []acp.McpServer{},
	})
	if err != nil {
		return QuitError{Error: fmt.Errorf("new session error: %w", err)}
	}
	ui.sessionID = resp.SessionId

	return nil
}

// newPrompt 开始一轮对话
func (ui *ChatUI) newPrompt(prompt string) tea.Cmd {
	return func() tea.Msg {
		req := acp.PromptRequest{
			SessionId: ui.sessionID,
			Prompt: []acp.ContentBlock{
				acp.TextBlock(prompt),
			},
		}
		ui.p.Send(req)
		resp, err := ui.conn.Prompt(ui.ctx, req)
		if err != nil {
			ui.p.Send(fmt.Errorf("new prompt error: %w", err))
		}
		return resp
	}
}

// cancelPrompt 取消当轮对话
func (ui *ChatUI) cancelPrompt() tea.Msg {
	err := ui.conn.Cancel(ui.ctx, acp.CancelNotification{
		SessionId: ui.sessionID,
	})
	if err != nil {
		return fmt.Errorf("cancel prompt error: %w", err)
	}
	return nil
}

// RequestPermission 请求授权
func (ui *ChatUI) RequestPermission(
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
func (ui *ChatUI) SessionUpdate(_ context.Context, params acp.SessionNotification) error {
	ui.p.Send(params)
	return nil
}
