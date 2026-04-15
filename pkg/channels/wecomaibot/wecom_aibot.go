package wecomaibot

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/coder/acp-go-sdk"
	"github.com/go-logr/logr"

	"github.com/yhlooo/nfa/pkg/agents"
	"github.com/yhlooo/nfa/pkg/channels"
)

const (
	DefaultURL        = "wss://openws.work.weixin.qq.com"
	replyReqIDMetaKey = "wecomAIBotReplyRequestID"
	replyMsgIDMetaKey = "wecomAIBotReplyMessageID"
)

// WeComAIBot 企业微信智能机器人
type WeComAIBot struct {
	BotID  string
	Secret string
	URL    string

	lock        sync.Mutex
	receiveChan chan channels.UserMessage
	conn        *Connection
	err         error
}

var _ channels.Channel = (*WeComAIBot)(nil)
var _ Handler = (*WeComAIBot)(nil)

// Start 开始运行
func (ch *WeComAIBot) Start(ctx context.Context) {
	logger := logr.FromContextOrDiscard(ctx)

	u := ch.URL
	if u == "" {
		u = DefaultURL
	}

	receiveChan := make(chan channels.UserMessage)
	ch.lock.Lock()
	ch.receiveChan = receiveChan
	ch.err = nil
	ch.lock.Unlock()

	go func() {
		defer close(receiveChan)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// 连接
			conn, err := Dial(ctx, u, ch)
			if err != nil {
				logger.Error(err, fmt.Sprintf("connect websocket %q error", u))
				time.Sleep(time.Second)
				continue
			}

			ch.lock.Lock()
			ch.conn = conn
			ch.lock.Unlock()

			// 订阅
			if err := ch.Subscribe(ctx); err != nil {
				logger.Error(err, fmt.Sprintf("subscribe wecom ai bot %q error", ch.BotID))
				if errors.Is(err, ErrSubscriptionError) {
					ch.err = err
					return
				}
				time.Sleep(time.Second)
				continue
			}

			// 等待连接断开
			select {
			case <-ctx.Done():
				return
			case <-conn.Done():
				logger.Error(err, "connection unexpected closed")
			}
		}
	}()
}

// Receive 获取接收用户消息的信道
func (ch *WeComAIBot) Receive() <-chan channels.UserMessage {
	return ch.receiveChan
}

// Send 发送消息
func (ch *WeComAIBot) Send(ctx context.Context, notification acp.SessionNotification) error {
	logger := logr.FromContextOrDiscard(ctx)

	conn, err := ch.getConn()
	if err != nil {
		return err
	}

	content := ""
	switch {
	case notification.Update.AgentThoughtChunk != nil && notification.Update.AgentThoughtChunk.Content.Text != nil:
		content = notification.Update.AgentThoughtChunk.Content.Text.Text
	case notification.Update.AgentMessageChunk != nil && notification.Update.AgentMessageChunk.Content.Text != nil:
		content = notification.Update.AgentMessageChunk.Content.Text.Text
	default:
		// 忽略其它内容
		return nil
	}

	reqID := agents.GetMetaStringValue(notification.Meta, replyReqIDMetaKey)
	msgID := agents.GetMetaStringValue(notification.Meta, replyMsgIDMetaKey)
	if reqID == "" || msgID == "" {
		logger.Info("ignore notification without request ID or message ID")
		return nil
	}

	resp, err := conn.Send(ctx, RespondMessageRequest{
		RequestMeta: RequestMeta{
			Cmd: CmdRespondMessage,
			Headers: Headers{
				RequestID: reqID,
			},
		},
		Body: RespondMessageRequestBody{
			MsgType: StreamMessage,
			Stream: &StreamMessageContent{
				ID:      msgID,
				Finish:  false,
				Content: content,
			},
		},
	})
	if err != nil {
		return err
	}
	if resp.ErrorCode != 0 {
		return fmt.Errorf("send response error: %s, code: %d", resp.ErrorMessage, resp.ErrorCode)
	}

	return nil
}

// MessageCallback 处理消息回调
func (ch *WeComAIBot) MessageCallback(ctx context.Context, req *MessageCallbackRequest) error {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("received message callback")

	content := ""
	switch req.Body.MsgType {
	case TextMessage:
		content = req.Body.Text.Content
	default:
		// TODO: 其它类型暂不支持
		logger.Info(fmt.Sprintf("unsupported message type: %s", req.Body.MsgType))
		return nil
	}

	ch.receiveChan <- channels.UserMessage{
		Meta: map[string]any{
			replyReqIDMetaKey: req.RequestMeta.Headers.RequestID,
			replyMsgIDMetaKey: req.Body.MsgID,
		},
		Prompt: []acp.ContentBlock{
			acp.TextBlock(content),
		},
	}

	return nil
}

// EventCallback 处理事件回调
func (ch *WeComAIBot) EventCallback(ctx context.Context, req *EventCallbackRequest) error {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info(fmt.Sprintf("received event callback: %s", req.Body.Event.EventType))
	// TODO: ...
	return nil
}

// Err 返回运行错误
func (ch *WeComAIBot) Err() error {
	return ch.err
}

// Subscribe 订阅
func (ch *WeComAIBot) Subscribe(ctx context.Context) error {
	conn, err := ch.getConn()
	if err != nil {
		return err
	}

	reqID := fmt.Sprintf("%x", rand.Uint64())

	// 发送订阅请求
	resp, err := conn.Send(ctx, SubscribeRequest{
		RequestMeta: RequestMeta{
			Cmd: CmdSubscribe,
			Headers: Headers{
				RequestID: reqID,
			},
		},
		Body: SubscribeRequestBody{
			BotID:  ch.BotID,
			Secret: ch.Secret,
		},
	})
	if err != nil {
		return err
	}

	if resp.ErrorCode != 0 {
		return fmt.Errorf("%w: code %d, msg: %s", ErrSubscriptionError, resp.ErrorCode, resp.ErrorMessage)
	}

	return nil
}

// getConn 获取连接
func (ch *WeComAIBot) getConn() (*Connection, error) {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	if ch.conn == nil {
		return nil, fmt.Errorf("not connected")
	}
	return ch.conn, nil
}
