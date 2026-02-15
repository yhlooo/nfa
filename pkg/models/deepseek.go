package models

import (
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/nfa/pkg/genkitplugins/oai"
)

const (
	// DeepseekProviderName Deepseek 模型供应商名
	DeepseekProviderName = "deepseek"
	// DeepseekBaseURL Deepseek 默认 API 地址
	DeepseekBaseURL = "https://api.deepseek.com"
)

// DeepSeekModels 建议的 DeepSeek 模型
var DeepSeekModels = []ModelConfig{
	{
		Name:      "deepseek-reasoner",
		Reasoning: true,
		Cost: ModelCost{
			Input:  0.002,
			Output: 0.003,
		},
		ContextWindow:   128000,
		MaxOutputTokens: 64000,
	},
	{
		Name: "deepseek-chat",
		Cost: ModelCost{
			Input:  0.002,
			Output: 0.003,
		},
		ContextWindow:   128000,
		MaxOutputTokens: 8000,
	},
}

// DeepseekOptions Deepseek 选项
type DeepseekOptions struct {
	// API 地址
	BaseURL string `json:"baseURL,omitempty"`
	// API 密钥
	APIKey string `json:"apiKey"`
	// 模型列表
	Models []ModelConfig `json:"models,omitempty"`
}

// Complete 使用默认值补全选项
func (opts *DeepseekOptions) Complete() {
	if opts.BaseURL == "" {
		opts.BaseURL = DeepseekBaseURL
	}

	definedModels := map[string]struct{}{}
	for _, m := range opts.Models {
		definedModels[m.Name] = struct{}{}
	}

	// 注册建议模型
	for _, m := range DeepSeekModels {
		if _, ok := definedModels[m.Name]; !ok {
			opts.Models = append(opts.Models, m)
		}
	}
}

// Plugin 基于选项创建 OpenAICompatible 插件
func (opts *DeepseekOptions) Plugin() *oai.OpenAICompatible {
	opts.Complete()
	return &oai.OpenAICompatible{
		Provider: DeepseekProviderName,
		BaseURL:  opts.BaseURL,
		APIKey:   opts.APIKey,
	}
}

// RegisterModels 注册模型
func (opts *DeepseekOptions) RegisterModels(
	g *genkit.Genkit,
	plugin *oai.OpenAICompatible,
) ([]string, error) {
	return (&OpenAICompatibleOptions{
		Name:    DeepseekProviderName,
		BaseURL: opts.BaseURL,
		APIKey:  opts.APIKey,
		Models:  opts.Models,
	}).RegisterModels(g, plugin)
}
