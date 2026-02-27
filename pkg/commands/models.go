package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

	"golang.org/x/term"

	"github.com/yhlooo/nfa/pkg/agents"
	"github.com/yhlooo/nfa/pkg/configs"
	"github.com/yhlooo/nfa/pkg/models"
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

// ModelsListOptions models list 命令选项
type ModelsListOptions struct {
	Provider   string // 按提供商过滤
	Capability string // 按能力过滤 (reasoning|vision)
	Format     string // 输出格式 (text|json)
}

// 列宽度常量
const (
	col1Width    = 35 // 模型名称
	col2Width    = 30 // 括号内容
	col3MaxWidth = 50 // 描述最大宽度
)

// newModelsListCommand 创建 models list 子命令
func newModelsListCommand() *cobra.Command {
	opts := ModelsListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available models",
		Long: `List available models with detailed information.

Display all configured models with their capabilities, context window size,
and descriptions. Supports filtering by provider and capability.`,
		Example: `  nfa models list                              # List all models
  nfa models list --provider deepseek          # Only DeepSeek models
  nfa models list --capability reasoning       # Only reasoning models
  nfa models list --format json                # JSON output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runModelsList(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Provider, "provider", "p", "", "Filter by provider name (e.g., deepseek, zhipu, aliyun)")
	cmd.Flags().StringVarP(&opts.Capability, "capability", "c", "", "Filter by capability (reasoning|vision)")
	cmd.Flags().StringVarP(&opts.Format, "format", "f", "text", "Output format (text|json)")

	return cmd
}

// runModelsList 执行 models list 命令
func runModelsList(cmd *cobra.Command, opts ModelsListOptions) error {
	ctx := cmd.Context()
	cfg := configs.ConfigFromContext(ctx)
	logger := logr.FromContextOrDiscard(ctx)

	// 创建 Agent 并获取可用模型
	agent := agents.NewNFA(agents.Options{
		Logger:         logger,
		ModelProviders: cfg.ModelProviders,
	})
	agent.InitGenkit(ctx)

	availableModels := agent.AvailableModels()

	// 过滤模型
	filtered, err := filterModels(availableModels, opts)
	if err != nil {
		return err
	}

	// 输出
	output := cmd.OutOrStdout()
	if opts.Format == "json" {
		return outputJSONModels(filtered, output)
	}
	return outputTextModels(filtered, output)
}

// filterModels 过滤模型列表
func filterModels(modelList []models.ModelConfig, opts ModelsListOptions) ([]models.ModelConfig, error) {
	var result []models.ModelConfig

	for _, model := range modelList {
		// 提供商过滤
		if opts.Provider != "" {
			provider := extractProvider(model.Name)
			if !strings.EqualFold(provider, opts.Provider) {
				continue
			}
		}

		// 能力过滤
		if opts.Capability != "" {
			switch strings.ToLower(opts.Capability) {
			case "reasoning":
				if !model.Reasoning {
					continue
				}
			case "vision":
				if !model.Vision {
					continue
				}
			default:
				return nil, fmt.Errorf("invalid capability: %s. Valid options: reasoning, vision", opts.Capability)
			}
		}

		result = append(result, model)
	}

	return result, nil
}

// extractProvider 从模型名称提取提供商
func extractProvider(modelName string) string {
	parts := strings.SplitN(modelName, "/", 2)
	if len(parts) >= 1 {
		return parts[0]
	}
	return ""
}

// formatContextWindow 格式化上下文窗口大小
func formatContextWindow(bytes int64) string {
	if bytes >= 1024 {
		return fmt.Sprintf("%dK", bytes/1024)
	}
	return fmt.Sprintf("%d", bytes)
}

// truncateByWidth 按显示宽度截断字符串
func truncateByWidth(s string, max int) string {
	if s == "" {
		return ""
	}

	currentWidth := 0
	runes := []rune(s)
	for i, r := range runes {
		w := runewidth.RuneWidth(r)
		if currentWidth+w > max {
			// 添加 "..." 时需要预留 3 个宽度
			if currentWidth+3 > max {
				return string(runes[:i])
			}
			return string(runes[:i]) + "..."
		}
		currentWidth += w
	}
	return s
}

// buildCapabilityTag 构建括号内容字符串
func buildCapabilityTag(model models.ModelConfig) string {
	var parts []string

	// 添加能力标签
	if model.Reasoning {
		parts = append(parts, "Reasoning")
	}
	if model.Vision {
		parts = append(parts, "Vision")
	}

	// 添加上下文大小
	contextSize := formatContextWindow(model.ContextWindow)
	parts = append(parts, contextSize)

	// 用括号包裹，用 ", " 分隔
	return "(" + strings.Join(parts, ", ") + ")"
}

// outputStyles 输出样式
type outputStyles struct {
	capabilityTag termenv.Style // 括号及内容弱化样式
}

// newOutputStyles 创建输出样式
func newOutputStyles() *outputStyles {
	output := termenv.NewOutput(os.Stdout)

	return &outputStyles{
		capabilityTag: output.String().Faint(), // 弱化/淡色
	}
}

// renderModelLine 渲染单个模型的文本行
func renderModelLine(model models.ModelConfig, styles *outputStyles) string {
	// 列 1: 模型名称（左对齐）
	name := model.Name
	nameWidth := runewidth.StringWidth(name)
	if nameWidth > col1Width {
		name = truncateByWidth(name, col1Width)
		nameWidth = col1Width
	}
	padding := col1Width - nameWidth
	if padding < 0 {
		padding = 0
	}
	col1 := name + strings.Repeat(" ", padding)

	// 列 2: 括号内容（左对齐）
	tag := buildCapabilityTag(model)
	tagWidth := runewidth.StringWidth(tag)
	padding = col2Width - tagWidth
	if padding < 0 {
		padding = 0
	}
	col2 := tag + strings.Repeat(" ", padding)

	// 列 3: 描述（左对齐，截断）
	desc := model.Description
	if desc == "" {
		desc = "-"
	}
	desc = truncateByWidth(desc, col3MaxWidth)

	return col1 + col2 + desc
}

// outputTextModels 输出人性化文本格式
func outputTextModels(modelList []models.ModelConfig, w io.Writer) error {
	if len(modelList) == 0 {
		fmt.Fprintln(w, "No models found matching the criteria")
		return nil
	}

	// 检测是否应该使用颜色
	useColor := shouldUseColor()

	var styles *outputStyles
	if useColor {
		styles = newOutputStyles()
	}

	for _, model := range modelList {
		line := renderModelLine(model, styles)
		fmt.Fprintln(w, line)
	}

	return nil
}

// shouldUseColor 检测是否应该使用颜色
func shouldUseColor() bool {
	// 检查 NO_COLOR 环境变量
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		return false
	}

	// 检查输出是否为终端
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// ModelJSON JSON 输出的模型结构
type ModelJSON struct {
	Name            string            `json:"name"`
	Provider        string            `json:"provider"`
	Description     string            `json:"description"`
	Capabilities    ModelCapabilities `json:"capabilities"`
	ContextWindow   int64             `json:"contextWindow"`
	MaxOutputTokens int64             `json:"maxOutputTokens,omitempty"`
	Cost            *models.ModelCost `json:"cost,omitempty"`
}

// ModelCapabilities 模型能力
type ModelCapabilities struct {
	Reasoning bool `json:"reasoning"`
	Vision    bool `json:"vision"`
}

// ModelsJSONOutput JSON 输出结构
type ModelsJSONOutput struct {
	Models []ModelJSON `json:"models"`
}

// outputJSONModels 输出 JSON 格式
func outputJSONModels(modelList []models.ModelConfig, w io.Writer) error {
	if len(modelList) == 0 {
		fmt.Fprintln(w, `{"models":[]}`)
		return nil
	}

	output := ModelsJSONOutput{
		Models: make([]ModelJSON, 0, len(modelList)),
	}

	for _, model := range modelList {
		modelJSON := ModelJSON{
			Name:        model.Name,
			Provider:    extractProvider(model.Name),
			Description: model.Description,
			Capabilities: ModelCapabilities{
				Reasoning: model.Reasoning,
				Vision:    model.Vision,
			},
			ContextWindow:   model.ContextWindow,
			MaxOutputTokens: model.MaxOutputTokens,
		}

		// 添加价格信息（如果有）
		if model.Cost.Input > 0 || model.Cost.Output > 0 {
			modelJSON.Cost = &model.Cost
		}

		output.Models = append(output.Models, modelJSON)
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
