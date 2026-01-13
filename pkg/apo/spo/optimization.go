package spo

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/nfa/pkg/ctxutil"
)

// OptimizationInput 优化输入
type OptimizationInput struct {
	// 额外要求
	Requirements string `json:"requirements,omitempty"`
	// 当前 Prompt
	Prompt string `json:"prompt"`
	// 当前回答
	Answers []string `json:"answers"`
}

// OptimizationOutput 优化输出
type OptimizationOutput struct {
	// 分析
	Analysis string `json:"analysis"`
	// 修改关键点
	Modification string `json:"modification"`
	// 修改后的 Prompt
	Prompt string `json:"prompt"`
}

// DefineOptimizationFlow 定义优化流程
func DefineOptimizationFlow(g *genkit.Genkit) *core.Flow[OptimizationInput, OptimizationOutput, struct{}] {
	return genkit.DefineFlow(g, "Optimization", OptimizationFlow(g))
}

// OptimizationFlow 优化流程
func OptimizationFlow(g *genkit.Genkit) core.Func[OptimizationInput, OptimizationOutput] {
	return func(ctx context.Context, in OptimizationInput) (OptimizationOutput, error) {
		prompt, err := OptimizationPrompt(in)
		if err != nil {
			return OptimizationOutput{}, fmt.Errorf("make prompt error: %w", err)
		}

		opts := []ai.GenerateOption{
			ai.WithPrompt(prompt),
		}
		if modelName, ok := ctxutil.ModelNameFromContext(ctx); ok {
			opts = append(opts, ai.WithModelName(modelName))
		}

		ret, _, err := genkit.GenerateData[OptimizationOutput](ctx, g, opts...)
		if err != nil {
			return OptimizationOutput{}, err
		}

		return *ret, nil
	}
}

// OptimizationPromptTpl 优化 Prompt 模版
var OptimizationPromptTpl = template.Must(template.New("OptimizationPrompt").
	Parse(`You are building a prompt to address user requirement.Based on the given prompt, please reconstruct and optimize it. You can add, modify, or delete prompts. Please include a single modification in XML tags in your reply. During the optimization, you can incorporate any thinking models.
This is a prompt that performed excellently in a previous iteration. You must make further optimizations and improvements based on this prompt. The modified prompt must differ from the provided example.

{{- if .Requirements }}
requirements:
` + "```" + `
{{ .Requirements }}
` + "```" + `
{{- end }}

reference prompt:
` + "```" + `
{{ .Prompt }}
` + "```" + `

The execution result of this reference prompt is(some cases):
{{- range .Answers }}
` + "```" + `
{{ . }}
` + "```" + `

{{ end }}

Provide your analysis, optimization points, and the complete optimized prompt using the following JSON format:

` + "```" + `json
{
  "analysis": "Analyze what drawbacks exist in the results produced by the reference prompt and how to improve them.",
  "modification": "Summarize the key points for improvement in one sentence",
  "prompt": "Provide the complete optimized prompt"
}
` + "```" + `
`))

// OptimizationPrompt 获取优化 Prompt
func OptimizationPrompt(in OptimizationInput) (string, error) {
	buf := &bytes.Buffer{}
	if err := OptimizationPromptTpl.Execute(buf, in); err != nil {
		return "", err
	}
	return buf.String(), nil
}
