package memory

import (
	"context"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

// DailySummarizeFlow 中期记忆总结流程类型
type DailySummarizeFlow = *core.Flow[SummarizeInput, SummarizeOutput, struct{}]

// DefineDailySummarizeFlow 定义中期记忆总结流程
func DefineDailySummarizeFlow(g *genkit.Genkit, name string) DailySummarizeFlow {
	return genkit.DefineFlow(g, name,
		func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
			historyText := historyToText(in.History)
			userPrompt := buildDailyUserPrompt(in.ExistingMemory, historyText)

			resp, err := genkit.Generate(ctx, g,
				ai.WithMessages(
					ai.NewSystemTextMessage(dailySummarySystemPrompt),
					ai.NewUserTextMessage(userPrompt),
				),
			)
			if err != nil {
				return SummarizeOutput{}, err
			}

			return SummarizeOutput{Memory: resp.Message.Text()}, nil
		},
	)
}

const dailySummarySystemPrompt = `你是一个中期记忆管理助手，负责从用户与 AI 的对话历史中提取当天的中期记忆。

## 记忆内容要求
提取并记录以下类型的**时间敏感**信息：
- 用户今天关注了什么（询问了哪些股票、话题、市场动态等）
- 今天对话中得出的结论或判断（用户或 AI 的分析结论、决策理由等）
- 用户提到的未来需要关注的事件（财报日期、经济数据发布、政策变化等）
- 跨天持续的讨论主题（同一话题在多天内的进展和变化）

## 记忆原则
- **只记录时间敏感信息**——长期稳定的偏好（投资风格、风险偏好、回答风格等）属于长期记忆范畴，**不要记录**
- 与当天现有记忆去重合并，不要重复记录已有的内容
- 保持记忆简洁，每条信息一行
- 如果对话中没有新的值得记录的信息，直接返回当天现有记忆原文（如果有），不做修改
- 如果当天没有任何值得记录的信息且当天也无现有记忆，返回空

## 输出格式
**直接**输出当天中期记忆的完整内容。**严禁**在输出中包含任何分析过程、推理步骤、解释说明、前缀或后缀文字。

输出使用 Markdown 格式，结构如下：

# YYYY-MM-DD 中期记忆

## 今日关注
- 关注股票/话题：xxx
- 查询内容：xxx

## 今日结论
- 关于 xxx 的判断：xxx
- 决策：xxx

## 待关注事件
- xxx（日期/时间范围）
`

func buildDailyUserPrompt(existingMemory, historyText string) string {
	var b strings.Builder
	if existingMemory != "" {
		b.WriteString("## 当天现有中期记忆\n")
		b.WriteString(existingMemory)
		b.WriteString("\n\n")
	} else {
		b.WriteString("## 当天现有中期记忆\n（暂无）\n\n")
	}
	b.WriteString("## 本轮对话历史\n")
	b.WriteString(historyText)
	b.WriteString(fmt.Sprintf("\n\n注意：以上对话发生在今天，请根据内容更新中期记忆。"))
	return b.String()
}
