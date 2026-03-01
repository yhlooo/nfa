package chat

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coder/acp-go-sdk"

	"github.com/yhlooo/nfa/pkg/agents"
	"github.com/yhlooo/nfa/pkg/configs"
	"github.com/yhlooo/nfa/pkg/version"
)

// initAgent 初始化 Agent
func (ui *ChatUI) initAgent(ctx context.Context) error {
	resp, err := ui.conn.Initialize(ctx, acp.InitializeRequest{
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
	ui.curModels = agents.GetMetaCurrentModelsValue(resp.Meta)
	ui.modelSelector.SetAvailableModels(agents.GetMetaAvailableModelsValue(resp.Meta))
	ui.modelSelector.SetCurrentModels(ui.curModels)

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("new session error: get current working directory error: %w", err)
	}
	ui.cwd = cwd

	return nil
}

// newSession 创建会话
func (ui *ChatUI) newSession() tea.Msg {
	resp, err := ui.conn.NewSession(ui.ctx, acp.NewSessionRequest{
		Cwd:        ui.cwd,
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
			Meta: map[string]any{
				agents.MetaKeyCurrentModels: ui.curModels,
			},
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

// enterModelSelectMode 进入模型选择模式
func (ui *ChatUI) enterModelSelectMode(t ModelType) tea.Cmd {
	ui.modelSelector.SetModelType(t)
	ui.viewState = viewStateModelSelect
	return nil
}

// exitModelSelectMode 退出模型选择模式
func (ui *ChatUI) exitModelSelectMode() tea.Cmd {
	ui.viewState = viewStateInput
	return nil
}

// handleDirectModelSet 处理直接设置模型的命令
func (ui *ChatUI) handleDirectModelSet(content string) (modelType ModelType, modelName string, ok bool) {
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
		ui.curModels.Primary = modelName
	case ModelTypeLight:
		ui.curModels.Light = modelName
	case ModelTypeVision:
		ui.curModels.Vision = modelName
	default:
		return "", "", false
	}

	ui.modelSelector.SetCurrentModels(ui.curModels)

	// 只保存 defaultModels 字段到文件
	if err := configs.SaveDefaultModels(ui.cfgPath, ui.curModels); err != nil {
		// 保存失败，忽略错误
		ui.logger.Error(err, "save default models error")
		return modelType, modelName, true
	}

	return modelType, modelName, true
}
