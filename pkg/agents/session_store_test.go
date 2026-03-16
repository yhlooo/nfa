package agents

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/coder/acp-go-sdk"
	"github.com/firebase/genkit/go/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAndLoadSession(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "nfa-session-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	sessionID := acp.SessionId("test-session-123")

	// 创建测试消息
	messages := []*ai.Message{
		ai.NewUserTextMessage("你好"),
		ai.NewModelTextMessage("你好！有什么我可以帮助你的吗？"),
	}

	// 保存会话
	err = SaveSession(tmpDir, sessionID, messages)
	require.NoError(t, err)

	// 验证文件存在
	sessionFile := filepath.Join(tmpDir, string(sessionID), SessionFileName)
	_, err = os.Stat(sessionFile)
	require.NoError(t, err)

	// 加载会话
	data, err := LoadSessionData(tmpDir, sessionID)
	require.NoError(t, err)

	// 验证消息数量
	assert.Len(t, data.Messages, 2)

	// 验证消息内容
	assert.Equal(t, ai.RoleUser, data.Messages[0].Role)
	assert.Equal(t, ai.RoleModel, data.Messages[1].Role)
	assert.Equal(t, "你好", data.Messages[0].Text())
	assert.Equal(t, "你好！有什么我可以帮助你的吗？", data.Messages[1].Text())
}

func TestLoadSessionNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nfa-session-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	sessionID := acp.SessionId("non-existent-session")

	_, err = LoadSessionData(tmpDir, sessionID)
	assert.ErrorContains(t, err, "not found")
}

func TestSaveSessionCreatesDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nfa-session-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// 使用不存在的子目录
	sessionsDir := filepath.Join(tmpDir, "sessions")
	sessionID := acp.SessionId("new-session")

	messages := []*ai.Message{
		ai.NewUserTextMessage("测试"),
	}

	// 保存会话（应该自动创建目录）
	err = SaveSession(sessionsDir, sessionID, messages)
	require.NoError(t, err)

	// 验证目录和文件已创建
	sessionFile := filepath.Join(sessionsDir, string(sessionID), SessionFileName)
	_, err = os.Stat(sessionFile)
	require.NoError(t, err)
}

func TestNormalizeMessageMergesTextParts(t *testing.T) {
	// 模拟流式输出：多个 text parts
	msg := &ai.Message{
		Role: ai.RoleModel,
		Content: []*ai.Part{
			ai.NewTextPart("Hello"),
			ai.NewTextPart("!"),
			ai.NewTextPart(" I"),
			ai.NewTextPart("'m"),
			ai.NewTextPart(" a financial analyst."),
		},
	}

	normalized := normalizeMessage(msg)

	// 应该合并为一个 text part
	require.Len(t, normalized.Content, 1)
	assert.Equal(t, "Hello! I'm a financial analyst.", normalized.Content[0].Text)
	assert.True(t, normalized.Content[0].IsText())
}

func TestNormalizeMessagePreservesNonTextParts(t *testing.T) {
	// 混合 text 和 tool request
	msg := &ai.Message{
		Role: ai.RoleModel,
		Content: []*ai.Part{
			ai.NewTextPart("Let me"),
			ai.NewTextPart(" check."),
			ai.NewToolRequestPart(&ai.ToolRequest{
				Name:  "get_stock_price",
				Input: map[string]any{"symbol": "AAPL"},
			}),
			ai.NewTextPart("The price is"),
			ai.NewTextPart(" $150."),
		},
	}

	normalized := normalizeMessage(msg)

	// 应该有 3 个 parts：合并的 text + tool request + 合并的 text
	require.Len(t, normalized.Content, 3)
	assert.Equal(t, "Let me check.", normalized.Content[0].Text)
	assert.True(t, normalized.Content[1].IsToolRequest())
	assert.Equal(t, "The price is $150.", normalized.Content[2].Text)
}
