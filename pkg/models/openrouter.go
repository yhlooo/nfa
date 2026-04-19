package models

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/nfa/pkg/genkitplugins/oai"
)

const (
	// OpenRouterProviderName OpenRouter 模型供应商名
	OpenRouterProviderName = "openrouter"
	// OpenRouterBaseURL OpenRouter 默认 API 地址
	OpenRouterBaseURL = "https://openrouter.ai/api/v1"
)

// OpenRouterModels 建议的 OpenRouter 模型
func OpenRouterModels(_ context.Context) []ModelConfig {
	return []ModelConfig{
		{
			Name:      "google/gemini-3.1-pro-preview",
			Reasoning: true,
			Vision:    true,
			Cost: ModelCost{
				Input:  2, // <= 200K
				Output: 12,
				Cached: 0.2,
			},
			ContextWindow:   1050000,
			MaxOutputTokens: 65500,
		},
		{
			Name:      "openai/gpt-5.4",
			Reasoning: true,
			Vision:    true,
			Cost: ModelCost{
				Input:  2.5, // 272K
				Output: 15,
				Cached: 0.25,
			},
			ContextWindow:   1050000,
			MaxOutputTokens: 128000,
		},
		{
			Name:      "anthropic/claude-sonnet-4.6",
			Reasoning: true,
			Vision:    true,
			Cost: ModelCost{
				Input:  3,
				Output: 15,
				Cached: 0.3,
			},
			ContextWindow:   1000000,
			MaxOutputTokens: 128000,
		},
		{
			Name:      "anthropic/claude-opus-4.7",
			Reasoning: true,
			Vision:    true,
			Cost: ModelCost{
				Input:  5,
				Output: 25,
				Cached: 0.5,
			},
			ContextWindow:   1000000,
			MaxOutputTokens: 128000,
		},
	}
}

// OpenRouterOptions OpenRouter 模型选项
type OpenRouterOptions struct {
	// API 地址
	BaseURL string `json:"baseURL,omitempty"`
	// API 密钥
	APIKey string `json:"apiKey"`
	// 模型列表
	Models []ModelConfig `json:"models,omitempty"`
}

// Complete 使用默认值补全选项
func (opts *OpenRouterOptions) Complete() {
	if opts.BaseURL == "" {
		opts.BaseURL = OpenRouterBaseURL
	}
}

// Plugin 基于选项创建 OpenAICompatible 插件
func (opts *OpenRouterOptions) Plugin() *oai.OpenAICompatible {
	opts.Complete()
	return &oai.OpenAICompatible{
		Provider: OpenRouterProviderName,
		BaseURL:  opts.BaseURL,
		APIKey:   opts.APIKey,
	}
}

// RegisterModels 注册模型
func (opts *OpenRouterOptions) RegisterModels(
	ctx context.Context,
	g *genkit.Genkit,
	plugin *oai.OpenAICompatible,
) ([]ModelConfig, error) {
	definedModels := map[string]struct{}{}
	for _, m := range opts.Models {
		definedModels[m.Name] = struct{}{}
	}

	// 注册建议模型
	for _, m := range OpenRouterModels(ctx) {
		if _, ok := definedModels[m.Name]; !ok {
			opts.Models = append(opts.Models, m)
		}
	}

	var registeredModels []ModelConfig
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
			EnableReasoningExtraFields: map[string]any{
				"reasoning": map[string]any{"effort": "high"},
			},
			DisableReasoningExtraFields: map[string]any{
				"reasoning": map[string]any{"effort": "none"},
			},
			ReasoningContentField: "reasoning",
		})

		registeredModel := cfg
		registeredModel.Name = m.Name()
		registeredModels = append(registeredModels, registeredModel)
	}

	return registeredModels, nil
}
