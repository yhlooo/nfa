package agents

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-logr/logr"

	"github.com/yhlooo/nfa/pkg/agents/flows"
	"github.com/yhlooo/nfa/pkg/models"
	"github.com/yhlooo/nfa/pkg/tools/fs"
	"github.com/yhlooo/nfa/pkg/tools/webbrowse"
)

const ChatFlowName = "Chat"

// InitGenkit 初始化 genkit
func (a *NFAAgent) InitGenkit(ctx context.Context) {
	ctx = logr.NewContext(ctx, a.logger)

	if a.g != nil {
		return
	}

	a.g, a.availableModels = NewGenkitWithModels(ctx, a.opts.ModelProviders, a.opts.DefaultModels)
	if a.opts.DefaultModels.Primary == "" && len(a.availableModels) > 0 {
		a.opts.DefaultModels.Primary = a.availableModels[0].Name
	}
	for _, m := range a.availableModels {
		a.logger.Info(fmt.Sprintf("registered model: %s", m.Name))
	}

	// 注册工具
	if a.opts.DataProviders.AlphaVantage != nil {
		alphaVantageTools, err := a.opts.DataProviders.AlphaVantage.RegisterTools(ctx, a.g)
		if err != nil {
			a.logger.Error(err, "register alpha vantage tools error")
		} else {
			a.availableTools = append(a.availableTools, alphaVantageTools...)
		}
	}
	if a.opts.DataProviders.TencentCloudWSA != nil {
		searchTool, err := a.opts.DataProviders.TencentCloudWSA.RegisterTool(ctx, a.g)
		if err != nil {
			a.logger.Error(err, "register tencent cloud wsa search tool error")
		} else {
			a.availableTools = append(a.availableTools, searchTool)
		}
	}
	// 网页浏览工具
	wb := webbrowse.NewWebBrowser()
	a.availableTools = append(a.availableTools, wb.RegisterTools(a.g)...)

	// 文件读取工具
	a.availableTools = append(a.availableTools, fs.DefineReadTool(a.g))

	// 注册 Skill 工具
	a.availableTools = append(a.availableTools, a.skillLoader.DefineSkillTool(a.g))

	for _, t := range a.availableTools {
		a.logger.Info(fmt.Sprintf("registered tool: %s", t.Name()))
	}

	// 注册 flows
	a.chatFlow = flows.DefineSimpleChatFlow(a.g, ChatFlowName,
		ai.WithSystemFn(AnalystSystemPrompt(a.skillLoader)),
		ai.WithTools(a.availableTools...),
	)
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
		modelRegisters []models.ModelRegister
		plugins        []api.Plugin
		modelConfigs   []models.ModelConfig
	)
	for _, p := range providers {
		modelRegister := p.Register()
		if modelRegister == nil {
			continue
		}
		plugins = append(plugins, modelRegister.GenkitPlugin())
		modelRegisters = append(modelRegisters, modelRegister)
	}

	genkitOpts := []genkit.GenkitOption{
		genkit.WithPlugins(plugins...),
	}
	if defaultModels.GetPrimary() != "" {
		genkitOpts = append(genkitOpts, genkit.WithDefaultModel(defaultModels.GetPrimary()))
	}
	g := genkit.Init(ctx, genkitOpts...)

	// 注册模型
	for i, reg := range modelRegisters {
		registeredModels, err := reg.RegisterModels(ctx, g)
		if err != nil {
			logger.Error(err, fmt.Sprintf("register model for provider %d error", i))
			continue
		}
		modelConfigs = append(modelConfigs, registeredModels...)
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
