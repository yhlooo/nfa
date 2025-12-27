package flows

import (
	"context"
	"fmt"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
)

// ChatInput 对话输入
type ChatInput struct {
	ModelName string        `json:"modelName,omitempty"`
	Prompt    string        `json:"prompt"`
	History   []*ai.Message `json:"history,omitempty"`
	Tools     []ai.ToolRef  `json:"tools,omitempty"`
}

// ChatOutput 对话输出
type ChatOutput struct {
	Messages []*ai.Message `json:"messages"`
}

// ChatFlow 对话流程
type ChatFlow = *core.Flow[ChatInput, ChatOutput, *ai.ModelResponseChunk]

// DefaultChatSystemPrompt 默认对话系统提示
func DefaultChatSystemPrompt(_ context.Context, _ any) (string, error) {
	return fmt.Sprintf(`你是一个专业的金融分析师，为用户提供专业的金融咨询服务。

# 目标
你的目标是回答用户咨询的问题

# 回答流程
1. 理解用户问题和意图；
2. 如果当前信息不足以回答用户问题，你可以先通过工具查询相关信息；
3. 当已有信息足以回答用户问题时，停止调用工具，整理已有信息并回答用户问题；

# 要求
- 对话应该逐渐深入，对话刚开始用户的要求较模糊时不要马上进行大量查询、分析和大段陈述，可以先进行简单的启发性陈述并引导用户进一步具体的提问；
- 用用户提问的语言回答问题，比如用户用中文提问就用中文回答，用户用英文提问就用英文回答；

# 其它信息
- 当前时间是： %s；
`, time.Now().Format(time.RFC1123)), nil
}

// ToolCallError 工具调用错误
type ToolCallError struct {
	Err string `json:"error"`
}

var _ error = ToolCallError{}

// Error 返回错误描述
func (e ToolCallError) Error() string {
	return e.Err
}
