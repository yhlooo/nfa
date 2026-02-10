package models

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/ollama"
)

// OllamaOptions Ollama 选项
type OllamaOptions struct {
	// Ollama 服务端地址
	//
	// 默认 http://localhost:11434
	ServerAddress string `json:"serverAddress,omitempty"`
	// 模型响应超时时间，秒
	//
	// 默认 300
	Timeout int `json:"timeout,omitempty"`
	// 模型列表
	Models []ModelConfig `json:"models,omitempty"`
}

// Complete 使用默认值补全选项
func (opts *OllamaOptions) Complete() {
	if opts.ServerAddress == "" {
		opts.ServerAddress = "http://localhost:11434"
	}
	if opts.Timeout == 0 {
		opts.Timeout = 300
	}
}

// OllamaPlugin 基于选项创建 Ollama 插件
func (opts *OllamaOptions) OllamaPlugin() *ollama.Ollama {
	opts.Complete()
	return &ollama.Ollama{
		ServerAddress: opts.ServerAddress,
		Timeout:       opts.Timeout,
	}
}

// RegisterModels 注册模型
func (opts *OllamaOptions) RegisterModels(
	ctx context.Context,
	g *genkit.Genkit,
	plugin *ollama.Ollama,
) ([]string, error) {
	if len(opts.Models) == 0 {
		return nil, nil // 空配置不注册任何模型
	}

	var definedModels []string
	for _, modelConfig := range opts.Models {
		m := plugin.DefineModel(g, ollama.ModelDefinition{
			Name: modelConfig.Name,
			Type: "chat",
		}, &ai.ModelOptions{
			Label: modelConfig.Name,
			Supports: &ai.ModelSupports{
				Multiturn:  true,
				SystemRole: true,
				Tools:      true,
			},
		})
		definedModels = append(definedModels, m.Name())
	}

	return definedModels, nil
}
