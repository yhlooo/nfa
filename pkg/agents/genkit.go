package agents

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/ollama"
	"github.com/go-logr/logr"

	"github.com/yhlooo/nfa/pkg/agents/flows"
	"github.com/yhlooo/nfa/pkg/genkitplugins/oai"
	"github.com/yhlooo/nfa/pkg/models"
	"github.com/yhlooo/nfa/pkg/tools/webbrowse"
)

const ChatFlowName = "Chat"

// InitGenkit 初始化 genkit
func (a *NFAAgent) InitGenkit(ctx context.Context) {
	ctx = logr.NewContext(ctx, a.logger)

	if a.g != nil {
		return
	}

	a.g, a.availableModels = NewGenkitWithModels(ctx, a.modelProviders, a.defaultModels)
	if a.defaultModels.Main == "" && len(a.availableModels) > 0 {
		a.defaultModels.Main = a.availableModels[0].Name
	}
	for _, m := range a.availableModels {
		a.logger.Info(fmt.Sprintf("registered model: %s", m.Name))
	}

	// 注册工具
	for _, p := range a.dataProviders {
		switch {
		case p.AlphaVantage != nil:
			alphaVantageTools, err := p.AlphaVantage.RegisterTools(ctx, a.g)
			if err != nil {
				a.logger.Error(err, "register alpha vantage tools error")
				continue
			}
			a.availableTools = append(a.availableTools, alphaVantageTools...)
		case p.TencentCloudWSA != nil:
			searchTool, err := p.TencentCloudWSA.RegisterTool(ctx, a.g)
			if err != nil {
				a.logger.Error(err, "register tencent cloud wsa search tool error")
				continue
			}
			a.availableTools = append(a.availableTools, searchTool)
		}
	}
	// 网页浏览工具
	wb := webbrowse.NewWebBrowser()
	a.availableTools = append(a.availableTools, wb.RegisterTools(a.g)...)

	// 注册 Skill 工具
	a.availableTools = append(a.availableTools, a.skillLoader.DefineSkillTool(a.g))

	for _, t := range a.availableTools {
		a.logger.Info(fmt.Sprintf("registered tool: %s", t.Name()))
	}

	// 注册 flows
	a.chatFlow = flows.DefineSimpleChatFlow(a.g, ChatFlowName, flows.FixedGenerateOptions(
		ai.WithSystemFn(AnalystSystemPrompt(a.skillLoader)),
		ai.WithTools(a.availableTools...),
	))
	a.summarizeFlow = flows.DefineSummarizeFlow(a.g)
	a.routingFlow = flows.DefineTopicRoutingFlow(a.g)
}

// NewGenkitWithModels 创建 genkit 对象并注册模型
func NewGenkitWithModels(
	ctx context.Context,
	providers []models.ModelProvider,
	defaultModels models.Models,
) (*genkit.Genkit, []models.ModelConfig) {
	logger := logr.FromContextOrDiscard(ctx)

	// 确定插件
	var (
		ollamaPlugin = &ollama.Ollama{}
		oaiPlugins   = map[int]*oai.OpenAICompatible{}
		plugins      []api.Plugin
		modelConfigs []models.ModelConfig
	)
	for i, p := range providers {
		switch {
		case p.Ollama != nil:
			ollamaPlugin = p.Ollama.OllamaPlugin()
			plugins = append(plugins, ollamaPlugin)
		case p.Zhipu != nil:
			plugin := p.Zhipu.Plugin()
			plugins = append(plugins, plugin)
			oaiPlugins[i] = plugin
		case p.Aliyun != nil:
			plugin := p.Aliyun.Plugin()
			plugins = append(plugins, plugin)
			oaiPlugins[i] = plugin
		case p.Deepseek != nil:
			plugin := p.Deepseek.Plugin()
			plugins = append(plugins, plugin)
			oaiPlugins[i] = plugin
		case p.OpenAICompatible != nil:
			plugin := p.OpenAICompatible.OpenAICompatiblePlugin()
			plugins = append(plugins, plugin)
			oaiPlugins[i] = plugin
		}
	}

	genkitOpts := []genkit.GenkitOption{
		genkit.WithPlugins(plugins...),
	}
	if defaultModels.GetMain() != "" {
		genkitOpts = append(genkitOpts, genkit.WithDefaultModel(defaultModels.GetMain()))
	}
	g := genkit.Init(ctx, genkitOpts...)

	// 注册模型
	for i, p := range providers {
		switch {
		case p.Ollama != nil:
			registeredModels, err := p.Ollama.RegisterModels(g, ollamaPlugin)
			if err != nil {
				logger.Error(err, "define ollama models error")
				continue
			}
			modelConfigs = append(modelConfigs, registeredModels...)
		case p.Zhipu != nil:
			registeredModels, err := p.Zhipu.RegisterModels(g, oaiPlugins[i])
			if err != nil {
				logger.Error(err, "define zhipu models error")
				continue
			}
			modelConfigs = append(modelConfigs, registeredModels...)
		case p.Aliyun != nil:
			registeredModels, err := p.Aliyun.RegisterModels(g, oaiPlugins[i])
			if err != nil {
				logger.Error(err, "define aliyun models error")
				continue
			}
			modelConfigs = append(modelConfigs, registeredModels...)
		case p.Deepseek != nil:
			registeredModels, err := p.Deepseek.RegisterModels(g, oaiPlugins[i])
			if err != nil {
				logger.Error(err, "define deepseek models error")
				continue
			}
			modelConfigs = append(modelConfigs, registeredModels...)
		case p.OpenAICompatible != nil:
			registeredModels, err := p.OpenAICompatible.RegisterModels(g, oaiPlugins[i])
			if err != nil {
				logger.Error(err, "define openai compatible models error")
				continue
			}
			modelConfigs = append(modelConfigs, registeredModels...)
		}
	}

	// 警告：如果没有配置任何模型
	if len(modelConfigs) == 0 {
		logger.Info("no models configured, please configure models in your config file")
	}

	return g, modelConfigs
}

// AvailableModels 获取可用模型列表
func (a *NFAAgent) AvailableModels() []models.ModelConfig {
	if len(a.availableModels) == 0 {
		return nil
	}

	ret := make([]models.ModelConfig, len(a.availableModels))
	copy(ret, a.availableModels)
	return ret
}
