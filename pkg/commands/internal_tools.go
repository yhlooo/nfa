package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/yhlooo/nfa/pkg/agents"
	"github.com/yhlooo/nfa/pkg/apo/dualphase"
	"github.com/yhlooo/nfa/pkg/ctxutil"
)

// newInternalToolsCommand 创建 internal-tools 子命令
func newInternalToolsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "internal-tools",
		Short:  "Internal tools",
		Hidden: true,
	}

	cmd.AddCommand(
		newInternalToolsAPOCommand(),
	)

	return cmd
}

// APOOptions internal-tools apo 子命令选项
type APOOptions struct {
	OptionsFile string
}

// NewAPOOptions 创建 internal-tools apo 子命令选项
func NewAPOOptions() APOOptions {
	return APOOptions{
		OptionsFile: "",
	}
}

// AddPFlags 将选项绑定到命令行参数
func (o *APOOptions) AddPFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.OptionsFile, "opts", o.OptionsFile, "Path of options file")
}

// newInternalToolsAPOCommand 创建 internal-tools apo 子命令
func newInternalToolsAPOCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apo",
		Short: "Automatic Prompt Optimization",
	}

	apoOpts := NewAPOOptions()

	cmd.AddCommand(
		&cobra.Command{
			Use:   "dual-phase",
			Short: "Dual-Phase Accelerated Prompt Optimization (See https://arxiv.org/abs/2406.13443)",
			RunE: func(cmd *cobra.Command, args []string) error {
				ctx := cmd.Context()
				cfg := ConfigFromContext(ctx)

				opts := dualphase.Options{}
				if apoOpts.OptionsFile == "" {
					fmt.Print("Prompt: ")
					if _, err := fmt.Scan(&opts.Initialization.PreviousP0); err != nil {
						return fmt.Errorf("get previous p0 error: %w", err)
					}
					opts.Initialization.GenerateTrainingDataPairs = 3
				} else {
					optsContent, err := os.ReadFile(apoOpts.OptionsFile)
					if err != nil {
						return fmt.Errorf("read options file %q error: %w", apoOpts.OptionsFile, err)
					}
					if err := json.Unmarshal(optsContent, &opts); err != nil {
						return fmt.Errorf("unmarshal options file %q from json error: %w", apoOpts.OptionsFile, err)
					}
				}

				g, modelNames := agents.NewGenkitWithModels(ctx, cfg.ModelProviders)
				if len(modelNames) == 0 {
					return fmt.Errorf("no available model found")
				}

				model := cfg.DefaultModel
				if model == "" {
					model = modelNames[0]
				}
				ctx = ctxutil.ContextWithModelName(ctx, model)
				ctx = ctxutil.ContextWithHandleStreamFn(ctx, handleModelStream(os.Stdout))

				optimizer := dualphase.NewOptimizer(g, opts)

				fmt.Println("================= Initialization =================")
				curPrompt, curAccuracy, err := optimizer.Initialize(ctx)
				if err != nil {
					return fmt.Errorf("initialization error: %w", err)
				}
				fmt.Println()

				defer func() {
					raw, _ := json.MarshalIndent(curPrompt, "", "  ")
					fmt.Println("Current prompt:")
					fmt.Println(string(raw))
				}()

				fmt.Println("----------------------- P0 -----------------------")
				fmt.Printf("Accuracy: %.4f\n", curAccuracy)
				fmt.Println(curPrompt.WithWeightColors())
				fmt.Println("--------------------------------------------------")

				return nil
			},
		},
	)

	apoOpts.AddPFlags(cmd.PersistentFlags())

	return cmd
}

// handleModelStream 处理模型流式输出
func handleModelStream(w io.Writer) ai.ModelStreamCallback {
	curPartType := ""
	return func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		prefix := ""
		indent := 0
		switch chunk.Role {
		case ai.RoleSystem:
			prefix = "📖 \033[34m"
			indent = 2
		case ai.RoleUser:
			prefix = "☝️ \033[32m"
			indent = 2
		case ai.RoleModel:
		}

		for _, part := range chunk.Content {
			switch {
			case part.IsReasoning():
				if curPartType != string(chunk.Role)+"-reasoning" {
					if curPartType != "" {
						_, _ = fmt.Fprintln(w, "\033[0m")
					}
					_, _ = fmt.Fprint(w, "🧠  \033[2m")
					curPartType = string(chunk.Role) + "-reasoning"
				}
				_, _ = fmt.Fprint(w, strings.ReplaceAll(part.Text, "\n", "\n  "))

			case part.IsText() || part.IsData():
				if curPartType != string(chunk.Role)+"-text" {
					if curPartType != "" {
						_, _ = fmt.Fprintln(w, "\033[0m")
					}
					_, _ = fmt.Fprint(w, prefix)
					curPartType = string(chunk.Role) + "-text"
				}
				_, _ = fmt.Fprint(w, strings.ReplaceAll(part.Text, "\n", "\n"+strings.Repeat(" ", indent)))

			case part.IsToolRequest() && part.ToolRequest != nil:
				if curPartType != string(chunk.Role)+"-tool" {
					if curPartType != "" {
						_, _ = fmt.Fprintln(w, "\033[0m")
					}
					_, _ = fmt.Fprint(w, prefix)
					curPartType = string(chunk.Role) + "-tool"
				}

				inputRaw, _ := json.Marshal(part.ToolRequest.Input)
				_, _ = fmt.Fprintf(
					w, "🔧 \033[34mToolCall: %s \033[2m%s\033[0m\n",
					part.ToolRequest.Name, string(inputRaw),
				)

			case part.IsToolResponse() && part.ToolResponse != nil:
				if curPartType != string(chunk.Role)+"-tool" {
					if curPartType != "" {
						_, _ = fmt.Fprintln(w, "\033[0m")
					}
					_, _ = fmt.Fprint(w, prefix)
					curPartType = string(chunk.Role) + "-tool"
				}

				outputRaw, _ := json.Marshal(part.ToolResponse.Output)
				_, _ = fmt.Fprintf(w, "           \u001B[34mcompleted \u001B[2m%s\033[0m\n", string(outputRaw))
			}
		}

		return nil
	}
}
