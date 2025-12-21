package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/coder/acp-go-sdk"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-logr/logr"
	"github.com/google/uuid"

	"github.com/yhlooo/nfa/pkg/acputil"
	"github.com/yhlooo/nfa/pkg/agents/models"
	"github.com/yhlooo/nfa/pkg/version"
)

const loggerName = "agent"

// Options Agent 运行选项
type Options struct {
	Logger         logr.Logger
	ModelProviders []models.ModelProvider
	DefaultModel   string
}

// Complete 使用默认值补全选项
func (opts *Options) Complete() {
	if len(opts.ModelProviders) == 0 {
		opts.ModelProviders = append(opts.ModelProviders, models.ModelProvider{Ollama: &models.OllamaOptions{}})
	}
}

// NewNFA 创建 NFA Agent
func NewNFA(opts Options) *NFAAgent {
	opts.Complete()
	return &NFAAgent{
		logger:         opts.Logger,
		modelProviders: opts.ModelProviders,
		defaultModel:   opts.DefaultModel,

		sessions: map[acp.SessionId]*Session{},
	}
}

// NFAAgent NFA Agent
type NFAAgent struct {
	lock sync.RWMutex

	logger         logr.Logger
	modelProviders []models.ModelProvider
	defaultModel   string

	conn *acp.AgentSideConnection
	g    *genkit.Genkit

	availableModels []string
	tools           []ai.ToolRef

	sessions map[acp.SessionId]*Session
}

// Session 会话
type Session struct {
	lock sync.RWMutex

	id           acp.SessionId
	cancelPrompt context.CancelFunc

	history []*ai.Message
}

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
	messages := session.history
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	session.cancelPrompt = cancel
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
	if prompt == "" {
		return acp.PromptResponse{StopReason: acp.StopReasonEndTurn}, nil
	}

	genOpts := []ai.GenerateOption{ai.WithPrompt(prompt)}
	a.logger.Info("prompt turn start")
	for {
		genOpts = append(genOpts,
			ai.WithMessages(messages...),
			ai.WithModelName(modelName),
			ai.WithTools(a.tools...),
			ai.WithReturnToolRequests(true),
			ai.WithStreaming(a.handleStreamChunk(params.SessionId)),
		)

		// 模型生成
		resp, err := genkit.Generate(ctx, a.g, genOpts...)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return acp.PromptResponse{StopReason: acp.StopReasonCancelled}, nil
			}
			return acp.PromptResponse{StopReason: acp.StopReasonRefusal}, err
		}
		messages = resp.History()
		genOpts = nil

		toolRequests := resp.ToolRequests()
		if len(toolRequests) == 0 {
			a.logger.Info("prompt turn end")
			break
		}

		// 调用工具
		var parts []*ai.Part
		for _, toolReq := range resp.ToolRequests() {
			part, err := a.handleToolRequest(ctx, params.SessionId, toolReq)
			if err != nil {
				a.logger.Error(err, "tool call error")
			}
			if part != nil {
				parts = append(parts, part)
			}
		}

		messages = append(messages, ai.NewMessage(ai.RoleTool, nil, parts...))
	}

	return acp.PromptResponse{StopReason: acp.StopReasonEndTurn}, nil
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

// handleToolRequest 处理工具请求
func (a *NFAAgent) handleToolRequest(ctx context.Context, sessionID acp.SessionId, req *ai.ToolRequest) (*ai.Part, error) {
	callID := acp.ToolCallId(uuid.New().String())
	inputRaw, _ := json.Marshal(req.Input)
	a.logger.Info(fmt.Sprintf("call tool %s: id: %s, input: %s", req.Name, callID, string(inputRaw)))
	if err := a.conn.SessionUpdate(ctx, acp.SessionNotification{
		SessionId: sessionID,
		Update: acp.StartToolCall(
			callID, req.Name,
			acp.WithStartContent([]acp.ToolCallContent{
				{Content: &acp.ToolCallContentContent{Content: acp.TextBlock(string(inputRaw))}},
			}),
			acp.WithStartStatus(acp.ToolCallStatusInProgress),
		),
	}); err != nil {
		return nil, fmt.Errorf("session update error: %w", err)
	}

	tool := genkit.LookupTool(a.g, req.Name)
	if tool == nil {
		// 找不到工具
		if err := a.conn.SessionUpdate(ctx, acp.SessionNotification{
			SessionId: sessionID,
			Update: acp.UpdateToolCall(
				callID,
				acp.WithUpdateStatus(acp.ToolCallStatusFailed),
				acp.WithUpdateContent([]acp.ToolCallContent{
					{
						Content: &acp.ToolCallContentContent{
							Content: acp.TextBlock(fmt.Sprintf("tool %s not found", req.Name)),
						},
					},
				}),
			),
		}); err != nil {
			return nil, fmt.Errorf("session update error: %w", err)
		}
		return nil, fmt.Errorf("tool %s not found", req.Name)
	}

	output, err := tool.RunRaw(ctx, req.Input)
	if err != nil {
		if err := a.conn.SessionUpdate(ctx, acp.SessionNotification{
			SessionId: sessionID,
			Update: acp.UpdateToolCall(
				callID,
				acp.WithUpdateStatus(acp.ToolCallStatusFailed),
				acp.WithUpdateContent([]acp.ToolCallContent{
					{
						Content: &acp.ToolCallContentContent{
							Content: acp.TextBlock(fmt.Sprintf("call tool %s error: %s", req.Name, err.Error())),
						},
					},
				}),
			),
		}); err != nil {
			return nil, fmt.Errorf("session update error: %w", err)
		}
		return nil, fmt.Errorf("call tool %s error: %w", req.Name, err)
	}

	ret := ai.NewToolResponsePart(&ai.ToolResponse{
		Name:   req.Name,
		Ref:    req.Ref,
		Output: output,
	})

	outputRaw, _ := json.Marshal(output)
	if err := a.conn.SessionUpdate(ctx, acp.SessionNotification{
		SessionId: sessionID,
		Update: acp.UpdateToolCall(
			callID,
			acp.WithUpdateStatus(acp.ToolCallStatusCompleted),
			acp.WithUpdateContent([]acp.ToolCallContent{
				{Content: &acp.ToolCallContentContent{Content: acp.TextBlock(string(outputRaw))}},
			}),
		),
	}); err != nil {
		return ret, fmt.Errorf("session update error: %w", err)
	}

	return ret, nil
}

// handleStreamChunk 处理模型流输出
func (a *NFAAgent) handleStreamChunk(sessionID acp.SessionId) ai.ModelStreamCallback {
	return func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var reasoning strings.Builder
		var text strings.Builder
		for _, part := range chunk.Content {
			if part.IsReasoning() {
				reasoning.WriteString(part.Text)
			}
			if part.IsText() || part.IsText() {
				text.WriteString(part.Text)
			}
		}
		if reasoning.Len() > 0 {
			a.logger.Info(fmt.Sprintf("output reasoning: %s", reasoning.String()))
			if err := a.conn.SessionUpdate(ctx, acp.SessionNotification{
				SessionId: sessionID,
				Update:    acp.UpdateAgentThoughtText(reasoning.String()),
			}); err != nil {
				return fmt.Errorf("session update error: %w", err)
			}
			return nil
		}
		if text.Len() > 0 {
			a.logger.Info(fmt.Sprintf("output text: %s", text.String()))
			if err := a.conn.SessionUpdate(ctx, acp.SessionNotification{
				SessionId: sessionID,
				Update:    acp.UpdateAgentMessageText(text.String()),
			}); err != nil {
				return fmt.Errorf("session update error: %w", err)
			}
		}

		return nil
	}
}
