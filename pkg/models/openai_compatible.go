package models

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/nfa/pkg/genkitplugins/oai"
)

// OpenAICompatibleOptions OpenAI 兼容选项
type OpenAICompatibleOptions struct {
	// 供应商名
	Name string `json:"name"`
	// API 地址
	BaseURL string `json:"baseURL"`
	// API 密钥
	APIKey string `json:"apiKey"`
	// 模型列表
	Models []ModelConfig `json:"models,omitempty"`
}

// OpenAICompatiblePlugin 基于选项创建 OpenAI 兼容插件
func (opts *OpenAICompatibleOptions) OpenAICompatiblePlugin() *oai.OpenAICompatible {
	return &oai.OpenAICompatible{
		Provider: opts.Name,
		BaseURL:  opts.BaseURL,
		APIKey:   opts.APIKey,
	}
}

// RegisterModels 注册模型
func (opts *OpenAICompatibleOptions) RegisterModels(
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
