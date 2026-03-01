package spo

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/nfa/pkg/ctxutil"
)

// ExecutionInput 执行输入
type ExecutionInput struct {
	Prompt    string   `json:"prompt"`
	Questions []string `json:"questions"`
}

// ExecutionOutput 执行输出
type ExecutionOutput struct {
	Answers []string `json:"answers"`
}

// DefineExecutionFlow 定义执行流程
func DefineExecutionFlow(g *genkit.Genkit) *core.Flow[ExecutionInput, ExecutionOutput, struct{}] {
	return genkit.DefineFlow(g, "Execution", ExecutionFlow(g))
}

// ExecutionFlow 执行流程
func ExecutionFlow(g *genkit.Genkit) core.Func[ExecutionInput, ExecutionOutput] {
	return func(ctx context.Context, in ExecutionInput) (ExecutionOutput, error) {
		ret := &ExecutionOutput{}
		for _, q := range in.Questions {
			prompt := fmt.Sprintf(`根据指令给出问题对应的回答

## 指令
`+"```"+`
%s
`+"```"+`

## 问题
`+"```"+`
%s
`+"```"+`
`, in.Prompt, q)
			opts := []ai.GenerateOption{
				ai.WithPrompt(prompt),
			}
			if m, ok := ctxutil.ModelsFromContext(ctx); ok {
				opts = append(opts, ai.WithModelName(m.GetPrimary()))
			}
			if handleStream := ctxutil.HandleStreamFnFromContext(ctx); handleStream != nil {
				opts = append(opts, ai.WithStreaming(handleStream))
			}

			resp, err := genkit.Generate(ctx, g, opts...)
			if err != nil {
				return *ret, err
			}
			ret.Answers = append(ret.Answers, resp.Text())
		}

		return *ret, nil
	}
}
