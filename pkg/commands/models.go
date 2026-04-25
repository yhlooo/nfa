package commands

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/yhlooo/nfa/pkg/agents"
	"github.com/yhlooo/nfa/pkg/configs"
	"github.com/yhlooo/nfa/pkg/i18n"
	"github.com/yhlooo/nfa/pkg/models"
)

// newModelsCommand 创建 models 子命令
func newModelsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "models",
		Aliases: []string{"model"},
		Short:   i18n.T(MsgCmdShortDescModels),
	}

	cmd.AddCommand(
		newModelsListCommand(),
		newModelsAddCommand(),
	)

	return cmd
}

// newModelsListCommand 创建 models list 子命令
func newModelsListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   i18n.T(MsgCmdShortDescModelsList),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runModelsList(cmd.Context())
		},
	}

	return cmd
}

// runModelsList 执行 models list 命令
func runModelsList(ctx context.Context) error {
	cfg := configs.ConfigFromContext(ctx)
	logger := logr.FromContextOrDiscard(ctx)

	// 创建 Agent 并获取可用模型
	agent := agents.NewNFA(agents.Options{
		Logger:         logger,
		ModelProviders: cfg.ModelProviders,
	})
	agent.InitGenkit(ctx)

	// 输出
	return outputModelList(ctx, agent.AvailableModels())
}

// outputModelList 输出模型列表
func outputModelList(ctx context.Context, modelList []models.ModelConfig) error {
	t := tablewriter.NewTable(os.Stdout,
		tablewriter.WithHeader([]string{
			i18n.TContext(ctx, MsgModelNameTag),
			i18n.TContext(ctx, MsgReasoningTag),
			i18n.TContext(ctx, MsgVisionTag),
			i18n.TContext(ctx, MsgModelContextTag),
			i18n.TContext(ctx, MsgScoreTag),
		}),
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.BorderNone,
			Settings: tw.Settings{
				Separators: tw.Separators{BetweenColumns: tw.Off},
			},
		}),
		tablewriter.WithAlignment([]tw.Align{
			tw.AlignLeft, tw.AlignCenter, tw.AlignCenter,
			tw.AlignRight, tw.AlignLeft,
		}),
	)
	defer func() { _ = t.Close() }()

	for _, model := range modelList {
		reasoning := "❌"
		if model.Reasoning {
			reasoning = "✅"
		}
		vision := "❌"
		if model.Vision {
			vision = "✅"
		}

		ctxSize := strconv.FormatInt(model.ContextWindow/1000, 10) + "K"
		if model.ContextWindow < 1000 {
			ctxSize = strconv.FormatInt(model.ContextWindow, 10)
		}

		_ = t.Append([]string{model.Name, reasoning, vision, ctxSize, scoreToStars(model.Score)})
	}

	return t.Render()
}

// scoreToStars 将 0-10 的评分转换为星标展示
func scoreToStars(score int) string {
	if score < 0 {
		score = 0
	}
	if score > 10 {
		score = 10
	}
	if score == 0 {
		return ""
	}
	stars := make([]string, 0, 5)
	for range score / 2 {
		stars = append(stars, "⭐️")
	}
	if score%2 == 1 {
		stars = append(stars, "✨")
	}
	return strings.Join(stars, " ")
}

// NewModelsAddOptions 创建默认 ModelsAddOptions
func NewModelsAddOptions() ModelsAddOptions {
	return ModelsAddOptions{
		APIKey:  "",
		BaseURL: "",
		Name:    "",
	}
}

// ModelsAddOptions models add 子命令选项
type ModelsAddOptions struct {
	APIKey  string
	BaseURL string
	Name    string
}

// AddPFlags 将选项绑定到命令行参数
func (opts *ModelsAddOptions) AddPFlags(fs *pflag.FlagSet) {
	fs.StringVar(&opts.APIKey, "api-key", opts.APIKey, i18n.T(MsgModelsAddOptAPIKeyDesc))
	fs.StringVar(&opts.BaseURL, "base-url", opts.BaseURL, i18n.T(MsgModelsAddOptBaseURLDesc))
	fs.StringVar(&opts.Name, "name", opts.Name, i18n.T(MsgModelsAddOptNameDesc))
}

// newModelsAddCommand 创建 models add 子命令
func newModelsAddCommand() *cobra.Command {
	opts := NewModelsAddOptions()
	cmd := &cobra.Command{
		Use:   "add <provider>",
		Short: i18n.T(MsgCmdShortDescModelsAdd),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runModelsAdd(cmd.Context(), args[0], opts)
		},
	}

	opts.AddPFlags(cmd.Flags())

	return cmd
}

// runModelsAdd 执行 models add 命令
func runModelsAdd(ctx context.Context, providerName string, opts ModelsAddOptions) error {
	cfg := configs.ConfigFromContext(ctx)
	cfgPath := configs.ConfigPathFromContext(ctx)

	// 查找供应商是否已经配置
	existingIdx := slices.IndexFunc(cfg.ModelProviders, func(p models.ModelProvider) bool {
		switch providerName {
		case "ollama":
			return p.Ollama != nil
		case "openai-compatible":
			return p.OpenAICompatible != nil
		case "openrouter":
			return p.OpenRouter != nil
		case "opencode":
			return p.OpenCode != nil
		case "opencode-go":
			return p.OpenCodeGo != nil
		case "deepseek":
			return p.Deepseek != nil
		case "qwen":
			return p.Qwen != nil
		case "moonshotai":
			return p.MoonshotAI != nil
		case "z-ai":
			return p.ZAI != nil
		case "minimax":
			return p.Minimax != nil
		}
		return false
	})

	provider := buildModelProvider(providerName, opts)
	if existingIdx >= 0 {
		// 更新已有供应商配置
		cfg.ModelProviders[existingIdx] = provider
	} else {
		// 添加供应商配置
		cfg.ModelProviders = append(cfg.ModelProviders, provider)
	}

	// 保存配置
	if err := configs.SaveConfig(cfgPath, cfg); err != nil {
		return fmt.Errorf("save config error: %w", err)
	}

	return nil
}

// buildModelProvider 根据供应商名构建对应的 ModelProvider
func buildModelProvider(key string, opts ModelsAddOptions) models.ModelProvider {
	switch key {
	case "ollama":
		return models.ModelProvider{
			Ollama: &models.OllamaOptions{
				BaseURL: opts.BaseURL,
			},
		}
	case "openai-compatible":
		return models.ModelProvider{
			OpenAICompatible: &models.OpenAICompatibleOptions{
				Name:    opts.Name,
				BaseURL: opts.BaseURL,
				APIKey:  opts.APIKey,
			},
		}
	case "openrouter":
		return models.ModelProvider{
			OpenRouter: &models.OpenAICompatibleOptions{Name: opts.Name, BaseURL: opts.BaseURL, APIKey: opts.APIKey},
		}
	case "opencode":
		return models.ModelProvider{
			OpenCode: &models.OpenAICompatibleOptions{Name: opts.Name, BaseURL: opts.BaseURL, APIKey: opts.APIKey},
		}
	case "opencode-go":
		return models.ModelProvider{
			OpenCodeGo: &models.OpenAICompatibleOptions{Name: opts.Name, BaseURL: opts.BaseURL, APIKey: opts.APIKey},
		}
	case "deepseek":
		return models.ModelProvider{
			Deepseek: &models.OpenAICompatibleOptions{Name: opts.Name, BaseURL: opts.BaseURL, APIKey: opts.APIKey},
		}
	case "qwen":
		return models.ModelProvider{
			Qwen: &models.OpenAICompatibleOptions{Name: opts.Name, BaseURL: opts.BaseURL, APIKey: opts.APIKey},
		}
	case "moonshotai":
		return models.ModelProvider{
			MoonshotAI: &models.OpenAICompatibleOptions{Name: opts.Name, BaseURL: opts.BaseURL, APIKey: opts.APIKey},
		}
	case "z-ai":
		return models.ModelProvider{
			ZAI: &models.OpenAICompatibleOptions{Name: opts.Name, BaseURL: opts.BaseURL, APIKey: opts.APIKey},
		}
	case "minimax":
		return models.ModelProvider{
			Minimax: &models.OpenAICompatibleOptions{Name: opts.Name, BaseURL: opts.BaseURL, APIKey: opts.APIKey},
		}
	}
	return models.ModelProvider{}
}
