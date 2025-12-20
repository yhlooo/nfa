package agents

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/coder/acp-go-sdk"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/ollama"
	"github.com/google/uuid"

	"github.com/yhlooo/nfa/pkg/acputil"
	"github.com/yhlooo/nfa/pkg/version"
)

// Options Agent 运行选项
type Options struct{}

// NewNFA 创建 NFA Agent
func NewNFA(_ Options) *NFAAgent {
	return &NFAAgent{}
}

// NFAAgent NFA Agent
type NFAAgent struct {
	lock sync.RWMutex

	conn *acp.AgentSideConnection
	g    *genkit.Genkit

	sessionID    acp.SessionId
	cancelPrompt context.CancelFunc
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

	if a.g == nil {
		o := &ollama.Ollama{
			ServerAddress: "http://localhost:11434",
			Timeout:       120,
		}
		a.g = genkit.Init(ctx, genkit.WithPlugins(o))
		o.DefineModel(a.g,
			ollama.ModelDefinition{
				Name: "qwen3:14b",
				Type: "chat",
			}, nil)
	}

	return acp.InitializeResponse{
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

	if a.cancelPrompt != nil {
		return acp.NewSessionResponse{}, fmt.Errorf(
			"%w: session %q in prompting, must cancel first",
			acputil.ErrInPrompting, a.sessionID,
		)
	}

	a.sessionID = acp.SessionId(uuid.New().String())

	return acp.NewSessionResponse{
		SessionId: a.sessionID,
	}, nil
}

// SetSessionMode 设置会话模式
func (a *NFAAgent) SetSessionMode(_ context.Context, _ acp.SetSessionModeRequest) (acp.SetSessionModeResponse, error) {
	return acp.SetSessionModeResponse{}, fmt.Errorf("%w: method session/set_mode not supported", acputil.ErrNotSupported)
}

// Prompt 对话
func (a *NFAAgent) Prompt(ctx context.Context, params acp.PromptRequest) (acp.PromptResponse, error) {
	a.lock.Lock()
	if params.SessionId != a.sessionID {
		a.lock.Unlock()
		return acp.PromptResponse{}, fmt.Errorf(
			"%w: session %q not found",
			acputil.ErrSessionNotFound, params.SessionId,
		)
	}
	if a.cancelPrompt != nil {
		a.lock.Unlock()
		return acp.PromptResponse{}, fmt.Errorf(
			"%w: session %q already in prompting",
			acputil.ErrInPrompting, a.sessionID,
		)
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	a.cancelPrompt = cancel
	defer func() {
		a.lock.Lock()
		a.cancelPrompt = nil
		a.lock.Unlock()
	}()
	a.lock.Unlock()

	prompt := ""
	for _, content := range params.Prompt {
		switch {
		case content.Text != nil:
			prompt += content.Text.Text + "\n"
		}
	}

	_, err := genkit.Generate(ctx, a.g,
		ai.WithPrompt(prompt),
		ai.WithModelName("ollama/qwen3:14b"),
		ai.WithStreaming(func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
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
				if err := a.conn.SessionUpdate(ctx, acp.SessionNotification{
					SessionId: params.SessionId,
					Update:    acp.UpdateAgentThoughtText(reasoning.String()),
				}); err != nil {
					return fmt.Errorf("session update error: %w", err)
				}
				return nil
			}
			if text.Len() > 0 {
				if err := a.conn.SessionUpdate(ctx, acp.SessionNotification{
					SessionId: params.SessionId,
					Update:    acp.UpdateAgentMessageText(chunk.Text()),
				}); err != nil {
					return fmt.Errorf("session update error: %w", err)
				}
			}

			return nil
		}),
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return acp.PromptResponse{StopReason: acp.StopReasonCancelled}, nil
		}
		return acp.PromptResponse{}, err
	}

	return acp.PromptResponse{StopReason: acp.StopReasonEndTurn}, nil
}

// Cancel 取消
func (a *NFAAgent) Cancel(_ context.Context, params acp.CancelNotification) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.sessionID != params.SessionId {
		return fmt.Errorf("%w: session %q not found", acputil.ErrSessionNotFound, params.SessionId)
	}

	if a.cancelPrompt != nil {
		a.cancelPrompt()
		a.cancelPrompt = nil
	}

	return nil
}
