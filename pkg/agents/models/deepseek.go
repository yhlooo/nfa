package models

import (
	oai "github.com/firebase/genkit/go/plugins/compat_oai"
)

const (
	DeepseekProviderName = "deepseek"
	DeepseekBaseURL      = "https://api.deepseek.com"
)

// DeepseekOptions Deepseek 选项
type DeepseekOptions struct {
	// Deepseek API 密钥
	APIKey string `json:"apiKey"`
}

// DeepseekPlugin 基于选项创建 Deepseek 插件
func (opts *DeepseekOptions) DeepseekPlugin() *oai.OpenAICompatible {
	return &oai.OpenAICompatible{
		Provider: DeepseekProviderName,
		BaseURL:  DeepseekBaseURL,
		APIKey:   opts.APIKey,
	}
}
