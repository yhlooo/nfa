package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

// SummarizeInput 记忆总结流程输入
type SummarizeInput struct {
	CurrentMemory string        `json:"currentMemory"`
	History       []*ai.Message `json:"history"`
}

// SummarizeOutput 记忆总结流程输出
type SummarizeOutput struct {
	Memory string `json:"memory"`
}

// SummarizeFlow 记忆总结流程
type SummarizeFlow = *core.Flow[SummarizeInput, SummarizeOutput, struct{}]

// SummarizeOptions 记忆总结选项
type SummarizeOptions struct {
	SystemPrompt string
}

// DefineSummarizeFlow 定义记忆总结流程
func DefineSummarizeFlow(g *genkit.Genkit, name string, opts SummarizeOptions) SummarizeFlow {
	return genkit.DefineFlow(g, name,
		func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
			resp, err := genkit.Generate(ctx, g,
				ai.WithSystem(opts.SystemPrompt),
				ai.WithMessages(
					ai.NewUserTextMessage(buildSummarizePrompt(in.CurrentMemory, in.History)),
				),
			)
			if err != nil {
				return SummarizeOutput{}, err
			}

			out := strings.TrimSpace(resp.Message.Text())
			out = strings.TrimPrefix(out, "```markdown")
			out = strings.TrimSuffix(out, "```")
			out += "\n"

			return SummarizeOutput{Memory: out}, nil
		},
	)
}

// buildSummarizePrompt 构建总结记忆的指令
func buildSummarizePrompt(currentMemory string, history []*ai.Message) string {
	var b strings.Builder
	b.WriteString("总结对话历史，更新当前记忆内容\n\n")
	b.WriteString("## 当前记忆\n\n")
	b.WriteString(currentMemory)
	b.WriteString("\n\n")
	b.WriteString("## 对话历史\n\n")
	b.WriteString(historyToText(history))
	return b.String()
}

// historyToText 将消息历史转换为易读的文本格式
func historyToText(history []*ai.Message) string {
	var b strings.Builder
	for _, msg := range history {
		switch msg.Role {
		case ai.RoleSystem:
			b.WriteString(fmt.Sprintf("<system>%s</system>\n", msg.Text()))
		case ai.RoleModel:
			b.WriteString("<assistant>")
			reasoning := ""
			for _, p := range msg.Content {
				if p.IsReasoning() {
					reasoning += p.Text
				}
			}
			if reasoning != "" {
				b.WriteString(fmt.Sprintf("<think>%s</think>\n", reasoning))
			}
			b.WriteString(msg.Text())
			for _, p := range msg.Content {
				if p.IsToolRequest() {
					inRaw, _ := json.Marshal(p.ToolRequest.Input)
					b.WriteString(fmt.Sprintf(
						"<tool-call><id>%s</id><name>%s</name><input>%s</input></tool-call>\n",
						p.ToolRequest.Ref, p.ToolRequest.Name, string(inRaw),
					))
				}
			}
			b.WriteString("</assistant>\n")
		case ai.RoleUser:
			b.WriteString(fmt.Sprintf("<user>%s</user>\n", msg.Text()))
		case ai.RoleTool:
			for _, p := range msg.Content {
				if p.IsToolResponse() {
					outRaw, _ := json.Marshal(p.ToolResponse.Output)
					b.WriteString(fmt.Sprintf(
						"<tool><id>%s</id><result>%s</result></tool>\n",
						p.ToolResponse.Ref, truncate(string(outRaw), 300),
					))
				}
			}
		}
	}

	return b.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

const summarizeDailyMemorySystemPrompt = `你是一个中期记忆管理助手，负责从用户与 AI 的今天的对话历史中提取中期记忆。

## 记忆内容要求
提取并记录以下类型的 **时间敏感** 信息：
- 用户今天关注了什么（询问了哪些股票、话题、市场动态等）
- 今天对话中得出的结论或判断（用户或 AI 的分析结论、决策理由等）
- 用户提到的未来需要关注的事件（财报日期、经济数据发布、政策变化等）
- 跨天持续的讨论主题（同一话题在多天内的进展和变化）

## 记忆原则
- **只记录时间敏感信息** ，长期稳定的偏好（投资风格、风险偏好、回答风格等）属于长期记忆范畴， **不要记录**
- 与当天现有记忆去重合并，不要重复记录已有的内容
- 保持记忆简洁，每条信息一行
- 如果对话中没有新的值得记录的信息，直接返回当天现有记忆原文（如果有），不做修改
- 如果当天没有任何值得记录的信息且当天也无现有记忆，返回空

## 输出格式
直接输出当天中期记忆的完整内容。 **严禁** 在输出中包含任何分析过程、推理步骤、解释说明、前缀或后缀文字。

输出使用 Markdown 格式，参考结构如下：

` + "```markdown" + `
#### 今日关注
- 关注股票/话题：xxx
- 查询内容：xxx

#### 今日结论
- 关于 xxx 的判断：xxx
- 决策：xxx

#### 待关注事件
- xxx（日期/时间范围）

#### .... （其它小节）
` + "```" + `
`

const summarizeLongTermMemorySystemPrompt = `你是一个记忆管理助手，负责从用户与 AI 的对话历史中提取和更新关于用户的长期记忆。

## 记忆内容要求
提取并记录以下类型的信息：
- 用户关注的股票、板块、市场（如询问过哪些股票、对哪些行业感兴趣）
- 用户的投资偏好（投资风格、风险偏好、持仓倾向）
- 用户对回答风格的偏好（详细程度、语气、格式等）
- 对话中获得的经验知识（信息来源、分析框架、工具使用技巧等）

## 记忆原则
- 只记录可长期复用的信息，排除时间敏感数据（如某日具体股价、短期行情、当日新闻）
- 与现有记忆去重合并，不要重复记录已有的内容
- 保持记忆简洁，每条信息一行
- 如果对话中没有新的可记忆信息，直接返回现有记忆原文，不做修改
- 如果现有记忆中的某些信息已经过时或与新信息矛盾，更新或替换它们

## 输出格式
直接输出更新后的完整记忆文件内容。 **严禁** 在输出中包含任何分析过程、推理步骤、解释说明、前缀或后缀文字。

输出使用 Markdown 格式，参考结构如下：

` + "```markdown" + `
### 用户关注
- 关注股票：xxx、xxx
- 关注板块：xxx、xxx

### 投资偏好
- 投资风格：xxx
- 风险偏好：xxx

### 回答风格
- xxx

### 经验与知识
- xxx

### .... （其它小节）
` + "```" + `
`
