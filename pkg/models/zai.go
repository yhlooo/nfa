package models

import (
	"context"

	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/nfa/pkg/genkitplugins/oai"
)

const (
	// ZAIProviderName 智谱模型供应商名
	ZAIProviderName = "z-ai"
	// ZAIBaseURL 智谱默认 API 地址
	ZAIBaseURL = "https://open.bigmodel.cn/api/paas/v4/"
)

// ZAIModels 建议的智谱 AI 模型
func ZAIModels(ctx context.Context) []ModelConfig {
	return []ModelConfig{
		{
			Name:      "glm-5-turbo",
			Reasoning: true,
			Cost: ModelCost{
				Input:  7, // 128K-256K
				Output: 26,
				Cached: 1.8,
			},
			ContextWindow:   200000,
			MaxOutputTokens: 128000,
		},
		{
			Name:      "glm-5",
			Reasoning: true,
			Cost: ModelCost{
				Input:  6, // 128K-256K
				Output: 22,
				Cached: 1.5,
			},
			ContextWindow:   200000,
			MaxOutputTokens: 128000,
		},
		{
			Name:      "glm-4.7",
			Reasoning: true,
			Cost: ModelCost{
				Input:  4, // 128K-256K
				Output: 16,
				Cached: 0.8,
			},
			ContextWindow:   200000,
			MaxOutputTokens: 128000,
		},
		{
			Name:      "glm-4.7-flashx",
			Reasoning: true,
			Cost: ModelCost{
				Input:  0.5,
				Output: 3,
				Cached: 0.1,
			},
			ContextWindow:   200000,
			MaxOutputTokens: 128000,
		},
		{
			Name:      "glm-4.6v",
			Reasoning: true,
			Vision:    true,
			Cost: ModelCost{
				Input:  2, // 128K-256K
				Output: 6,
				Cached: 0.4,
			},
			ContextWindow:   128000,
			MaxOutputTokens: 32000,
		},
		{
			Name:      "glm-4.6v-flashx",
			Reasoning: true,
			Vision:    true,
			Cost: ModelCost{
				Input:  0.3, // 128K-256K
				Output: 3,
				Cached: 0.03,
			},
			ContextWindow:   128000,
			MaxOutputTokens: 32000,
		},
	}
}

// ZAIOptions 智谱 AI 模型选项
type ZAIOptions struct {
	// API 地址
	BaseURL string `json:"baseURL,omitempty"`
	// API 密钥
	APIKey string `json:"apiKey"`
	// 模型列表
	Models []ModelConfig `json:"models,omitempty"`
}

// Complete 使用默认值补全选项
func (opts *ZAIOptions) Complete() {
	if opts.BaseURL == "" {
		opts.BaseURL = ZAIBaseURL
	}
}

// Plugin 基于选项创建 OpenAICompatible 插件
func (opts *ZAIOptions) Plugin() *oai.OpenAICompatible {
	opts.Complete()
	return &oai.OpenAICompatible{
		Provider: ZAIProviderName,
		BaseURL:  opts.BaseURL,
		APIKey:   opts.APIKey,
	}
}

// RegisterModels 注册模型
func (opts *ZAIOptions) RegisterModels(
	ctx context.Context,
	g *genkit.Genkit,
	plugin *oai.OpenAICompatible,
) ([]ModelConfig, error) {
	definedModels := map[string]struct{}{}
	for _, m := range opts.Models {
		definedModels[m.Name] = struct{}{}
	}

	// 注册建议模型
	for _, m := range ZAIModels(ctx) {
		if _, ok := definedModels[m.Name]; !ok {
			opts.Models = append(opts.Models, m)
		}
	}

	return (&OpenAICompatibleOptions{
		Name:    ZAIProviderName,
		BaseURL: opts.BaseURL,
		APIKey:  opts.APIKey,
		Models:  opts.Models,
	}).RegisterModels(g, plugin)
}
