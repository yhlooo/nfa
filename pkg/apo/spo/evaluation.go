package spo

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-logr/logr"

	"github.com/yhlooo/nfa/pkg/ctxutil"
)

// EvaluationInput 评估输入
type EvaluationInput struct {
	// 额外要求
	Requirements string `json:"requirements,omitempty"`
	// 回答 A
	AnswerA string `json:"answerA"`
	// 回答 B
	AnswerB string `json:"answerB"`
}

// EvaluationOutput 评估输出
type EvaluationOutput struct {
	// 分析
	Analysis string `json:"analysis"`
	// 选择
	Choice string `json:"choose"`
}

// DefineEvaluationFlow 定义评估流程
func DefineEvaluationFlow(g *genkit.Genkit) *core.Flow[EvaluationInput, EvaluationOutput, struct{}] {
	return genkit.DefineFlow(g, "Evaluation", EvaluationFlow(g))
}

// EvaluationFlow 评估流程
func EvaluationFlow(g *genkit.Genkit) core.Func[EvaluationInput, EvaluationOutput] {
	return func(ctx context.Context, in EvaluationInput) (EvaluationOutput, error) {
		logger := logr.FromContextOrDiscard(ctx)

		prompt, err := EvaluationPrompt(in)
		if err != nil {
			return EvaluationOutput{}, fmt.Errorf("make prompt error: %w", err)
		}

		opts := []ai.GenerateOption{
			ai.WithPrompt(prompt),
		}
		if m, ok := ctxutil.ModelsFromContext(ctx); ok {
			opts = append(opts, ai.WithModelName(m.GetPrimary()))
		}

		for i := 0; i < 3; i++ {
			var ret *EvaluationOutput
			ret, _, err = genkit.GenerateData[EvaluationOutput](ctx, g, opts...)
			if err != nil {
				logger.Error(err, "generate evaluation result error")
				continue
			}
			if ret.Choice != "A" && ret.Choice != "B" {
				logger.Info(fmt.Sprintf("invalid evaluation result: %s", ret.Choice))
				continue
			}
			return *ret, nil
		}

		return EvaluationOutput{}, err
	}
}

// EvaluationPromptTpl 评估 Prompt 模版
var EvaluationPromptTpl = template.Must(template.New("EvaluationPrompt").
	Parse(`Based on the original requirements, evaluate the two responses, A and B, and determine which one better meets the requirements. If a reference requirement is provided, strictly follow the format/content of the reference requirement.

{{- if .Requirements }}
# Requirement
` + "```" + `
{{ .Requirements }}
` + "```" + `
{{- end }}

# A
` + "```" + `
{{ .AnswerA }}
` + "```" + `

# B
` + "```" + `
{{ .AnswerB }}
` + "```" + `

Provide your analysis and the choice you believe is better, using JSON to encapsulate your response.

` + "```" + `json
{
  "analysis": "Some analysis",
  "choose": "A/B (the better answer in your opinion)"
}
` + "```" + `
`))

// EvaluationPrompt 获取评估 Prompt
func EvaluationPrompt(in EvaluationInput) (string, error) {
	buf := &bytes.Buffer{}
	if err := EvaluationPromptTpl.Execute(buf, in); err != nil {
		return "", err
	}
	return buf.String(), nil
}
