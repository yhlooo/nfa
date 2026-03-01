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
	Logger         logr.Logger
	Localizer      *i18n.Localizer
	ModelProviders []models.ModelProvider
	DataProviders  []DataProvider
	DefaultModels  models.Models
	DataRoot       string
}

// DataProvider 数据供应商配置
type DataProvider struct {
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
}

// NewNFA 创建 NFA Agent
func NewNFA(opts Options) *NFAAgent {
	opts.Complete()
	return &NFAAgent{
		logger:         opts.Logger.WithName(loggerName),
		localizer:      opts.Localizer,
		modelProviders: opts.ModelProviders,
		dataProviders:  opts.DataProviders,
		defaultModels:  opts.DefaultModels,
		dataRoot:       opts.DataRoot,

		sessions: map[acp.SessionId]*Session{},
	}
}

// NFAAgent NFA Agent
type NFAAgent struct {
	lock sync.RWMutex

	logger         logr.Logger
	localizer      *i18n.Localizer
	modelProviders []models.ModelProvider
	dataProviders  []DataProvider
	defaultModels  models.Models
	dataRoot       string

	conn        ACPAgentSideConnection
	g           *genkit.Genkit
	skillLoader *skills.SkillLoader

	availableModels []models.ModelConfig
	availableTools  []ai.ToolRef

	chatFlow      flows.ChatFlow
	routingFlow   flows.TopicRoutingFlow
	summarizeFlow flows.SummarizeFlow

	sessions map[acp.SessionId]*Session
}

// Session 会话
type Session struct {
	lock sync.RWMutex

	id           acp.SessionId
	cancelPrompt context.CancelFunc

	history    []*ai.Message
	modelUsage ai.GenerationUsage
}
