package flows

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"

	"github.com/yhlooo/nfa/pkg/agents/gencfg"
	"github.com/yhlooo/nfa/pkg/ctxutil"
)

// TopicRoutingInput 话题分类输入
type TopicRoutingInput struct {
	Messages []*ai.Message `json:"messages"`
}

// TopicRoutingOutput 话题分类输出
type TopicRoutingOutput struct {
	Continue bool  `json:"continue"`
	Topic    Topic `json:"topic,omitempty"`
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
	TopicOthers                 Topic = "Others"
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
			ai.WithConfig(gencfg.GenerateConfig{
				Reasoning: false, // 关闭思考
			}),
		}
		if m, ok := ctxutil.ModelsFromContext(ctx); ok {
			opts = append(opts, ai.WithModelName(m.GetFast()))
		}
		if handleStream := ctxutil.HandleStreamFnFromContext(ctx); handleStream != nil {
			opts = append(opts, ai.WithStreaming(handleTextStream(handleStream, true, false)))
		}

		ret, resp, err := genkit.GenerateData[TopicRoutingOutput](ctx, g, opts...)
		if err != nil {
			return TopicRoutingOutput{}, err
		}
		ctxutil.AddModelUsageToContext(ctx, resp.Usage)
		if ret == nil {
			return TopicRoutingOutput{Continue: true}, nil
		}

		return *ret, nil
	}
}

// TopicRoutingPromptTpl 话题分类 Prompt 模版
var TopicRoutingPromptTpl = template.Must(template.New("TopicRoutingPrompt").
	Parse(`根据对话判断 user 当前是否延续之前话题以及当前讨论的话题类型

## 输出格式
输出为 JSON 格式，包含以下字段
- **continue** (bool) 是否继续之前的话题，当前问题是否是之前正在讨论话题的延续、简单追问、补充说明等。
  澄清几种易混淆情况：
  - 话题类型改变时不属于继续
  - 话题类型没有改变但是讨论的对象发生了改变也不属于继续
- **topic** (string) 当前讨论的话题，可选以下值
  - ` + "`" + `Query` + "`" + `: 信息查询，单纯查询股票价格、资讯等信息，不需要分析
  - ` + "`" + `StockAnalysis` + "`" + `: 个股分析，针对单个股票的技术面、基本面等进行分析
  - ` + "`" + `PortfolioAnalysis` + "`" + `: 投资组合分析，针对投资组合进行分析
  - ` + "`" + `ShortTermTrendForecast` + "`" + `: 短期趋势预测，根据技术面分析、基本面分析、市场资讯、情绪等预测股票近期（一个月内）涨跌趋势
  - ` + "`" + `Basic` + "`" + `: 基础咨询，对一般性的基本的金融知识的咨询
  - ` + "`" + `Comprehensive` + "`" + `: 综合问题，较复杂的综合性问题，需要结合多种分析模式才能解答的问题
  - ` + "`" + `Others` + "`" + `: 其它。不符合以上类型的其它类型话题

## 对话
以下是需要判断话题的对话过程：
` + "```" + `
{{- range .Messages }}
{{- if .Text }} 
{{ .Role }}:
{{ .Text }}
{{- end }}

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
