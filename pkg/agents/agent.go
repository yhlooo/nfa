package agents

import (
	"context"
	"sync"

	"github.com/coder/acp-go-sdk"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-logr/logr"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/yhlooo/nfa/pkg/agents/flows"
	"github.com/yhlooo/nfa/pkg/models"
	"github.com/yhlooo/nfa/pkg/skills"
	"github.com/yhlooo/nfa/pkg/tools/alphavantage"
	"github.com/yhlooo/nfa/pkg/tools/websearch"
)

const loggerName = "agent"

// Options Agent 运行选项
type Options struct {
	Logger           logr.Logger
	Localizer        *i18n.Localizer
	ModelProviders   []models.ModelProvider
	DataProviders    DataProviders
	DefaultModels    models.Models
	DataRoot         string
	MaxContextWindow int64
}

// DataProviders 数据供应商配置
type DataProviders struct {
	AlphaVantage    *alphavantage.Options             `json:"alphaVantage,omitempty"`
	TencentCloudWSA *websearch.TencentCloudWSAOptions `json:"tcloudWSA,omitempty"`
}

// Complete 使用默认值补全选项
func (opts *Options) Complete() {
	if len(opts.ModelProviders) == 0 {
		opts.ModelProviders = append(opts.ModelProviders, models.ModelProvider{Ollama: &models.OllamaOptions{}})
	}
	if opts.DataRoot == "" {
		opts.DataRoot = ".nfa"
	}
	if opts.MaxContextWindow == 0 {
		opts.MaxContextWindow = 200000
	}
}

// NewNFA 创建 NFA Agent
func NewNFA(opts Options) *NFAAgent {
	opts.Complete()
	return &NFAAgent{
		opts:      opts,
		logger:    opts.Logger.WithName(loggerName),
		localizer: opts.Localizer,
		sessions:  map[acp.SessionId]*Session{},
	}
}

// NFAAgent NFA Agent
type NFAAgent struct {
	opts Options

	lock sync.RWMutex

	logger      logr.Logger
	localizer   *i18n.Localizer
	client      acp.Client
	g           *genkit.Genkit
	skillLoader *skills.SkillLoader

	availableModels []models.ModelConfig
	availableTools  []ai.ToolRef

	chatFlow flows.ChatFlow

	sessions map[acp.SessionId]*Session
}

// Session 会话
type Session struct {
	lock sync.RWMutex

	id           acp.SessionId
	cancelPrompt context.CancelFunc

	currentModels     models.Models
	history           []*ai.Message
	modelUsage        ai.GenerationUsage
	lastContextWindow int64
}
