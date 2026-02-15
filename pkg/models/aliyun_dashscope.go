package models

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/nfa/pkg/genkitplugins/oai"
)

const (
	// DashScopeProviderName 阿里云模型供应商名（又称 dashscope 、灵积）
	DashScopeProviderName = "aliyun"
	// DashScopeBaseURL 阿里云默认 API 地址
	DashScopeBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
)

// DashScopeModels 建议的阿里云模型
var DashScopeModels = []ModelConfig{
	{
		Name:      "qwen3-max",
		Reasoning: true,
		Cost: ModelCost{
			Input:  0.0025,
			Output: 0.01,
			Cached: 0.0005,
		},
		ContextWindow:   256000,
		MaxOutputTokens: 64000,
	},
	{
		Name: "qwen3-coder-plus",
		Cost: ModelCost{
			Input:  0.004,
			Output: 0.016,
			Cached: 0.0008,
		},
		ContextWindow:   1000000,
		MaxOutputTokens: 64000,
	},
	{
		Name: "qwen3-coder-flash",
		Cost: ModelCost{
			Input:  0.001,
			Output: 0.004,
			Cached: 0.0002,
		},
		ContextWindow:   1000000,
		MaxOutputTokens: 64000,
	},
	{
		Name:      "qwen3-vl-plus",
		Reasoning: true,
		Vision:    true,
		Cost: ModelCost{
			Input:  0.001,
			Output: 0.01,
			Cached: 0.0002,
		},
		ContextWindow:   256000,
		MaxOutputTokens: 32000,
	},
	{
		Name:      "qwen3-vl-flash",
		Reasoning: true,
		Vision:    true,
		Cost: ModelCost{
			Input:  0.00015,
			Output: 0.0015,
			Cached: 0.00003,
		},
		ContextWindow:   256000,
		MaxOutputTokens: 32000,
	},
}

// DashScopeOptions 阿里云模型选项
type DashScopeOptions struct {
	// API 地址
	BaseURL string `json:"baseURL,omitempty"`
	// API 密钥
	APIKey string `json:"apiKey"`
	// 模型列表
	Models []ModelConfig `json:"models,omitempty"`
}

// Complete 使用默认值补全选项
func (opts *DashScopeOptions) Complete() {
	if opts.BaseURL == "" {
		opts.BaseURL = DashScopeBaseURL
	}

	definedModels := map[string]struct{}{}
	for _, m := range opts.Models {
		definedModels[m.Name] = struct{}{}
	}

	// 注册建议模型
	for _, m := range DashScopeModels {
		if _, ok := definedModels[m.Name]; !ok {
			opts.Models = append(opts.Models, m)
		}
	}
}

// Plugin 基于选项创建 OpenAICompatible 插件
func (opts *DashScopeOptions) Plugin() *oai.OpenAICompatible {
	opts.Complete()
	return &oai.OpenAICompatible{
		Provider: DashScopeProviderName,
		BaseURL:  opts.BaseURL,
		APIKey:   opts.APIKey,
	}
}

// RegisterModels 注册模型
func (opts *DashScopeOptions) RegisterModels(
	g *genkit.Genkit,
	plugin *oai.OpenAICompatible,
) ([]string, error) {
	var definedModels []string
	for _, cfg := range opts.Models {
		m := plugin.DefineModel(g, oai.ModelOptions{
			ModelOptions: ai.ModelOptions{
				Label: cfg.Name,
				Supports: &ai.ModelSupports{
					Multiturn:  true,
					Tools:      true,
					SystemRole: true,
					Media:      true,
					ToolChoice: true,
				},
			},
			Reasoning: cfg.Reasoning,
			ReasoningExtraFields: map[string]any{
				"enable_thinking": true,
			},
			ReasoningContentField: "reasoning_content",
		})
		definedModels = append(definedModels, m.Name())
	}

	return definedModels, nil
}
