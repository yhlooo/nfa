package memory

import (
	"strings"
	"testing"

	"github.com/firebase/genkit/go/ai"
	"github.com/stretchr/testify/assert"
)

func TestHistoryToText(t *testing.T) {
	a := assert.New(t)

	history := []*ai.Message{
		ai.NewUserTextMessage("你好，帮我分析一下特斯拉"),
		ai.NewModelTextMessage("好的，我来帮你分析特斯拉。让我先查一下最新数据。"),
		{
			Role: ai.RoleModel,
			Content: []*ai.Part{
				ai.NewToolRequestPart(&ai.ToolRequest{
					Name:  "WebBrowse",
					Input: map[string]any{"url": "https://example.com/tsla"},
				}),
			},
		},
	}

	text := historyToText(history)
	a.Contains(text, "用户: 你好，帮我分析一下特斯拉")
	a.Contains(text, "助手: 好的，我来帮你分析特斯拉")
	a.Contains(text, "助手: [调用工具: WebBrowse]")
}

func TestBuildLongTermUserPrompt_WithExistingMemory(t *testing.T) {
	a := assert.New(t)

	existing := "# NFA Memory\n## 用户关注\n- 关注股票：AAPL"
	historyText := "用户: 分析一下TSLA"

	result := buildLongTermUserPrompt(existing, historyText)
	a.Contains(result, "## 现有记忆")
	a.Contains(result, "## 用户关注")
	a.Contains(result, "## 本轮对话历史")
	a.Contains(result, "用户: 分析一下TSLA")
}

func TestBuildLongTermUserPrompt_EmptyMemory(t *testing.T) {
	a := assert.New(t)
	result := buildLongTermUserPrompt("", "用户: 你好")
	a.NotContains(result, "## 现有记忆")
	a.Contains(result, "## 本轮对话历史")
}

func TestBuildDailyUserPrompt_WithExistingMemory(t *testing.T) {
	a := assert.New(t)

	existing := "# 2026-01-01 中期记忆\n## 今日关注\n- 关注股票：AAPL"
	historyText := "用户: 分析一下TSLA"

	result := buildDailyUserPrompt(existing, historyText)
	a.Contains(result, "## 当天现有中期记忆")
	a.Contains(result, existing)
	a.Contains(result, "## 本轮对话历史")
	a.Contains(result, "用户: 分析一下TSLA")
	a.Contains(result, "注意：以上对话发生在今天")
}

func TestBuildDailyUserPrompt_EmptyMemory(t *testing.T) {
	a := assert.New(t)
	result := buildDailyUserPrompt("", "用户: 你好")
	a.Contains(result, "（暂无）")
	a.Contains(result, "## 本轮对话历史")
}

func TestTruncate(t *testing.T) {
	a := assert.New(t)
	a.Equal("abc", truncate("abc", 5))
	a.Equal("abcde", truncate("abcde", 5))
	a.True(strings.HasSuffix(truncate("abcdefghij", 5), "..."))
	a.Equal(8, len(truncate("abcdefghij", 5)))
}
