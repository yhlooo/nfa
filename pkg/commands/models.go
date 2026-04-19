package commands

import (
	"context"
	"os"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"

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
			"Name",
			i18n.TContext(ctx, MsgReasoningTag),
			i18n.TContext(ctx, MsgVisionTag),
			"Context",
		}),
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.BorderNone,
			Settings: tw.Settings{
				Separators: tw.Separators{BetweenColumns: tw.Off},
			},
		}),
		tablewriter.WithAlignment([]tw.Align{tw.AlignLeft, tw.AlignCenter, tw.AlignCenter, tw.AlignRight}),
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

		_ = t.Append([]string{model.Name, reasoning, vision, ctxSize})
	}

	return t.Render()
}
