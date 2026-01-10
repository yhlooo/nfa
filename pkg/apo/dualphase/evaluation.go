package dualphase

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

// BatchEvaluationInput 批量评估输入
type BatchEvaluationInput struct {
	// 待评估 Prompt
	Prompt string `json:"prompt"`
	// 验证数据
	ValidationData []InputOutputPair `json:"validationData"`
}

// BatchEvaluationOutput 批量评估输出
type BatchEvaluationOutput struct {
	// 正确数目
	Correct int `json:"correct"`
	// 错误数
	Wrong int `json:"wrong"`
	// 失败案例
	FailedCases []FailedCase `json:"failedCases,omitempty"`
}

// FailedCase 失败案例
type FailedCase struct {
	// 输入
	Input string `json:"input"`
	// 预期输出
	Expected string `json:"expected"`
	// 实际输出
	Actual string `json:"actual"`
}

// DefineBatchEvaluationFlow 定义批量评估流程
func DefineBatchEvaluationFlow(g *genkit.Genkit) *core.Flow[BatchEvaluationInput, BatchEvaluationOutput, struct{}] {
	return genkit.DefineFlow(g, "", BatchEvaluationFlow(g))
}

// BatchEvaluationFlow 批量评估流程
func BatchEvaluationFlow(g *genkit.Genkit) core.Func[BatchEvaluationInput, BatchEvaluationOutput] {
	return func(ctx context.Context, in BatchEvaluationInput) (BatchEvaluationOutput, error) {
		logger := logr.FromContextOrDiscard(ctx)

		prompt, err := BatchEvaluationPrompt(in)
		if err != nil {
			return BatchEvaluationOutput{}, fmt.Errorf("make prompt error: %w", err)
		}
		handleStream := ctxutil.HandleStreamFnFromContext(ctx)
		if handleStream != nil {
			_ = handleStream(ctx, &ai.ModelResponseChunk{
				Content: []*ai.Part{ai.NewTextPart("Write outputs for inputs ...")},
				Role:    ai.RoleUser,
			})
		}
		messages := []*ai.Message{ai.NewUserTextMessage(prompt)}

		inputLen := len(in.ValidationData)

		// Prompt+Input 生成输出
		var genErr error
		var results *[]string
		for i := 0; i < 3; i++ {
			modelOutput := ""
			opts := []ai.GenerateOption{
				ai.WithMessages(messages...),
				ai.WithStreaming(func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
					if chunk.Role == ai.RoleModel {
						modelOutput += chunk.Text()
					}
					if handleStream != nil {
						return handleStream(ctx, chunk)
					}
					return nil
				}),
			}
			if modelName, ok := ctxutil.ModelNameFromContext(ctx); ok {
				opts = append(opts, ai.WithModelName(modelName))
			}

			results, _, genErr = genkit.GenerateData[[]string](ctx, g, opts...)
			outputLen := 0
			if results != nil {
				outputLen = len(*results)
			}
			if genErr == nil && outputLen != inputLen {
				// 输入输出项数目不匹配
				genErr = fmt.Errorf(
					"the number of output items does not match the number of input items, output: %d (expected %d)",
					outputLen, inputLen,
				)
			}
			if genErr != nil {
				logger.Error(err, fmt.Sprintf("generate evaluation output error (retries: %d)", i))
				fixPrompt := fmt.Sprintf("输出错误，根据错误提示修复后重新输出： %s", genErr.Error())
				if handleStream != nil {
					_ = handleStream(ctx, &ai.ModelResponseChunk{
						Content: []*ai.Part{ai.NewTextPart(fixPrompt)},
						Role:    ai.RoleUser,
					})
				}
				messages = append(messages, ai.NewModelTextMessage(modelOutput), ai.NewUserTextMessage(fixPrompt))
				continue
			}

			genErr = nil
			break
		}
		if genErr != nil {
			return BatchEvaluationOutput{}, genErr
		}

		// 校验输出
		correct := 0
		wrong := 0
		var failedCases []FailedCase
		for i, item := range in.ValidationData {
			ret := ""
			if i < len(*results) {
				ret = (*results)[i]
			}
			if ret == item.Output {
				correct++
			} else {
				wrong++
				failedCases = append(failedCases, FailedCase{
					Input:    item.Input,
					Expected: item.Output,
					Actual:   ret,
				})
			}
		}

		return BatchEvaluationOutput{
			Correct:     correct,
			Wrong:       wrong,
			FailedCases: failedCases,
		}, nil
	}
}

// BatchEvaluationPromptTpl 批量评估 Prompt 模版
var BatchEvaluationPromptTpl = template.Must(template.New("BatchEvaluationPrompt").
	Parse(`根据以下指令，对每个输入数据给出正确的输出

###指令###
` + "```" + `
{{ .Prompt }}
` + "```" + `

###输出格式###
以 JSON 字符串列表格式输出，每一项是一个输出，按顺序与输入对应

###输入###
{{- range $i, $v := .ValidationData }}
input {{ $i }}:
` + "```" + `
{{ $v.Input }}
` + "```" + `

{{- end }}
`))

// BatchEvaluationPrompt 获取批量评估 Prompt
func BatchEvaluationPrompt(in BatchEvaluationInput) (string, error) {
	buf := &bytes.Buffer{}
	if err := BatchEvaluationPromptTpl.Execute(buf, in); err != nil {
		return "", err
	}
	return buf.String(), nil
}
