package flows

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

// TopicRoutingInput 话题分类输入
type TopicRoutingInput struct {
	Messages []*ai.Message `json:"messages"`
}

// TopicRoutingOutput 话题分类输出
type TopicRoutingOutput struct {
	Topic Topic `json:"topic"`
}

// Topic 话题
type Topic string

const (
	TopicContinue               Topic = "Continue"
	TopicQuery                  Topic = "Query"
	TopicStockAnalysis          Topic = "StockAnalysis"
	TopicPortfolioAnalysis      Topic = "PortfolioAnalysis"
	TopicShortTermTrendForecast Topic = "ShortTermTrendForecast"
	TopicBasic                  Topic = "Basic"
	TopicComprehensive          Topic = "Comprehensive"
)

// TopicRoutingFlow 话题路由流程
type TopicRoutingFlow = *core.Flow[TopicRoutingInput, TopicRoutingOutput, struct{}]

// DefineTopicRoutingFlow 定义话题分类流程
func DefineTopicRoutingFlow(g *genkit.Genkit) TopicRoutingFlow {
	return genkit.DefineFlow(g, "TopicRouting", NewTopicRoutingFlow(g))
}

// NewTopicRoutingFlow 创建话题分类流程
func NewTopicRoutingFlow(g *genkit.Genkit) core.Func[TopicRoutingInput, TopicRoutingOutput] {
	return func(ctx context.Context, in TopicRoutingInput) (TopicRoutingOutput, error) {
		prompt, err := TopicRoutingPrompt(in)
		if err != nil {
			return TopicRoutingOutput{}, fmt.Errorf("make prompt error: %w", err)
		}
		opts := []ai.GenerateOption{
			ai.WithPrompt(prompt),
		}
		if m, ok := ctxutil.ModelsFromContext(ctx); ok {
			opts = append(opts, ai.WithModelName(m.GetFast()))
		}

		resp, err := genkit.Generate(ctx, g, opts...)
		if err != nil {
			return TopicRoutingOutput{}, err
		}

		return TopicRoutingOutput{Topic: Topic(strings.TrimSpace(resp.Text()))}, nil
	}
}

// TopicRoutingPromptTpl 话题分类 Prompt 模版
var TopicRoutingPromptTpl = template.Must(template.New("TopicRoutingPrompt").
	Parse(`根据对话判断 user 当前是延续之前话题还是要讨论新话题以及新话题类型，可输出以下类型之一：
- ` + "`" + `Continue` + "`" + `: 继续之前的话题，之前正在讨论问题的延续、继续追问、补充说明等
- ` + "`" + `Query` + "`" + `: 信息查询，单纯查询股票价格、资讯等信息，不需要分析
- ` + "`" + `StockAnalysis` + "`" + `: 个股分析，针对单个股票的技术面、基本面等进行分析
- ` + "`" + `PortfolioAnalysis` + "`" + `: 投资组合分析，针对投资组合进行分析
- ` + "`" + `ShortTermTrendForecast` + "`" + `: 短期趋势预测，根据技术面分析、基本面分析、市场资讯、情绪等预测股票近期（一个月内）涨跌趋势
- ` + "`" + `Basic` + "`" + `: 基础咨询，对一般性的基本的金融知识的咨询
- ` + "`" + `Comprehensive` + "`" + `: 综合问题，较复杂的综合性问题，需要结合多种分析模式才能解答的问题

## 输出格式
输出判断的话题类型，不带其它任何内容

## 对话
以下是需要判断话题的对话过程：
` + "```" + `
{{- range .Messages }}
{{ .Role }}: {{ .Text }}
{{- end }}
` + "```" + `
`))

// TopicRoutingPrompt 话题分类
func TopicRoutingPrompt(in TopicRoutingInput) (string, error) {
	buf := &bytes.Buffer{}
	if err := TopicRoutingPromptTpl.Execute(buf, in); err != nil {
		return "", err
	}
	return buf.String(), nil
}
