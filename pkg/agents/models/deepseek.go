package models

import (
	"github.com/yhlooo/nfa/pkg/genkitplugins/deepseek"
)

// DeepseekOptions Deepseek 选项
type DeepseekOptions struct {
	// Deepseek API 密钥
	APIKey string `json:"apiKey"`
}

// DeepseekPlugin 基于选项创建 Deepseek 插件
func (opts *DeepseekOptions) DeepseekPlugin() *deepseek.Deepseek {
	return &deepseek.Deepseek{
		APIKey: opts.APIKey,
	}
}
