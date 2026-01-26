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
	"github.com/yhlooo/nfa/pkg/genkitplugins/deepseek"
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
		a.defaultModels.Main = a.availableModels[0]
	}

	for _, m := range a.availableModels {
		a.logger.Info(fmt.Sprintf("registered model: %s", m))
	}

	// 注册工具
	for _, p := range a.dataProviders {
		switch {
		case p.AlphaVantage != nil:
			comprehensiveAnalysisTools,
				macroeconomicAnalysisTools,
				fundamentalAnalysisTools,
				technicalAnalysisTools,
				allTools, err := p.AlphaVantage.RegisterTools(ctx, a.g)
			if err != nil {
				a.logger.Error(err, "register alpha vantage tools error")
				continue
			}
			a.comprehensiveAnalysisTools = append(a.comprehensiveAnalysisTools, comprehensiveAnalysisTools...)
			a.macroeconomicAnalysisTools = append(a.macroeconomicAnalysisTools, macroeconomicAnalysisTools...)
			a.fundamentalAnalysisTools = append(a.fundamentalAnalysisTools, fundamentalAnalysisTools...)
			a.technicalAnalysisTools = append(a.technicalAnalysisTools, technicalAnalysisTools...)
			a.allTools = append(a.allTools, allTools...)
		case p.TencentCloudWSA != nil:
			searchTool, err := p.TencentCloudWSA.RegisterTool(ctx, a.g)
			if err != nil {
				a.logger.Error(err, "register tencent cloud wsa search tool error")
				continue
			}
			a.commonTools = append(a.commonTools, searchTool)
			a.allTools = append(a.allTools, searchTool)
		}
	}

	// 网页浏览工具
	wb := webbrowse.NewWebBrowser()
	webBrowseTool := wb.DefineBrowseTool(a.g)
	a.commonTools = append(a.commonTools, webBrowseTool)
	a.allTools = append(a.allTools, webBrowseTool)

	for _, t := range a.allTools {
		a.logger.Info(fmt.Sprintf("registered tool: %s", t.Name()))
	}

	// 注册 flows
	a.logger.Info("registing flows ...")
	if a.singleAgent || len(a.allTools) < 20 {
		a.logger.Info("using single agent mode")
		a.mainFlow = flows.DefineSimpleChatFlow(a.g, ChatFlowName, flows.FixedGenerateOptions(
			ai.WithSystemFn(AllAroundAnalystSystemPrompt),
			ai.WithTools(a.allTools...),
		))
	} else {
		a.logger.Info("using multi agent mode")
		mainAgent, subAgents := NewDefaultAgents(
			a.commonTools,
			a.comprehensiveAnalysisTools,
			a.macroeconomicAnalysisTools,
			a.fundamentalAnalysisTools,
			a.technicalAnalysisTools,
		)
		a.mainFlow = flows.DefineMultiAgentsChatFlow(a.g, ChatFlowName, mainAgent, subAgents)
	}
	a.summarizeFlow = flows.DefineSummarizeFlow(a.g)
	a.routingFlow = flows.DefineTopicRoutingFlow(a.g)
}

// NewGenkitWithModels 创建 genkit 对象并注册模型
func NewGenkitWithModels(
	ctx context.Context,
	providers []models.ModelProvider,
	defaultModels models.Models,
) (*genkit.Genkit, []string) {
	logger := logr.FromContextOrDiscard(ctx)

	// 确定插件
	var (
		ollamaPlugin   = &ollama.Ollama{}
		deepseekPlugin = &deepseek.Deepseek{}
		plugins        []api.Plugin
		modelNames     []string
	)
	for _, p := range providers {
		switch {
		case p.Ollama != nil:
			ollamaPlugin = p.Ollama.OllamaPlugin()
			plugins = append(plugins, ollamaPlugin)
		case p.Deepseek != nil:
			deepseekPlugin = p.Deepseek.DeepseekPlugin()
			plugins = append(plugins, deepseekPlugin)
		case p.OpenAICompatible != nil:
			plugin := p.OpenAICompatible.OpenAICompatiblePlugin()
			// 注册插件后自动注册模型，这里仅获取模型名
			oaiModelNames, err := models.ListOpenAICompatibleModels(ctx, plugin)
			if err != nil {
				logger.Error(err, fmt.Sprintf("list %s models error", p.OpenAICompatible.Name))
				continue
			}
			modelNames = append(modelNames, oaiModelNames...)
			plugins = append(plugins, plugin)
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
	for _, p := range providers {
		switch {
		case p.Ollama != nil:
			ollamaModelNames, err := p.Ollama.RegisterModels(ctx, g, ollamaPlugin)
			if err != nil {
				logger.Error(err, "define ollama models error")
				continue
			}
			modelNames = append(modelNames, ollamaModelNames...)
		case p.Deepseek != nil:
			dsModelNames, err := deepseekPlugin.RegisterModels(ctx, g)
			if err != nil {
				logger.Error(err, "define deepseek models error")
				continue
			}
			modelNames = append(modelNames, dsModelNames...)
		}
	}

	return g, modelNames
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
