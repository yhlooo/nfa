package flows

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
)

// ChatInput 对话输入
type ChatInput struct {
	ModelName     string        `json:"modelName,omitempty"`
	Prompt        string        `json:"prompt"`
	History       []*ai.Message `json:"history,omitempty"`
	AssistantName string        `json:"assistantName,omitempty"`
}

// ChatOutput 对话输出
type ChatOutput struct {
	Messages             []*ai.Message `json:"messages"`
	CurrentAssistantName string        `json:"currentAssistantName,omitempty"`
}

// ChatFlow 对话流程
type ChatFlow = *core.Flow[ChatInput, ChatOutput, *ai.ModelResponseChunk]

// ToolCallError 工具调用错误
type ToolCallError struct {
	Err string `json:"error"`
}

var _ error = ToolCallError{}

// Error 返回错误描述
func (e ToolCallError) Error() string {
	return e.Err
}

// GenerateOptionsFn 获取生成选项的方法
type GenerateOptionsFn func() []ai.GenerateOption

// FixedGenerateOptions 固定的生成选项
func FixedGenerateOptions(opts ...ai.GenerateOption) GenerateOptionsFn {
	return func() []ai.GenerateOption {
		return opts
	}
}
