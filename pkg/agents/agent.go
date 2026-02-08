package agents

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/coder/acp-go-sdk"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-logr/logr"

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
	ModelProviders []models.ModelProvider
	DataProviders  []DataProvider
	DefaultModels  models.Models
	DataRoot       string // nfa 数据目录，默认 ~/.nfa
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
		homeDir, err := os.UserHomeDir()
		if err == nil {
			opts.DataRoot = filepath.Join(homeDir, ".nfa")
		}
	}
}

// NewNFA 创建 NFA Agent
func NewNFA(opts Options) *NFAAgent {
	opts.Complete()
	return &NFAAgent{
		logger:         opts.Logger.WithName(loggerName),
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
	modelProviders []models.ModelProvider
	dataProviders  []DataProvider
	defaultModels  models.Models
	dataRoot       string // nfa 数据目录

	conn        *acp.AgentSideConnection
	g           *genkit.Genkit
	skillLoader *skills.SkillLoader

	availableModels []string
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
