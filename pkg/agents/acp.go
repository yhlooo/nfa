package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/coder/acp-go-sdk"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/go-logr/logr"
	"github.com/google/uuid"

	"github.com/yhlooo/nfa/pkg/acputil"
	"github.com/yhlooo/nfa/pkg/agents/flows"
	"github.com/yhlooo/nfa/pkg/version"
)

var _ acp.Agent = (*NFAAgent)(nil)

// Connect 连接
func (a *NFAAgent) Connect(in io.Reader, out io.Writer) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.conn != nil {
		select {
		case <-a.conn.Done():
		default:
			return fmt.Errorf("already connected")
		}
	}

	a.conn = acp.NewAgentSideConnection(a, out, in)

	return nil
}

// Initialize 初始化连接
func (a *NFAAgent) Initialize(ctx context.Context, _ acp.InitializeRequest) (acp.InitializeResponse, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.InitGenkit(ctx)

	return acp.InitializeResponse{
		Meta: map[string]any{
			MetaKeyAvailableModels: a.availableModels,
			MetaKeyDefaultModel:    a.defaultModel,
		},
		AgentCapabilities: acp.AgentCapabilities{},
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
func (a *NFAAgent) NewSession(_ context.Context, _ acp.NewSessionRequest) (acp.NewSessionResponse, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	sessionID := acp.SessionId(uuid.New().String())
	a.sessions[sessionID] = &Session{
		id: sessionID,
	}

	return acp.NewSessionResponse{
		SessionId: sessionID,
	}, nil
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
	defer func() {
		session.lock.Lock()
		session.cancelPrompt = nil
		session.history = messages
		session.lock.Unlock()
	}()
	session.lock.Unlock()

	modelName := GetMetaStringValue(params.Meta, MetaKeyModelName)
	if modelName == "" {
		modelName = a.defaultModel
	}
	if modelName == "" {
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
	ctx = flows.ContextWithModelName(ctx, modelName)
	ctx = logr.NewContext(ctx, a.logger)

	a.logger.Info("prompt turn start")

	handleStream := a.handleStreamChunk(params.SessionId)
	resp := acp.PromptResponse{StopReason: acp.StopReasonEndTurn}
	var finalErr error

	// 斜杠命令
	switch strings.TrimSpace(prompt) {
	case "/clear":
		_ = a.conn.SessionUpdate(ctx, acp.SessionNotification{
			SessionId: params.SessionId,
			Update:    acp.UpdateAgentMessageText("The context has been cleared."),
		})
		messages = nil
	case "/summarize", "/summary":
		a.summarizeFlow.Stream(ctx, flows.SummarizeInput{
			History: messages,
		})(func(chunk *core.StreamingFlowValue[flows.SummarizeOutput, *ai.ModelResponseChunk], err error) bool {
			if err != nil {
				if errors.Is(err, context.Canceled) {
					resp.StopReason = acp.StopReasonCancelled
				} else {
					resp.StopReason = acp.StopReasonRefusal
					finalErr = err
				}
				return false
			}

			if chunk.Stream != nil {
				if err := handleStream(ctx, chunk.Stream); err != nil {
					resp.StopReason = acp.StopReasonRefusal
					finalErr = err
					return false
				}
			}
			if chunk.Done {
				content := fmt.Sprintf(`# %s

%s

## 过程概述

%s`, chunk.Output.Title, chunk.Output.Description, chunk.Output.ProcessOverview)
				content = strings.TrimRight(content, "\n")
				if chunk.Output.MethodologySummary != "" {
					content += fmt.Sprintf(`

# 方法论

%s`, chunk.Output.MethodologySummary)
				}

				if err := a.conn.SessionUpdate(ctx, acp.SessionNotification{
					SessionId: params.SessionId,
					Update:    acp.UpdateAgentMessageText(content),
				}); err != nil {
					resp.StopReason = acp.StopReasonRefusal
					finalErr = fmt.Errorf("session update error: %w", err)
					return false
				}
			}

			return !chunk.Done
		})
	default:
		a.mainFlow.Stream(ctx, flows.ChatInput{
			Prompt:  prompt,
			History: messages,
		})(func(chunk *core.StreamingFlowValue[flows.ChatOutput, *ai.ModelResponseChunk], err error) bool {
			if err != nil {
				if errors.Is(err, context.Canceled) {
					resp.StopReason = acp.StopReasonCancelled
				} else {
					resp.StopReason = acp.StopReasonRefusal
					finalErr = err
				}
				return false
			}

			if chunk.Stream != nil {
				if err := handleStream(ctx, chunk.Stream); err != nil {
					resp.StopReason = acp.StopReasonRefusal
					finalErr = err
					return false
				}
			}
			if chunk.Done {
				messages = append(messages, chunk.Output.Messages...)
			}

			return !chunk.Done
		})
	}

	return resp, finalErr
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
func (a *NFAAgent) handleStreamChunk(sessionID acp.SessionId) ai.ModelStreamCallback {
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
				if err := a.flushBufferText(ctx, sessionID, acp.UpdateAgentMessageText, text); err != nil {
					return err
				}
				reasoning.WriteString(part.Text)
			case part.IsText() || part.IsData():
				if err := a.flushBufferText(ctx, sessionID, acp.UpdateAgentThoughtText, reasoning); err != nil {
					return err
				}
				text.WriteString(part.Text)
			case part.IsToolRequest() && part.ToolRequest != nil:
				if err := a.flushBufferText(ctx, sessionID, acp.UpdateAgentThoughtText, reasoning); err != nil {
					return err
				}
				if err := a.flushBufferText(ctx, sessionID, acp.UpdateAgentMessageText, text); err != nil {
					return err
				}

				inputRaw, _ := json.Marshal(part.ToolRequest.Input)
				if err := a.conn.SessionUpdate(ctx, acp.SessionNotification{
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
				if err := a.flushBufferText(ctx, sessionID, acp.UpdateAgentThoughtText, reasoning); err != nil {
					return err
				}
				if err := a.flushBufferText(ctx, sessionID, acp.UpdateAgentMessageText, text); err != nil {
					return err
				}
				outputRaw, _ := json.Marshal(part.ToolResponse.Output)
				if err := a.conn.SessionUpdate(ctx, acp.SessionNotification{
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

		if err := a.flushBufferText(ctx, sessionID, acp.UpdateAgentThoughtText, reasoning); err != nil {
			return err
		}
		if err := a.flushBufferText(ctx, sessionID, acp.UpdateAgentMessageText, text); err != nil {
			return err
		}

		return nil
	}
}

// flushBufferText 刷文本消息缓存
func (a *NFAAgent) flushBufferText(
	ctx context.Context,
	sessionID acp.SessionId,
	buildUpdateFn func(string) acp.SessionUpdate,
	buff strings.Builder,
) error {
	if buff.Len() == 0 {
		return nil
	}

	err := a.conn.SessionUpdate(ctx, acp.SessionNotification{
		SessionId: sessionID,
		Update:    buildUpdateFn(buff.String()),
	})
	buff.Reset()
	if err != nil {
		return fmt.Errorf("session update error: %w", err)
	}

	return nil
}
