package models

import (
	oai "github.com/firebase/genkit/go/plugins/compat_oai"
)

// OpenAICompatibleOptions OpenAI 兼容选项
type OpenAICompatibleOptions struct {
	Name    string `json:"name"`
	BaseURL string `json:"baseURL"`
	APIKey  string `json:"apiKey"`
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
