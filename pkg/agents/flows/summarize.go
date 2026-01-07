package flows

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/nfa/pkg/ctxutil"
)

// SummarizeFlowName 总结流名
const SummarizeFlowName = "Summarize"

// SummarizeInput 总结输入
type SummarizeInput struct {
	History []*ai.Message `json:"history,omitempty"`
}

// SummarizeOutput 总结输出
type SummarizeOutput struct {
	Title              string `json:"title"`
	Description        string `json:"description"`
	ProcessOverview    string `json:"processOverview"`
	MethodologySummary string `json:"methodologySummary,omitempty"`
}

type SummarizeFlow = *core.Flow[SummarizeInput, SummarizeOutput, *ai.ModelResponseChunk]

// DefineSummarizeFlow 定义总结工作流
func DefineSummarizeFlow(g *genkit.Genkit) SummarizeFlow {
	return genkit.DefineStreamingFlow(g, SummarizeFlowName,
		func(ctx context.Context, in SummarizeInput, handleStream ai.ModelStreamCallback) (SummarizeOutput, error) {
			opts := []ai.GenerateOption{
				ai.WithMessages(in.History...),
				ai.WithSystem(`你是一个金融领域的客户信息归档员，负责归档与用户的对话记录。

将对话记录按以下几个维度总结并归档输出：
- **title**: 标题。
- **description**: 描述。这次对话的一句话描述。
- **processOverview**: 过程概述。简述对话过程。
- **methodologySummary** (optional) 方法论总结。在这个对话中能提炼出什么什么解决某类问题的方法论。
`),
				ai.WithPrompt("总结以上对话"),
			}
			if modelName, ok := ctxutil.ModelNameFromContext(ctx); ok {
				opts = append(opts, ai.WithModelName(modelName))
			}
			if handleStream != nil {
				opts = append(opts, ai.WithStreaming(handleTextStream(handleStream, true, false)))
			}

			output, _, err := genkit.GenerateData[SummarizeOutput](ctx, g, opts...)
			if err != nil {
				return SummarizeOutput{}, err
			}
			return *output, nil
		},
	)
}
