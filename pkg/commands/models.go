package commands

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/yhlooo/nfa/pkg/agents"
	"github.com/yhlooo/nfa/pkg/configs"
)

// newModelsCommand 创建 models 自命令
func newModelsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models",
		Short: "Manage LLMs used by the agent",
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
		Short:   "List available models",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := configs.ConfigFromContext(ctx)
			logger := logr.FromContextOrDiscard(ctx)

			agent := agents.NewNFA(agents.Options{
				Logger:         logger,
				ModelProviders: cfg.ModelProviders,
			})
			agent.InitGenkit(ctx)

			for _, model := range agent.AvailableModels() {
				fmt.Println(model)
			}

			return nil
		},
	}

	return cmd
}
