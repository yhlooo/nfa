package dualphase

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/nfa/pkg/ctxutil"
)

// OptimizationInput 优化输入
type OptimizationInput struct {
	// 待优化 Prompt
	Prompt string `json:"prompt"`
	// 待优化句子
	Sentence string `json:"sentence"`
	// 当前失败案例
	FailedCases []FailedCase `json:"failedCases"`
}

// OptimizationOutput 优化输出
type OptimizationOutput struct {
	// 优化后的句子
	NewSentence string `json:"newSentence"`
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
		if m, ok := ctxutil.ModelsFromContext(ctx); ok {
			opts = append(opts, ai.WithModelName(m.GetPrimary()))
		}
		if handleStream := ctxutil.HandleStreamFnFromContext(ctx); handleStream != nil {
			opts = append(opts, ai.WithStreaming(handleStream))
			_ = handleStream(ctx, &ai.ModelResponseChunk{
				Content: []*ai.Part{ai.NewTextPart(fmt.Sprintf("Optimize sentence %q ...", in.Sentence))},
				Role:    ai.RoleUser,
			})
		}

		resp, err := genkit.Generate(ctx, g, opts...)
		if err != nil {
			return OptimizationOutput{}, err
		}

		return OptimizationOutput{
			NewSentence: strings.TrimSpace(resp.Text()),
		}, nil
	}
}

// OptimizationPromptTpl 优化元 Prompt 模版
var OptimizationPromptTpl = template.Must(template.New("OptimizationPrompt").
	Parse(`I'm trying to write a zero-shot prompt which consists of four parts.
My current prompt is:
` + "```" + `
{{ .Prompt }}
` + "```" + `

But it gets the following outputs that fail to match the expected outputs:
{{ range .FailedCases -}}
Input: {{ .Input }}
Output: {{ .Actual }}
Expected: {{ .Expected }}

{{ end -}}

The sentence I want to revise is:
` + "`" + `{{ .Sentence }}` + "`" + `

Comparing the wrong outputs with their corresponding expected answers under the same input, optimize the above sentence to help AI understand the task more comprehensively and accomplish this task better.
Your response is the sentence ` + "`" + `{{ .Sentence }}` + "`" + ` should be revised to, without any explanation.
`))

// OptimizationPrompt 获取优化元 Prompt
func OptimizationPrompt(in OptimizationInput) (string, error) {
	buf := &bytes.Buffer{}
	if err := OptimizationPromptTpl.Execute(buf, in); err != nil {
		return "", err
	}
	return buf.String(), nil
}
