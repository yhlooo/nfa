package agents

import (
	"context"

	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	oai "github.com/firebase/genkit/go/plugins/compat_oai"
	"github.com/firebase/genkit/go/plugins/ollama"
	"github.com/go-logr/logr"

	"github.com/yhlooo/nfa/pkg/agents/models"
)

// InitGenkit 初始化 genkit
func (a *NFAAgent) InitGenkit(ctx context.Context) {
	ctx = logr.NewContext(ctx, a.logger)

	if a.g != nil {
		return
	}

	// 确定插件
	var (
		ollamaPlugin   = &ollama.Ollama{}
		deepseekPlugin = &oai.OpenAICompatible{}
		plugins        []api.Plugin
	)
	for _, p := range a.modelProviders {
		switch {
		case p.Ollama != nil:
			ollamaPlugin = p.Ollama.OllamaPlugin()
			plugins = append(plugins, ollamaPlugin)
		case p.Deepseek != nil:
			deepseekPlugin = p.Deepseek.DeepseekPlugin()
			plugins = append(plugins, deepseekPlugin)
		}
	}

	a.g = genkit.Init(ctx, genkit.WithPlugins(plugins...))

	// 注册模型
	for _, p := range a.modelProviders {
		switch {
		case p.Ollama != nil:
			modelNames, err := p.Ollama.DefineModels(ctx, a.g, ollamaPlugin)
			if err != nil {
				a.logger.Error(err, "define ollama models error")
			}
			a.availableModels = append(a.availableModels, modelNames...)
		case p.Deepseek != nil:
			modelNames, err := models.ListOpenAICompatibleModels(ctx, deepseekPlugin)
			if err != nil {
				a.logger.Error(err, "list deepseek models error")
			}
			a.availableModels = append(a.availableModels, modelNames...)
		}
	}

	if a.defaultModel == "" && len(a.availableModels) > 0 {
		a.defaultModel = a.availableModels[0]
	}

	// 注册工具
	a.tools = append(a.tools, DefineToolCheckAssetPriceTrends(a.g))
}

// AvailableModels 获取可用模型名列表
func (a *NFAAgent) AvailableModels() []string {
	if len(a.availableModels) == 0 {
		return nil
	}

	ret := make([]string, len(a.availableModels))
	copy(ret, a.availableModels)
	return ret
}
