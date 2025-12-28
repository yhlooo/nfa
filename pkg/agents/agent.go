package agents

import (
	"context"
	"sync"

	"github.com/coder/acp-go-sdk"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-logr/logr"

	"github.com/yhlooo/nfa/pkg/agents/dataproviders"
	"github.com/yhlooo/nfa/pkg/agents/flows"
	"github.com/yhlooo/nfa/pkg/agents/models"
)

const loggerName = "agent"

// Options Agent 运行选项
type Options struct {
	Logger         logr.Logger
	ModelProviders []models.ModelProvider
	DataProviders  []dataproviders.DataProvider
	DefaultModel   string
	SingleAgent    bool
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
		logger:         opts.Logger.WithName(loggerName),
		modelProviders: opts.ModelProviders,
		dataProviders:  opts.DataProviders,
		defaultModel:   opts.DefaultModel,
		singleAgent:    opts.SingleAgent,

		sessions: map[acp.SessionId]*Session{},
	}
}

// NFAAgent NFA Agent
type NFAAgent struct {
	lock sync.RWMutex

	logger         logr.Logger
	modelProviders []models.ModelProvider
	dataProviders  []dataproviders.DataProvider
	defaultModel   string
	singleAgent    bool

	conn *acp.AgentSideConnection
	g    *genkit.Genkit

	availableModels []string

	allTools                   []ai.ToolRef
	comprehensiveAnalysisTools []ai.ToolRef
	macroeconomicAnalysisTools []ai.ToolRef
	fundamentalAnalysisTools   []ai.ToolRef
	technicalAnalysisTools     []ai.ToolRef

	mainFlow      flows.ChatFlow
	summarizeFlow flows.SummarizeFlow

	sessions map[acp.SessionId]*Session
}

// Session 会话
type Session struct {
	lock sync.RWMutex

	id           acp.SessionId
	cancelPrompt context.CancelFunc

	history []*ai.Message
}
