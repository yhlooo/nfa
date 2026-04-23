package flows

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
)

// ChatInput 对话输入
type ChatInput struct {
	Prompt           string        `json:"prompt"`
	History          []*ai.Message `json:"history,omitempty"`
	MaxContextWindow int64         `json:"maxContextWindow,omitempty"`
}

// ChatOutput 对话输出
type ChatOutput struct {
	Messages          []*ai.Message `json:"messages"`
	LastContextWindow int64         `json:"lastContextWindow,omitempty"`
}

// ChatFlow 对话流程
type ChatFlow = *core.Flow[ChatInput, ChatOutput, struct{}]

// ToolCallError 工具调用错误
type ToolCallError struct {
	Err string `json:"error"`
}

var _ error = ToolCallError{}

// Error 返回错误描述
func (e ToolCallError) Error() string {
	return e.Err
}
