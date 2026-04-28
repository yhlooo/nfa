package memory

import (
	"context"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

// SummarizeInput 记忆总结流程输入
type SummarizeInput struct {
	ExistingMemory string        `json:"existingMemory"`
	History        []*ai.Message `json:"history"`
}

// SummarizeOutput 记忆总结流程输出
type SummarizeOutput struct {
	Memory string `json:"memory"`
}

// SummarizeFlow 记忆总结流程
type SummarizeFlow = *core.Flow[SummarizeInput, SummarizeOutput, struct{}]

// DefineSummarizeFlow 定义记忆总结流程
func DefineSummarizeFlow(g *genkit.Genkit, name string) SummarizeFlow {
	return genkit.DefineFlow(g, name,
		func(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
			historyText := historyToText(in.History)
			userPrompt := buildUserPrompt(in.ExistingMemory, historyText)

			resp, err := genkit.Generate(ctx, g,
				ai.WithMessages(
					ai.NewSystemTextMessage(summarySystemPrompt),
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

const summarySystemPrompt = `你是一个记忆管理助手，负责从用户与 AI 的对话历史中提取和更新关于用户的长期记忆。

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
**直接**输出更新后的完整记忆文件内容。**严禁**在输出中包含任何分析过程、推理步骤、解释说明、前缀或后缀文字。你的回复必须以 "# NFA Memory" 开头，只包含以下 Markdown 格式的记忆文件内容：

# NFA Memory

## 用户关注
- 关注股票：xxx、xxx
- 关注板块：xxx、xxx

## 投资偏好
- 投资风格：xxx
- 风险偏好：xxx

## 回答风格
- xxx

## 经验与知识
- xxx
`

func buildUserPrompt(existingMemory, historyText string) string {
	var b strings.Builder
	if existingMemory != "" {
		b.WriteString("## 现有记忆\n")
		b.WriteString(existingMemory)
		b.WriteString("\n\n")
	}
	b.WriteString("## 本轮对话历史\n")
	b.WriteString(historyText)
	return b.String()
}

// historyToText 将消息历史转换为易读的文本格式
func historyToText(history []*ai.Message) string {
	var b strings.Builder
	for _, msg := range history {
		roleLabel := roleLabel(msg.Role)
		for _, part := range msg.Content {
			text := partText(part)
			if text == "" {
				continue
			}
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(fmt.Sprintf("%s: %s", roleLabel, text))
		}
	}
	return b.String()
}

func roleLabel(role ai.Role) string {
	switch role {
	case ai.RoleUser:
		return "用户"
	case ai.RoleModel:
		return "助手"
	case ai.RoleTool:
		return "工具"
	case ai.RoleSystem:
		return "系统"
	default:
		return string(role)
	}
}

func partText(part *ai.Part) string {
	switch part.Kind {
	case ai.PartText:
		return part.Text
	case ai.PartReasoning:
		return ""
	case ai.PartToolRequest:
		if part.ToolRequest != nil {
			return fmt.Sprintf("[调用工具: %s]", part.ToolRequest.Name)
		}
		return "[调用工具]"
	case ai.PartToolResponse:
		if part.ToolResponse != nil {
			output := toolOutputText(part.ToolResponse.Output)
			return "[工具返回: " + truncate(output, 300) + "]"
		}
		return "[工具返回]"
	default:
		if part.Text != "" {
			return part.Text
		}
		return ""
	}
}

func toolOutputText(output any) string {
	if output == nil {
		return ""
	}
	switch v := output.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
