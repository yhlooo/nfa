package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/coder/acp-go-sdk"
	"github.com/firebase/genkit/go/ai"
	"github.com/go-logr/logr"
	"github.com/google/uuid"

	"github.com/yhlooo/nfa/pkg/acputil"
	"github.com/yhlooo/nfa/pkg/agents/flows"
	"github.com/yhlooo/nfa/pkg/ctxutil"
	"github.com/yhlooo/nfa/pkg/i18n"
	i18nutil "github.com/yhlooo/nfa/pkg/i18n"
	"github.com/yhlooo/nfa/pkg/skills"
	"github.com/yhlooo/nfa/pkg/tokentracker"
	"github.com/yhlooo/nfa/pkg/version"
)

var _ acp.Agent = (*NFAAgent)(nil)
var _ acp.AgentLoader = (*NFAAgent)(nil)

// ConnectClientIO 连接客户端输入输出流
func (a *NFAAgent) ConnectClientIO(in io.Reader, out io.Writer) {
	a.SetClient(acp.NewAgentSideConnection(a, out, in))
}

// SetClient 设置客户端
func (a *NFAAgent) SetClient(client acp.Client) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.client = client
}

// Initialize 初始化连接
func (a *NFAAgent) Initialize(ctx context.Context, _ acp.InitializeRequest) (acp.InitializeResponse, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	ctx = i18n.ContextWithLocalizer(ctx, a.localizer)
	ctx = logr.NewContext(ctx, a.logger)

	// 初始化技能加载器并加载技能
	a.skillLoader = skills.NewSkillLoader(filepath.Join(a.opts.DataRoot, "skills"))
	if err := a.skillLoader.LoadMeta(ctx); err != nil {
		a.logger.Error(err, "load skills error")
	}

	// 初始化 genkit
	a.InitGenkit(ctx)

	return acp.InitializeResponse{
		AgentCapabilities: acp.AgentCapabilities{
			LoadSession: true,
		},
		AgentInfo: &acp.Implementation{
			Name:    "NFA",
			Title:   acp.Ptr("NFA (Not Financial Advice)"),
			Version: version.Version,
		},
	}, nil
}

// Authenticate 认证
func (a *NFAAgent) Authenticate(_ context.Context, _ acp.AuthenticateRequest) (acp.AuthenticateResponse, error) {
	return acp.AuthenticateResponse{}, nil
}

// NewSession 创建会话
func (a *NFAAgent) NewSession(ctx context.Context, _ acp.NewSessionRequest) (acp.NewSessionResponse, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	curModels := a.opts.DefaultModels
	availableModels := make([]acp.ModelInfo, len(a.availableModels))
	for i, m := range a.availableModels {
		availableModels[i] = acp.ModelInfo{
			ModelId: acp.ModelId(m.Provider + "/" + m.Name),
			Name:    m.Name,
		}
	}

	sessionID := acp.SessionId(uuid.New().String())
	a.sessions[sessionID] = &Session{
		id:            sessionID,
		currentModels: curModels,
		tokenTracker:  tokentracker.NewTracker(a.availableModels),
	}

	go a.sendAvailableCommands(ctx, sessionID)
	return acp.NewSessionResponse{
		SessionId: sessionID,
		Models: &acp.SessionModelState{
			AvailableModels: availableModels,
			CurrentModelId:  acp.ModelId(curModels.GetPrimary()),
		},
	}, nil
}

// LoadSession 加载已有会话
func (a *NFAAgent) LoadSession(ctx context.Context, params acp.LoadSessionRequest) (acp.LoadSessionResponse, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	// 从文件加载会话数据
	data, err := LoadSessionData(filepath.Join(a.opts.DataRoot, SessionsDirName), params.SessionId)
	if err != nil {
		return acp.LoadSessionResponse{}, fmt.Errorf("load session data error: %w", err)
	}

	curModels := a.opts.DefaultModels
	availableModels := make([]acp.ModelInfo, len(a.availableModels))
	for i, m := range a.availableModels {
		availableModels[i] = acp.ModelInfo{
			ModelId: acp.ModelId(m.Provider + "/" + m.Name),
			Name:    m.Name,
		}
	}

	// 创建会话
	a.sessions[params.SessionId] = &Session{
		id:            params.SessionId,
		history:       data.Messages,
		currentModels: curModels,
		tokenTracker:  tokentracker.NewTracker(a.availableModels),
	}

	// 回放历史消息
	handleFn := a.handleStreamChunk(params.SessionId, nil)
	for _, msg := range data.Messages {
		time.Sleep(5 * time.Millisecond) // 由于客户端并行处理，发送太快的话到达客户端是乱序的，所以这里略微 sleep 一下
		switch msg.Role {
		case ai.RoleUser:
			if err := a.client.SessionUpdate(ctx, acp.SessionNotification{
				SessionId: params.SessionId,
				Update:    acp.UpdateUserMessageText(msg.Text()),
			}); err != nil {
				return acp.LoadSessionResponse{}, err
			}
		default:
			if err := handleFn(ctx, &ai.ModelResponseChunk{
				Content: msg.Content,
				Role:    msg.Role,
			}); err != nil {
				return acp.LoadSessionResponse{}, err
			}
		}
	}

	go a.sendAvailableCommands(ctx, params.SessionId)
	return acp.LoadSessionResponse{
		Models: &acp.SessionModelState{
			AvailableModels: availableModels,
			CurrentModelId:  acp.ModelId(curModels.GetPrimary()),
		},
	}, nil
}

// sendAvailableCommands 发送可用命令列表
func (a *NFAAgent) sendAvailableCommands(ctx context.Context, sessionID acp.SessionId) {
	time.Sleep(10 * time.Millisecond)
	commands := []acp.AvailableCommand{
		{
			Name:        "clear",
			Description: i18nutil.TContext(ctx, MsgCmdDescClear),
		},
	}
	for _, skill := range a.skillLoader.ListMeta() {
		commands = append(commands, acp.AvailableCommand{
			Name:        skill.Name,
			Description: skill.Description,
		})
	}

	if err := a.client.SessionUpdate(ctx, acp.SessionNotification{
		SessionId: sessionID,
		Update: acp.SessionUpdate{
			AvailableCommandsUpdate: &acp.SessionAvailableCommandsUpdate{
				AvailableCommands: commands,
			},
		},
	}); err != nil {
		a.logger.Error(err, "send available commands error")
	}
}

// SetSessionMode 设置会话模式
func (a *NFAAgent) SetSessionMode(_ context.Context, _ acp.SetSessionModeRequest) (acp.SetSessionModeResponse, error) {
	return acp.SetSessionModeResponse{},
		fmt.Errorf("%w: method session/set_mode not supported", acputil.ErrNotSupported)
}

// Prompt 对话
func (a *NFAAgent) Prompt(ctx context.Context, params acp.PromptRequest) (acp.PromptResponse, error) {
	a.lock.RLock()
	session, ok := a.sessions[params.SessionId]
	a.lock.RUnlock()

	if !ok {
		return acp.PromptResponse{StopReason: acp.StopReasonRefusal}, fmt.Errorf(
			"%w: session %q not found",
			acputil.ErrSessionNotFound, params.SessionId,
		)
	}

	session.lock.Lock()
	if session.cancelPrompt != nil {
		session.lock.Unlock()
		return acp.PromptResponse{StopReason: acp.StopReasonRefusal}, fmt.Errorf(
			"%w: session %q already in prompting",
			acputil.ErrInPrompting, session.id,
		)
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	session.cancelPrompt = cancel
	messages := session.history
	lastContextWindow := session.lastContextWindow
	defer func() {
		session.lock.Lock()
		session.cancelPrompt = nil
		session.history = messages
		session.lastContextWindow = lastContextWindow
		session.lock.Unlock()
	}()
	session.lock.Unlock()

	if session.currentModels.Primary == "" {
		return acp.PromptResponse{StopReason: acp.StopReasonRefusal}, fmt.Errorf("no available model")
	}

	prompt := ""
	for _, content := range params.Prompt {
		switch {
		case content.Text != nil:
			prompt += content.Text.Text + "\n"
		}
	}
	prompt = strings.TrimRight(prompt, "\n")
	if prompt == "" {
		return acp.PromptResponse{StopReason: acp.StopReasonEndTurn}, nil
	}
	ctx = ctxutil.ContextWithModels(ctx, session.currentModels)
	ctx = tokentracker.ContextWithTokenTracker(ctx, session.tokenTracker)
	ctx = logr.NewContext(ctx, a.logger)

	a.logger.Info("prompt turn start")

	extraMeta, _ := params.Meta.(map[string]any)
	handleStreamFn := a.handleStreamChunk(params.SessionId, extraMeta)
	ctx = ctxutil.ContextWithHandleStreamFn(ctx, handleStreamFn)
	resp := acp.PromptResponse{Meta: map[string]any{}, StopReason: acp.StopReasonEndTurn}

	// 斜杠命令
	if strings.HasPrefix(prompt, "/") {
		switch strings.TrimSpace(prompt) {
		case "/clear":
			_ = a.client.SessionUpdate(ctx, acp.SessionNotification{
				SessionId: params.SessionId,
				Update:    acp.UpdateAgentMessageText("The context has been cleared."),
			})
			messages = nil
			lastContextWindow = 0
			return resp, nil
		default:
			divided := strings.SplitN(prompt, " ", 2)
			if len(divided) >= 1 {
				cmdName := strings.TrimPrefix(divided[0], "/")
				skill, err := a.skillLoader.Get(cmdName)
				if err != nil {
					resp.StopReason = acp.StopReasonRefusal
					return resp, err
				}
				toolCallID := uuid.New().String()
				toolReq := ai.NewToolRequestPart(&ai.ToolRequest{
					Name:  skills.LoadSkillToolName,
					Input: skills.LoadSkillInput{Name: skill.Meta.Name},
					Ref:   toolCallID,
				})
				toolResp := ai.NewToolResponsePart(&ai.ToolResponse{
					Name: skills.LoadSkillToolName,
					Output: skills.LoadSkillOutput{
						Name:        skill.Meta.Name,
						Description: skill.Meta.Description,
						Content:     skill.Content,
					},
					Ref: toolCallID,
				})

				_ = handleStreamFn(ctx, &ai.ModelResponseChunk{
					Content: []*ai.Part{toolReq},
					Role:    ai.RoleModel,
				})
				_ = handleStreamFn(ctx, &ai.ModelResponseChunk{
					Content: []*ai.Part{toolResp},
					Role:    ai.RoleTool,
				})
				messages = append(messages,
					ai.NewUserTextMessage(prompt),
					ai.NewModelMessage(toolReq),
					ai.NewMessage(ai.RoleTool, nil, toolResp),
				)
			}
		}
	}

	history := make([]*ai.Message, len(messages))
	copy(history, messages)
	messages = append(messages, ai.NewUserTextMessage(prompt))

	if lastContextWindow > a.opts.MaxContextWindow {
		resp.StopReason = acp.StopReasonMaxTokens
		return resp, nil
	}
	chatOut, err := a.chatFlow.Run(ctx, flows.ChatInput{
		Prompt:           prompt,
		History:          history,
		MaxContextWindow: a.opts.MaxContextWindow,
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			resp.StopReason = acp.StopReasonCancelled
			err = nil
		} else {
			resp.StopReason = acp.StopReasonRefusal
			messages = append(messages, ai.NewModelTextMessage("Error: "+err.Error()))
		}
		SetMetaCurrentModelUsage(resp.Meta, session.tokenTracker.Summary())
		return resp, err
	}

	messages = append(messages, chatOut.Messages...)
	lastContextWindow = chatOut.LastContextWindow
	if lastContextWindow > a.opts.MaxContextWindow {
		resp.StopReason = acp.StopReasonMaxTokens
	}

	// 保存会话
	if err := SaveSession(filepath.Join(a.opts.DataRoot, SessionsDirName), params.SessionId, messages); err != nil {
		a.logger.Error(err, "save session error")
	}

	SetMetaCurrentModelUsage(resp.Meta, session.tokenTracker.Summary())
	return resp, nil
}

// Cancel 取消
func (a *NFAAgent) Cancel(_ context.Context, params acp.CancelNotification) error {
	a.lock.RLock()
	session, ok := a.sessions[params.SessionId]
	a.lock.RUnlock()

	if !ok {
		return fmt.Errorf("%w: session %q not found", acputil.ErrSessionNotFound, params.SessionId)
	}

	session.lock.Lock()
	defer session.lock.Unlock()

	if session.cancelPrompt != nil {
		session.cancelPrompt()
		session.cancelPrompt = nil
	}

	return nil
}

// handleStreamChunk 处理模型流输出
func (a *NFAAgent) handleStreamChunk(sessionID acp.SessionId, extraMeta map[string]any) ai.ModelStreamCallback {
	return func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		raw, _ := json.Marshal(chunk)
		a.logger.V(1).Info(fmt.Sprintf("model chunk: %s", string(raw)))

		var reasoning strings.Builder
		var text strings.Builder
		for _, part := range chunk.Content {
			switch {
			case part.IsReasoning():
				if err := a.flushBufferText(ctx, sessionID, extraMeta, acp.UpdateAgentMessageText, text); err != nil {
					return err
				}
				reasoning.WriteString(part.Text)
			case part.IsText() || part.IsData():
				if err := a.flushBufferText(
					ctx, sessionID, extraMeta, acp.UpdateAgentThoughtText,
					reasoning,
				); err != nil {
					return err
				}
				text.WriteString(part.Text)
			case part.IsToolRequest() && part.ToolRequest != nil:
				if err := a.flushBufferText(
					ctx, sessionID, extraMeta, acp.UpdateAgentThoughtText,
					reasoning,
				); err != nil {
					return err
				}
				if err := a.flushBufferText(ctx, sessionID, extraMeta, acp.UpdateAgentMessageText, text); err != nil {
					return err
				}

				inputRaw, _ := json.Marshal(part.ToolRequest.Input)
				meta := make(map[string]any)
				for k, v := range extraMeta {
					meta[k] = v
				}
				if err := a.client.SessionUpdate(ctx, acp.SessionNotification{
					Meta:      meta,
					SessionId: sessionID,
					Update: acp.StartToolCall(
						acp.ToolCallId(part.ToolRequest.Ref), part.ToolRequest.Name,
						acp.WithStartContent([]acp.ToolCallContent{
							{Content: &acp.ToolCallContentContent{Content: acp.TextBlock(string(inputRaw))}},
						}),
						acp.WithStartStatus(acp.ToolCallStatusInProgress),
					),
				}); err != nil {
					return fmt.Errorf("session update error: %w", err)
				}

			case part.IsToolResponse() && part.ToolResponse != nil:
				if err := a.flushBufferText(
					ctx, sessionID, extraMeta, acp.UpdateAgentThoughtText,
					reasoning,
				); err != nil {
					return err
				}
				if err := a.flushBufferText(ctx, sessionID, extraMeta, acp.UpdateAgentMessageText, text); err != nil {
					return err
				}
				outputRaw, _ := json.Marshal(part.ToolResponse.Output)
				meta := make(map[string]any)
				for k, v := range extraMeta {
					meta[k] = v
				}
				if err := a.client.SessionUpdate(ctx, acp.SessionNotification{
					Meta:      meta,
					SessionId: sessionID,
					Update: acp.UpdateToolCall(
						acp.ToolCallId(part.ToolResponse.Ref),
						acp.WithUpdateStatus(acp.ToolCallStatusCompleted),
						acp.WithUpdateContent([]acp.ToolCallContent{
							{Content: &acp.ToolCallContentContent{Content: acp.TextBlock(string(outputRaw))}},
						}),
					),
				}); err != nil {
					return fmt.Errorf("session update error: %w", err)
				}
			}
		}

		if err := a.flushBufferText(ctx, sessionID, extraMeta, acp.UpdateAgentThoughtText, reasoning); err != nil {
			return err
		}
		if err := a.flushBufferText(ctx, sessionID, extraMeta, acp.UpdateAgentMessageText, text); err != nil {
			return err
		}

		return nil
	}
}

// flushBufferText 刷文本消息缓存
func (a *NFAAgent) flushBufferText(
	ctx context.Context,
	sessionID acp.SessionId,
	extraMeta map[string]any,
	buildUpdateFn func(string) acp.SessionUpdate,
	buff strings.Builder,
) error {
	if buff.Len() == 0 {
		return nil
	}

	meta := make(map[string]any)
	for k, v := range extraMeta {
		meta[k] = v
	}
	err := a.client.SessionUpdate(ctx, acp.SessionNotification{
		Meta:      meta,
		SessionId: sessionID,
		Update:    buildUpdateFn(buff.String()),
	})
	buff.Reset()
	if err != nil {
		return fmt.Errorf("session update error: %w", err)
	}

	return nil
}
