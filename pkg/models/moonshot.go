package models

import (
	"context"

	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/nfa/pkg/genkitplugins/oai"
	"github.com/yhlooo/nfa/pkg/i18n"
)

const (
	// MoonshotProviderName 月之暗面模型供应商名
	MoonshotProviderName = "moonshotai"
	// MoonshotBaseURL 月之暗面默认 API 地址
	MoonshotBaseURL = "https://api.moonshot.cn/v1"
)

// MoonshotModels 建议的月之暗面模型
func MoonshotModels(ctx context.Context) []ModelConfig {
	return []ModelConfig{
		{
			Name:        "kimi-k2.5",
			Description: i18n.TContext(ctx, MsgModelDescKimiK25),
			Reasoning:   true,
			Vision:      true,
			Cost: ModelCost{
				Input:  4,
				Output: 21,
				Cached: 0.7,
			},
			ContextWindow:   256000,
			MaxOutputTokens: 256000,
		},
	}
}

// MoonshotOptions 月之暗面模型选项
type MoonshotOptions struct {
	// API 地址
	BaseURL string `json:"baseURL,omitempty"`
	// API 密钥
	APIKey string `json:"apiKey"`
	// 模型列表
	Models []ModelConfig `json:"models,omitempty"`
}

// Complete 使用默认值补全选项
func (opts *MoonshotOptions) Complete() {
	if opts.BaseURL == "" {
		opts.BaseURL = MoonshotBaseURL
	}
}

// Plugin 基于选项创建 OpenAICompatible 插件
func (opts *MoonshotOptions) Plugin() *oai.OpenAICompatible {
	opts.Complete()
	return &oai.OpenAICompatible{
		Provider: MoonshotProviderName,
		BaseURL:  opts.BaseURL,
		APIKey:   opts.APIKey,
	}
}

// RegisterModels 注册模型
func (opts *MoonshotOptions) RegisterModels(
	ctx context.Context,
	g *genkit.Genkit,
	plugin *oai.OpenAICompatible,
) ([]ModelConfig, error) {
	definedModels := map[string]struct{}{}
	for _, m := range opts.Models {
		definedModels[m.Name] = struct{}{}
	}

	// 注册建议模型
	for _, m := range MoonshotModels(ctx) {
		if _, ok := definedModels[m.Name]; !ok {
			opts.Models = append(opts.Models, m)
		}
	}

	return (&OpenAICompatibleOptions{
		Name:    MoonshotProviderName,
		BaseURL: opts.BaseURL,
		APIKey:  opts.APIKey,
		Models:  opts.Models,
	}).RegisterModels(g, plugin)
}
