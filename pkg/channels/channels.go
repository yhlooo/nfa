package channels

import (
	"context"

	"github.com/coder/acp-go-sdk"
)

// Channel 通道
type Channel interface {
	// Receive 获取接收用户消息的信道
	Receive() <-chan UserMessage
	// Send 发送消息
	Send(ctx context.Context, meta any, notification *acp.SessionNotification, end bool) error
	// Err 获取运行错误
	Err() error
}

// UserMessage 用户消息
type UserMessage struct {
	Meta   map[string]any
	Prompt []acp.ContentBlock
}
