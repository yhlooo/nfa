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
	"github.com/yhlooo/nfa/pkg/agents/models"
	"github.com/yhlooo/nfa/pkg/genkitplugins/deepseek"
)

const ChatFlowName = "Chat"

// InitGenkit 初始化 genkit
func (a *NFAAgent) InitGenkit(ctx context.Context) {
	ctx = logr.NewContext(ctx, a.logger)

	if a.g != nil {
		return
	}

	// 确定插件
	var (
		ollamaPlugin   = &ollama.Ollama{}
		deepseekPlugin = &deepseek.Deepseek{}
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
		case p.OpenAICompatible != nil:
			plugin := p.OpenAICompatible.OpenAICompatiblePlugin()
			// 注册插件后自动注册模型，这里仅获取模型名
			modelNames, err := models.ListOpenAICompatibleModels(ctx, plugin)
			if err != nil {
				a.logger.Error(err, fmt.Sprintf("list %s models error", p.OpenAICompatible.Name))
				continue
			}
			a.availableModels = append(a.availableModels, modelNames...)
			plugins = append(plugins, plugin)
		}
	}

	a.g = genkit.Init(ctx, genkit.WithPlugins(plugins...))

	// 注册模型
	for _, p := range a.modelProviders {
		switch {
		case p.Ollama != nil:
			modelNames, err := p.Ollama.RegisterModels(ctx, a.g, ollamaPlugin)
			if err != nil {
				a.logger.Error(err, "define ollama models error")
				continue
			}
			a.availableModels = append(a.availableModels, modelNames...)
		case p.Deepseek != nil:
			modelNames, err := deepseekPlugin.RegisterModels(ctx, a.g)
			if err != nil {
				a.logger.Error(err, "define deepseek models error")
				continue
			}
			a.availableModels = append(a.availableModels, modelNames...)
		}
	}

	if a.defaultModel == "" && len(a.availableModels) > 0 {
		a.defaultModel = a.availableModels[0]
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
		}
	}

	for _, t := range a.allTools {
		a.logger.Info(fmt.Sprintf("registered tool: %s", t.Name()))
	}

	// 注册 flows
	if a.singleAgent {
		a.mainFlow = flows.DefineSimpleChatFlow(a.g, ChatFlowName, flows.FixedGenerateOptions(
			ai.WithSystemFn(AllAroundAnalystSystemPrompt),
			ai.WithTools(a.allTools...),
		))
	} else {
		mainAgent, subAgents := NewDefaultAgents(
			a.comprehensiveAnalysisTools,
			a.macroeconomicAnalysisTools,
			a.fundamentalAnalysisTools,
			a.technicalAnalysisTools,
		)
		a.mainFlow = flows.DefineMultiAgentsChatFlow(a.g, ChatFlowName, mainAgent, subAgents)
	}
	a.logger.Info(fmt.Sprintf("registered main flow: %s", a.mainFlow.Name()))
	a.summarizeFlow = flows.DefineSummarizeFlow(a.g)
	a.logger.Info(fmt.Sprintf("registered summarize flow: %s", a.summarizeFlow.Name()))
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
