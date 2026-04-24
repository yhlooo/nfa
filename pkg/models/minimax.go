package models

import (
	"context"

	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/nfa/pkg/genkitplugins/oai"
)

const (
	// MinimaxProviderName MiniMax 模型供应商名
	MinimaxProviderName = "minimax"
	// MinimaxBaseURL MiniMax 默认 API 地址
	MinimaxBaseURL = "https://api.minimaxi.com/v1"
)

// MinimaxModels 建议的 MiniMax 模型
func MinimaxModels() []ModelConfig {
	return []ModelConfig{
		{
			Name:      "minimax-m2.7",
			Reasoning: true,
			Prices: ModelPrices{
				Input:  2.1,
				Output: 8.4,
				Cached: 0.42,
			},
			ContextWindow: 200000,
		},
	}
}

// MinimaxOptions MiniMax 模型选项
type MinimaxOptions struct {
	// API 地址
	BaseURL string `json:"baseURL,omitempty"`
	// API 密钥
	APIKey string `json:"apiKey"`
	// 模型列表
	Models []ModelConfig `json:"models,omitempty"`
}

// Complete 使用默认值补全选项
func (opts *MinimaxOptions) Complete() {
	if opts.BaseURL == "" {
		opts.BaseURL = MinimaxBaseURL
	}
}

// Plugin 基于选项创建 OpenAICompatible 插件
func (opts *MinimaxOptions) Plugin() *oai.OpenAICompatible {
	opts.Complete()
	return &oai.OpenAICompatible{
		Provider: MinimaxProviderName,
		BaseURL:  opts.BaseURL,
		APIKey:   opts.APIKey,
	}
}

// RegisterModels 注册模型
func (opts *MinimaxOptions) RegisterModels(
	_ context.Context,
	g *genkit.Genkit,
	plugin *oai.OpenAICompatible,
) ([]ModelConfig, error) {
	definedModels := map[string]struct{}{}
	for _, m := range opts.Models {
		definedModels[m.Name] = struct{}{}
	}

	// 注册建议模型
	for _, m := range MinimaxModels() {
		if _, ok := definedModels[m.Name]; !ok {
			opts.Models = append(opts.Models, m)
		}
	}

	return (&OpenAICompatibleOptions{
		Name:    MinimaxProviderName,
		BaseURL: opts.BaseURL,
		APIKey:  opts.APIKey,
		Models:  opts.Models,
	}).RegisterModels(g, plugin)
}
