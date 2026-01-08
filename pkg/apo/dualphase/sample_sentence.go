package dualphase

import (
	"context"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/nfa/pkg/ctxutil"
)

// Sentence 句子
type Sentence struct {
	// 句子内容
	Content string `json:"content"`
	// 后缀分割符
	Suffix string `json:"suffix"`
}

// WeightedSentence 带权重的句子
type WeightedSentence struct {
	Sentence

	// 权重
	Weight float64
	// 忽略优化句子或部分
	Ignore bool
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
	return genkit.DefineFlow(g, "DualPhaseAPODivideToSentences", DivideToSentencesFlow(g))
}

// DivideToSentencesFlow 将 Prompt 按句子分割流程
func DivideToSentencesFlow(g *genkit.Genkit) core.Func[DivideToSentencesInput, DivideToSentencesOutput] {
	return func(ctx context.Context, in DivideToSentencesInput) (DivideToSentencesOutput, error) {
		prompt := `将以下内容按句子分割，输出分割后的句子列表。每个句子包含：
- content: 包含结束标点的句子内容
- suffix: 可选的后缀。句子后到下一个句子前任何无意义内容都属于后缀，包括换行符、空格等。
注意：将分割后的句子列表中所有 content 和 suffix 首尾相接后必须可以还原成原始输入内容。

输入：
` + fmt.Sprintf("```\n%s\n```\n", in.Content)

		opts := []ai.GenerateOption{
			ai.WithPrompt(prompt),
		}
		if modelName, ok := ctxutil.ModelNameFromContext(ctx); ok {
			opts = append(opts, ai.WithModelName(modelName))
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
			return DivideToSentencesOutput{}, fmt.Errorf("empty result: %s", resp.Text())
		}
		return *ret, nil
	}
}

// PromptSentences 按句子分割的 Prompt
type PromptSentences []WeightedSentence

// String 转为文本形式
func (p PromptSentences) String() string {
	ret := &strings.Builder{}
	for _, s := range p {
		ret.WriteString(s.Content)
		ret.WriteString(s.Suffix)
	}
	return ret.String()
}

var (
	weightColors = []string{"\033[48;5;17m", "\033[48;5;18m", "\033[48;5;19m", "\033[48;5;20m", "\033[48;5;21m"}
	resetColor   = "\033[0m"
)

// WithWeightColors 转为带颜色表示权重的文本形式
func (p PromptSentences) WithWeightColors() string {
	minWeight := 0.
	maxWeight := 0.
	for _, s := range p {
		if s.Weight > maxWeight {
			maxWeight = s.Weight
		}
		if s.Weight < minWeight {
			minWeight = s.Weight
		}
	}
	step := (maxWeight - minWeight) / float64(len(weightColors))

	ret := &strings.Builder{}
	for _, s := range p {
		colorI := int((s.Weight - minWeight) / step)
		if colorI < 0 {
			colorI = 0
		}
		if colorI > len(weightColors)-1 {
			colorI = len(weightColors) - 1
		}

		ret.WriteString(weightColors[colorI] + s.Content + resetColor)
		ret.WriteString(s.Suffix)
	}
	return ret.String()
}
