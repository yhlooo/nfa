package dualphase

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

// InitializationInput 初始化输入
type InitializationInput struct {
	PreviousP0   string            `json:"previousP0,omitempty"`
	TrainingData []InputOutputPair `json:"trainingData,omitempty"`
	InChinese    bool              `json:"inChinese,omitempty"`
}

// InitializationOutput 初始化输出
type InitializationOutput struct {
	// 生成的 Prompt
	Prompt string `json:"prompt"`
}

// DefineInitializationFlow 定义初始化流程
func DefineInitializationFlow(
	g *genkit.Genkit,
) *core.Flow[InitializationInput, InitializationOutput, struct{}] {
	return genkit.DefineFlow(g, "Initialization", InitializationFlow(g))
}

// InitializationFlow 初始化流程
//
// 根据少量示例输入输出和元提示词生成初始 Prompt p0
func InitializationFlow(g *genkit.Genkit) core.Func[InitializationInput, InitializationOutput] {
	return func(ctx context.Context, in InitializationInput) (InitializationOutput, error) {
		prompt, err := InitializationPrompt(in)
		if err != nil {
			return InitializationOutput{}, fmt.Errorf("make prompt error: %w", err)
		}

		opts := []ai.GenerateOption{
			ai.WithPrompt(prompt),
		}
		if in.PreviousP0 != "" {
			opts = append(opts, ai.WithMessages(ai.NewModelTextMessage(in.PreviousP0)))
		}
		if m, ok := ctxutil.ModelsFromContext(ctx); ok {
			opts = append(opts, ai.WithModelName(m.GetMain()))
		}
		if handleStream := ctxutil.HandleStreamFnFromContext(ctx); handleStream != nil {
			opts = append(opts, ai.WithStreaming(handleStream))
			if in.PreviousP0 != "" {
				_ = handleStream(ctx, &ai.ModelResponseChunk{
					Content: []*ai.Part{ai.NewTextPart(in.PreviousP0)},
					Role:    ai.RoleModel,
				})
			}
			_ = handleStream(ctx, &ai.ModelResponseChunk{
				Content: []*ai.Part{ai.NewTextPart(prompt)},
				Role:    ai.RoleUser,
			})
		}

		resp, err := genkit.Generate(ctx, g, opts...)
		if err != nil {
			return InitializationOutput{}, err
		}
		return InitializationOutput{
			Prompt: resp.Text(),
		}, nil
	}
}

// InitializationPromptTpl 初始化元提示词模版
var InitializationPromptTpl = template.Must(template.New("InitializationPrompt").
	Parse(`You gave me an instruction on a certain task and some example inputs with chain-of-thought. I read the instruction carefully and wrote an output with chain-of-thought for every input correctly. Here are some correct input-output pairs which strictly meet all your requirements:

###Input-Output Pairs###
{{ range .TrainingData -}}
Input: {{ .Input }}
Output: {{ .Output }}

{{ end -}}

The instruction given contains the following parts. Based on the input-output pairs provided, give me the final complete instruction in {{ if .InChinese }}Chinese{{ else }}English{{ end }} without any explanation:

###Task type###
Task type: This is a <...> task.

###Task detailed description###
Task detailed description: <Task detailed description>

###Your output must satisfy the following format and constraints###
Output format(type): <Output format or its type>
Output constraints: <constraints on output>

###You must follow the reasoning process###
<add several reasoning steps if it's necessary>

###Tips###
<add several useful tips from a professional point of view to accomplish this task better>
`))

// InitializationPrompt 获取初始化元提示词
func InitializationPrompt(in InitializationInput) (string, error) {
	buf := &bytes.Buffer{}
	if err := InitializationPromptTpl.Execute(buf, in); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Sentence 句子
type Sentence struct {
	// 句子内容
	Content string `json:"content"`
	// 后缀分割符
	Suffix string `json:"suffix"`
}

// DivideToSentencesInput 将 Prompt 按句子分割输入
type DivideToSentencesInput struct {
	Content string `json:"content"`
}

// DivideToSentencesOutput 将 Prompt 按句子分割输出
type DivideToSentencesOutput struct {
	Sentences []Sentence `json:"sentences"`
}

// DefineDivideToSentencesFlow 定义将 Prompt 按句子分割流程
func DefineDivideToSentencesFlow(
	g *genkit.Genkit,
) *core.Flow[DivideToSentencesInput, DivideToSentencesOutput, struct{}] {
	return genkit.DefineFlow(g, "DivideToSentences", DivideToSentencesFlow(g))
}

// DivideToSentencesFlow 将 Prompt 按句子分割流程
func DivideToSentencesFlow(g *genkit.Genkit) core.Func[DivideToSentencesInput, DivideToSentencesOutput] {
	return func(ctx context.Context, in DivideToSentencesInput) (DivideToSentencesOutput, error) {
		prompt := `将以下 Markdown 格式内容按句子分割，输出分割后的句子列表。每个句子包含：
- content: 包含结束标点的句子内容
- suffix: 可选的后缀。句子后到下一个句子前任何无意义内容都属于后缀，包括换行符、空格等。

注意：
- 将分割后的句子列表中所有 content 和 suffix 首尾相接后必须可以还原成原始输入内容。
- 列表的每一项至少是一个句子

输入：
` + fmt.Sprintf("```\n%s\n```\n", in.Content)

		opts := []ai.GenerateOption{
			ai.WithPrompt(prompt),
		}
		if m, ok := ctxutil.ModelsFromContext(ctx); ok {
			opts = append(opts, ai.WithModelName(m.GetMain()))
		}
		if handleStream := ctxutil.HandleStreamFnFromContext(ctx); handleStream != nil {
			opts = append(opts, ai.WithStreaming(handleStream))
			_ = handleStream(ctx, &ai.ModelResponseChunk{
				Content: []*ai.Part{ai.NewTextPart(prompt)},
				Role:    ai.RoleUser,
			})
		}

		ret, resp, err := genkit.GenerateData[DivideToSentencesOutput](ctx, g, opts...)
		if err != nil {
			return DivideToSentencesOutput{}, err
		}
		if ret == nil {
			return DivideToSentencesOutput{}, fmt.Errorf("invalid result: %s", resp.Text())
		}
		return *ret, nil
	}
}
