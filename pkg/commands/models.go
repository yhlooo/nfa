package commands

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"

	"github.com/yhlooo/nfa/pkg/agents"
	"github.com/yhlooo/nfa/pkg/configs"
	"github.com/yhlooo/nfa/pkg/i18n"
	"github.com/yhlooo/nfa/pkg/models"
)

// newModelsCommand 创建 models 自命令
func newModelsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models",
		Short: i18n.T(MsgCmdShortDescModels),
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
	t := tablewriter.NewTable(os.Stdout, tablewriter.WithRendition(tw.Rendition{
		Borders: tw.BorderNone,
		Settings: tw.Settings{
			Separators: tw.Separators{BetweenColumns: tw.Off},
		},
	}))
	defer func() { _ = t.Close() }()

	for _, model := range modelList {
		var tags []string
		if model.Reasoning {
			tags = append(tags, i18n.TContext(ctx, MsgReasoningTag))
		}
		if model.Vision {
			tags = append(tags, i18n.TContext(ctx, MsgVisionTag))
		}

		ctxSize := strconv.FormatInt(model.ContextWindow/1000, 10) + "K"
		if model.ContextWindow < 1000 {
			ctxSize = strconv.FormatInt(model.ContextWindow, 10)
		}
		tags = append(tags, ctxSize)

		_ = t.Append([]string{model.Name, strings.Join(tags, ", "), model.Description})
	}

	return t.Render()
}
