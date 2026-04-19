package models

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/nfa/pkg/genkitplugins/oai"
)

const (
	// QwenProviderName 通义千问模型供应商名
	QwenProviderName = "qwen"
	// QwenBaseURL 通义千问默认 API 地址（也可称 dashscope 、灵积）
	QwenBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
)

// QwenModels 建议的通义千问模型
func QwenModels(ctx context.Context) []ModelConfig {
	return []ModelConfig{
		{
			Name:      "qwen3.5-plus",
			Reasoning: true,
			Vision:    true,
			Cost: ModelCost{
				Input:  2, // 128K-256K
				Output: 12,
			},
			ContextWindow:   1000000,
			MaxOutputTokens: 64000,
		},
		{
			Name:      "qwen3.5-397b-a17b",
			Reasoning: true,
			Vision:    true,
			Cost: ModelCost{
				Input:  3, // 128K-256K
				Output: 18,
			},
			ContextWindow:   256000,
			MaxOutputTokens: 64000,
		},
		{
			Name:      "qwen3.5-flash",
			Reasoning: true,
			Vision:    true,
			Cost: ModelCost{
				Input:  0.8, // 128K-256K
				Output: 8,
			},
			ContextWindow:   1000000,
			MaxOutputTokens: 64000,
		},
		{
			Name:      "qwen3-max",
			Reasoning: true,
			Cost: ModelCost{
				Input:  2.5,
				Output: 10,
				Cached: 0.5,
			},
			ContextWindow:   256000,
			MaxOutputTokens: 64000,
		},
		{
			Name:      "qwen3-vl-plus",
			Reasoning: true,
			Vision:    true,
			Cost: ModelCost{
				Input:  1,
				Output: 10,
				Cached: 0.2,
			},
			ContextWindow:   256000,
			MaxOutputTokens: 32000,
		},
		{
			Name:      "qwen3-vl-flash",
			Reasoning: true,
			Vision:    true,
			Cost: ModelCost{
				Input:  0.15,
				Output: 1.5,
				Cached: 0.03,
			},
			ContextWindow:   256000,
			MaxOutputTokens: 32000,
		},
	}
}

// QwenOptions 通义千问模型选项
type QwenOptions struct {
	// API 地址
	BaseURL string `json:"baseURL,omitempty"`
	// API 密钥
	APIKey string `json:"apiKey"`
	// 模型列表
	Models []ModelConfig `json:"models,omitempty"`
}

// Complete 使用默认值补全选项
func (opts *QwenOptions) Complete() {
	if opts.BaseURL == "" {
		opts.BaseURL = QwenBaseURL
	}
}

// Plugin 基于选项创建 OpenAICompatible 插件
func (opts *QwenOptions) Plugin() *oai.OpenAICompatible {
	opts.Complete()
	return &oai.OpenAICompatible{
		Provider: QwenProviderName,
		BaseURL:  opts.BaseURL,
		APIKey:   opts.APIKey,
	}
}

// RegisterModels 注册模型
func (opts *QwenOptions) RegisterModels(
	ctx context.Context,
	g *genkit.Genkit,
	plugin *oai.OpenAICompatible,
) ([]ModelConfig, error) {
	definedModels := map[string]struct{}{}
	for _, m := range opts.Models {
		definedModels[m.Name] = struct{}{}
	}

	// 注册建议模型
	for _, m := range QwenModels(ctx) {
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
				"enable_thinking": true,
			},
			DisableReasoningExtraFields: map[string]any{
				"enable_thinking": false,
			},
			ReasoningContentField: "reasoning_content",
		})

		registeredModel := cfg
		registeredModel.Name = m.Name()
		registeredModels = append(registeredModels, registeredModel)
	}

	return registeredModels, nil
}
