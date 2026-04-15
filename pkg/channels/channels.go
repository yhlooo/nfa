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
	Send(ctx context.Context, notification acp.SessionNotification) error
}

// UserMessage 用户消息
type UserMessage struct {
	Meta   any
	Prompt []acp.ContentBlock
}
