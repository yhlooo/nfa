package models

import (
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/nfa/pkg/genkitplugins/oai"
)

const (
	// BigModelProviderName 智谱模型供应商名（又称 bigmodel ）
	BigModelProviderName = "zhipu"
	// BigModelBaseURL 智谱默认 API 地址
	BigModelBaseURL = "https://open.bigmodel.cn/api/paas/v4/"
)

// BigModelModels 建议的智谱模型
var BigModelModels = []ModelConfig{
	{
		Name:        "glm-5",
		Description: "智谱 GLM-5，新一代旗舰模型，强大的推理和理解能力",
		Reasoning:   true,
		Cost: ModelCost{
			Input:  0.006,
			Output: 0.022,
			Cached: 0.0015,
		},
		ContextWindow:   200000,
		MaxOutputTokens: 128000,
	},
	{
		Name:        "glm-4.7",
		Description: "智谱 GLM-4.7，高性能通用模型",
		Reasoning:   true,
		Cost: ModelCost{
			Input:  0.004,
			Output: 0.016,
			Cached: 0.0008,
		},
		ContextWindow:   200000,
		MaxOutputTokens: 128000,
	},
	{
		Name:        "glm-4.7-flashx",
		Description: "智谱 GLM-4.7 FlashX，快速响应模型",
		Reasoning:   true,
		Cost: ModelCost{
			Input:  0.0005,
			Output: 0.003,
			Cached: 0.0001,
		},
		ContextWindow:   200000,
		MaxOutputTokens: 128000,
	},
	{
		Name:        "glm-4.6v",
		Description: "智谱 GLM-4.6V，视觉理解多模态模型",
		Reasoning:   true,
		Vision:      true,
		Cost: ModelCost{
			Input:  0.002,
			Output: 0.006,
			Cached: 0.0004,
		},
		ContextWindow:   128000,
		MaxOutputTokens: 32000,
	},
	{
		Name:        "glm-4.6v-flashx",
		Description: "智谱 GLM-4.6V FlashX，快速视觉多模态模型",
		Reasoning:   true,
		Vision:      true,
		Cost: ModelCost{
			Input:  0.0003,
			Output: 0.003,
			Cached: 0.00003,
		},
		ContextWindow:   128000,
		MaxOutputTokens: 32000,
	},
}

// BigModelOptions 智谱选项
type BigModelOptions struct {
	// API 地址
	BaseURL string `json:"baseURL,omitempty"`
	// API 密钥
	APIKey string `json:"apiKey"`
	// 模型列表
	Models []ModelConfig `json:"models,omitempty"`
}

// Complete 使用默认值补全选项
func (opts *BigModelOptions) Complete() {
	if opts.BaseURL == "" {
		opts.BaseURL = BigModelBaseURL
	}

	definedModels := map[string]struct{}{}
	for _, m := range opts.Models {
		definedModels[m.Name] = struct{}{}
	}

	// 注册建议模型
	for _, m := range BigModelModels {
		if _, ok := definedModels[m.Name]; !ok {
			opts.Models = append(opts.Models, m)
		}
	}
}

// Plugin 基于选项创建 OpenAICompatible 插件
func (opts *BigModelOptions) Plugin() *oai.OpenAICompatible {
	opts.Complete()
	return &oai.OpenAICompatible{
		Provider: BigModelProviderName,
		BaseURL:  opts.BaseURL,
		APIKey:   opts.APIKey,
	}
}

// RegisterModels 注册模型
func (opts *BigModelOptions) RegisterModels(
	g *genkit.Genkit,
	plugin *oai.OpenAICompatible,
) ([]ModelConfig, error) {
	return (&OpenAICompatibleOptions{
		Name:    BigModelProviderName,
		BaseURL: opts.BaseURL,
		APIKey:  opts.APIKey,
		Models:  opts.Models,
	}).RegisterModels(g, plugin)
}
