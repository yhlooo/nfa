package models

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/nfa/pkg/genkitplugins/oai"
)

const (
	DeepseekProviderName = "deepseek"
	DeepseekBaseURL      = "https://api.deepseek.com"
)

// DeepseekOptions Deepseek 选项
type DeepseekOptions struct {
	// Deepseek API 密钥
	APIKey string `json:"apiKey"`
	// 模型列表
	Models []ModelConfig `json:"models,omitempty"`
}

// DeepseekPlugin 基于选项创建 Deepseek 插件
func (opts *DeepseekOptions) DeepseekPlugin() *oai.OpenAICompatible {
	return &oai.OpenAICompatible{
		Provider: DeepseekProviderName,
		BaseURL:  DeepseekBaseURL,
		APIKey:   opts.APIKey,
	}
}

// RegisterModels 注册模型
func (opts *DeepseekOptions) RegisterModels(
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
				"thinking": map[string]any{"type": "enabled"},
			},
			ReasoningContentField: "reasoning_content",
		})
		definedModels = append(definedModels, m.Name())
	}

	return definedModels, nil
}
