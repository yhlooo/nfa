package models

import (
	"github.com/yhlooo/nfa/pkg/genkitplugins/deepseek"
)

// DeepseekOptions Deepseek 选项
type DeepseekOptions struct {
	// Deepseek API 密钥
	APIKey string `json:"apiKey"`
	// 模型列表
	Models []ModelConfig `json:"models,omitempty"`
}

// DeepseekPlugin 基于选项创建 Deepseek 插件
func (opts *DeepseekOptions) DeepseekPlugin() *deepseek.Deepseek {
	return &deepseek.Deepseek{
		APIKey: opts.APIKey,
		Models: modelConfigsToStrings(opts.Models),
	}
}

// modelConfigsToStrings 转换模型配置列表为字符串列表
func modelConfigsToStrings(configs []ModelConfig) []string {
	if len(configs) == 0 {
		return nil
	}
	names := make([]string, len(configs))
	for i, cfg := range configs {
		names[i] = cfg.Name
	}
	return names
}
