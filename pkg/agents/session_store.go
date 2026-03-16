package agents

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coder/acp-go-sdk"
	"github.com/firebase/genkit/go/ai"
)

// SessionData 会话持久化数据结构
type SessionData struct {
	Messages []*ai.Message `json:"messages"`
}

// SessionsDirName 会话存储目录名
const SessionsDirName = "sessions"

// SessionFileName 会话文件名
const SessionFileName = "session.json"

// SaveSession 保存会话到文件
func SaveSession(sessionsDir string, sessionID acp.SessionId, history []*ai.Message) error {
	// 确保目录存在
	sessionDir := filepath.Join(sessionsDir, string(sessionID))
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		return fmt.Errorf("create session directory error: %w", err)
	}

	// 规范化消息（合并连续的 text parts）
	normalizedHistory := normalizeMessages(history)

	// 准备数据
	data := SessionData{
		Messages: normalizedHistory,
	}

	// 序列化
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session data error: %w", err)
	}

	// 写入文件
	sessionFile := filepath.Join(sessionDir, SessionFileName)
	if err := os.WriteFile(sessionFile, content, 0o644); err != nil {
		return fmt.Errorf("write session file error: %w", err)
	}

	return nil
}

// normalizeMessages 规范化消息，合并每条消息中连续的 text parts
func normalizeMessages(messages []*ai.Message) []*ai.Message {
	result := make([]*ai.Message, 0, len(messages))
	for _, msg := range messages {
		result = append(result, normalizeMessage(msg))
	}
	return result
}

// normalizeMessage 规范化单条消息，合并连续的 text parts
func normalizeMessage(msg *ai.Message) *ai.Message {
	if msg == nil {
		return nil
	}

	// 合并连续的 text parts
	var mergedParts []*ai.Part
	var textBuilder strings.Builder

	flushText := func() {
		if textBuilder.Len() > 0 {
			mergedParts = append(mergedParts, ai.NewTextPart(textBuilder.String()))
			textBuilder.Reset()
		}
	}

	for _, part := range msg.Content {
		if part.IsText() {
			textBuilder.WriteString(part.Text)
		} else {
			// 遇到非 text part，先 flush 累积的 text
			flushText()
			mergedParts = append(mergedParts, part)
		}
	}
	// flush 最后累积的 text
	flushText()

	// 如果没有 parts，保留原始
	if len(mergedParts) == 0 && len(msg.Content) > 0 {
		mergedParts = msg.Content
	}

	return &ai.Message{
		Content:  mergedParts,
		Role:     msg.Role,
		Metadata: msg.Metadata,
	}
}

// LoadSessionData 从文件加载会话数据
func LoadSessionData(sessionsDir string, sessionID acp.SessionId) (*SessionData, error) {
	sessionFile := filepath.Join(sessionsDir, string(sessionID), SessionFileName)

	content, err := os.ReadFile(sessionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session %q not found", sessionID)
		}
		return nil, fmt.Errorf("read session file error: %w", err)
	}

	var data SessionData
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("unmarshal session data error: %w", err)
	}

	return &data, nil
}
