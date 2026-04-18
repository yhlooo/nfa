package yuanbaobot

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/coder/acp-go-sdk"
	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"github.com/yhlooo/nfa/pkg/agents"
	"github.com/yhlooo/nfa/pkg/channels"
	pb "github.com/yhlooo/nfa/pkg/channels/yuanbaobot/proto"
)

const (
	DefaultBaseURL      = "https://bot.yuanbao.tencent.com"
	DefaultWebSocketURL = "wss://bot-wss.yuanbao.tencent.com/wss/connection"

	replyToAccountMetaKey = "yuanbaoBotReplyToAccount"
	replyMsgIDMetaKey     = "yuanbaoBotReplyMsgID"
	botIDMetaKey          = "yuanbaoBotBotID"

	// 回复心跳间隔
	replyHeartbeatInterval = 2 * time.Second
	// 回复心跳最大空闲时间
	replyHeartbeatMaxIdle = 30 * time.Second
)

// 重连退避延迟
var reconnectDelays = []time.Duration{
	1 * time.Second, 2 * time.Second, 5 * time.Second,
	10 * time.Second, 30 * time.Second, 60 * time.Second,
}

// YuanbaoBot 元宝机器人
type YuanbaoBot struct {
	AppKey       string
	AppSecret    string
	BaseURL      string
	WebSocketURL string

	lock        sync.Mutex
	receiveChan chan channels.UserMessage
	conn        *Connection
	auth        *AuthManager
	err         error
	replyBuff   map[string]string
	botID       string

	// 回复心跳管理
	heartbeatLock   sync.Mutex
	heartbeatCancel map[string]context.CancelFunc // key: toAccount
}

var _ channels.Channel = (*YuanbaoBot)(nil)
var _ Handler = (*YuanbaoBot)(nil)

// Start 开始运行
func (ch *YuanbaoBot) Start(ctx context.Context) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("start connecting yuanbao bot")

	baseURL := ch.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	wsURL := ch.WebSocketURL
	if wsURL == "" {
		wsURL = DefaultWebSocketURL
	}

	ch.auth = NewAuthManager(ch.AppKey, ch.AppSecret, baseURL, logger)
	ch.heartbeatCancel = make(map[string]context.CancelFunc)

	receiveChan := make(chan channels.UserMessage)
	ch.lock.Lock()
	ch.receiveChan = receiveChan
	ch.err = nil
	ch.lock.Unlock()

	go ch.connectLoop(ctx, wsURL, receiveChan, logger)
}

// connectLoop 连接循环（含重连）
func (ch *YuanbaoBot) connectLoop(ctx context.Context, u string, receiveChan chan channels.UserMessage, logger logr.Logger) {
	defer func() {
		if ch.conn != nil {
			_ = ch.conn.Close()
		}
		close(receiveChan)
	}()

	reconnectAttempts := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 1. 获取 Token
		tokenCache, err := ch.auth.SignToken(ctx)
		if err != nil {
			logger.Error(err, "sign token error")
			ch.sleepWithBackoff(ctx, reconnectAttempts)
			reconnectAttempts++
			continue
		}
		ch.lock.Lock()
		ch.botID = tokenCache.BotID
		ch.lock.Unlock()

		// 2. 建立 WebSocket 连接
		conn, err := Dial(ctx, u, ch)
		if err != nil {
			logger.Error(err, fmt.Sprintf("connect websocket %q error", u))
			ch.sleepWithBackoff(ctx, reconnectAttempts)
			reconnectAttempts++
			continue
		}
		ch.lock.Lock()
		ch.conn = conn
		ch.lock.Unlock()
		reconnectAttempts = 0

		// 3. 认证绑定
		if err := ch.authBind(ctx, tokenCache); err != nil {
			logger.Error(err, "auth bind error")
			_ = conn.Close()
			ch.lock.Lock()
			ch.conn = nil
			ch.lock.Unlock()

			ch.sleepWithBackoff(ctx, reconnectAttempts)
			reconnectAttempts++
			continue
		}

		logger.Info("yuanbao bot connected and authenticated", "botID", tokenCache.BotID)

		// 4. 启动 Token 刷新
		refreshDone := make(chan struct{})
		go ch.tokenRefreshLoop(ctx, tokenCache.Duration, refreshDone, logger)

		// 5. 等待连接断开
		select {
		case <-ctx.Done():
			close(refreshDone)
			return
		case <-conn.Done():
			close(refreshDone)
			logger.Info("connection closed, will reconnect")
		}

		// 检查是否为不可重连的关闭码
		if closeErr := conn.Err(); closeErr != nil {
			if websocket.IsCloseError(closeErr, 4012, 4013, 4014, 4018, 4019, 4021) {
				logger.Error(closeErr, "connection closed with non-reconnect code")
				ch.lock.Lock()
				ch.err = closeErr
				ch.lock.Unlock()
				return
			}
		}

		_ = conn.Close()
		ch.lock.Lock()
		ch.conn = nil
		ch.lock.Unlock()
	}
}

// tokenRefreshLoop Token 自动刷新
func (ch *YuanbaoBot) tokenRefreshLoop(ctx context.Context, durationSec int64, done chan struct{}, logger logr.Logger) {
	interval := time.Duration(durationSec)*time.Second - TokenRefreshMargin
	if interval <= 0 {
		interval = time.Duration(durationSec-300) * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
			if _, err := ch.auth.RefreshToken(ctx); err != nil {
				logger.Error(err, "refresh token error")
			}
		}
	}
}

// authBind 认证绑定
func (ch *YuanbaoBot) authBind(ctx context.Context, tokenCache *TokenCache) error {
	conn, err := ch.getConn()
	if err != nil {
		return err
	}

	respConnMsg, err := conn.SendPB(ctx, CmdAuthBind, ModuleConnAccess, &pb.AuthBindReq{
		BizId: "ybBot",
		AuthInfo: &pb.AuthInfo{
			Uid:    tokenCache.BotID,
			Source: tokenCache.Source,
			Token:  tokenCache.Token,
		},
		DeviceInfo: &pb.DeviceInfo{
			AppVersion:         "1.0.0",
			AppOperationSystem: "linux",
			BotVersion:         "1.0.0",
			InstanceId:         "16",
		},
	})
	if err != nil {
		return fmt.Errorf("send auth-bind error: %w", err)
	}

	// 解码响应
	authRsp := &pb.AuthBindRsp{}
	if err := proto.Unmarshal(respConnMsg.Data, authRsp); err != nil {
		return fmt.Errorf("unmarshal AuthBindRsp error: %w", err)
	}

	code := authRsp.Code
	if code == int32(pb.RetCode_ALREADY_AUTH) {
		// 重复绑定，视为成功
		return nil
	}
	if code != 0 {
		if TokenExpiredCodes[int(code)] {
			// Token 过期，强制刷新
			_, _ = ch.auth.RefreshToken(ctx)
		}
		return fmt.Errorf("%w: auth-bind failed with code %d, msg: %s",
			ErrAuthFailed, code, authRsp.Message)
	}

	return nil
}

// Receive 获取接收用户消息的信道
func (ch *YuanbaoBot) Receive() <-chan channels.UserMessage {
	return ch.receiveChan
}

// Send 发送消息
func (ch *YuanbaoBot) Send(ctx context.Context, meta any, notification *acp.SessionNotification, end bool) error {
	logger := logr.FromContextOrDiscard(ctx)

	conn, err := ch.getConn()
	if err != nil {
		return err
	}

	content := ""
	if notification != nil {
		switch {
		case notification.Update.AgentThoughtChunk != nil && notification.Update.AgentThoughtChunk.Content.Text != nil:
			content = notification.Update.AgentThoughtChunk.Content.Text.Text
		case notification.Update.AgentMessageChunk != nil && notification.Update.AgentMessageChunk.Content.Text != nil:
			content = notification.Update.AgentMessageChunk.Content.Text.Text
		default:
		}
	}
	if content == "" && !end {
		return nil
	}

	toAccount := agents.GetMetaStringValue(meta, replyToAccountMetaKey)
	msgID := agents.GetMetaStringValue(meta, replyMsgIDMetaKey)
	botID := agents.GetMetaStringValue(meta, botIDMetaKey)
	if toAccount == "" || msgID == "" {
		logger.Info("ignore notification without toAccount or msgID")
		return nil
	}

	// 停止回复心跳
	ch.stopReplyHeartbeat(toAccount)

	// 获取之前缓冲的内容
	ch.lock.Lock()
	if ch.replyBuff == nil {
		ch.replyBuff = make(map[string]string)
	}
	content = ch.replyBuff[msgID] + content
	if end {
		delete(ch.replyBuff, msgID)
	} else {
		ch.replyBuff[msgID] = content
	}
	ch.lock.Unlock()

	if !end {
		return nil
	}

	// 发送完整消息
	respConnMsg, err := conn.SendPB(ctx, CmdSendC2CMessage, ModuleBiz, &pb.SendC2CMessageReq{
		MsgId:       msgID,
		ToAccount:   toAccount,
		FromAccount: botID,
		MsgRandom:   rand.Uint32(),
		MsgSeq:      uint64(time.Now().UnixMilli()),
		MsgBody: []*pb.MsgBodyElement{
			{
				MsgType: MsgTypeText,
				MsgContent: &pb.MsgContent{
					Text: content,
				},
			},
		},
	})
	if err != nil {
		return err
	}

	// 解码响应
	sendRsp := &pb.SendC2CMessageRsp{}
	if err := proto.Unmarshal(respConnMsg.Data, sendRsp); err != nil {
		return fmt.Errorf("unmarshal SendC2CMessageRsp error: %w", err)
	}
	if sendRsp.Code != 0 {
		return fmt.Errorf("send c2c message error: code %d, msg: %s", sendRsp.Code, sendRsp.Message)
	}

	return nil
}

// OnMessage 处理入站消息
func (ch *YuanbaoBot) OnMessageJSON(ctx context.Context, msg *InboundMessageJSON) error {
	logger := logr.FromContextOrDiscard(ctx)

	switch msg.CallbackCommand {
	case CallbackC2CSendMsg:
		// 私聊消息
	default:
		logger.Info(fmt.Sprintf("ignore callback command: %s", msg.CallbackCommand))
		return nil
	}

	// 提取文本内容
	text := ""
	for _, body := range msg.MsgBody {
		if body.MsgType == MsgTypeText {
			text = body.MsgContent.Text
			break
		}
	}
	if text == "" {
		logger.Info("ignore message without text content")
		return nil
	}

	toAccount := msg.FromAccount
	msgID := msg.MsgID

	// 启动回复心跳
	ch.startReplyHeartbeat(ctx, toAccount, ch.getBotID())

	ch.receiveChan <- channels.UserMessage{
		Meta: map[string]any{
			replyToAccountMetaKey: toAccount,
			replyMsgIDMetaKey:     msgID,
			botIDMetaKey:          ch.getBotID(),
		},
		Prompt: []acp.ContentBlock{
			acp.TextBlock(text),
		},
	}

	return nil
}

// Err 返回运行错误
func (ch *YuanbaoBot) Err() error {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	return ch.err
}

// getConn 获取连接
func (ch *YuanbaoBot) getConn() (*Connection, error) {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	if ch.conn == nil {
		return nil, ErrNotConnected
	}
	return ch.conn, nil
}

// getBotID 获取 bot ID
func (ch *YuanbaoBot) getBotID() string {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	return ch.botID
}

// startReplyHeartbeat 启动回复心跳
func (ch *YuanbaoBot) startReplyHeartbeat(ctx context.Context, toAccount, botID string) {
	ch.heartbeatLock.Lock()
	defer ch.heartbeatLock.Unlock()

	// 已有心跳则跳过
	if _, ok := ch.heartbeatCancel[toAccount]; ok {
		return
	}

	hbCtx, cancel := context.WithCancel(ctx)
	ch.heartbeatCancel[toAccount] = cancel

	go func() {
		defer func() {
			ch.heartbeatLock.Lock()
			delete(ch.heartbeatCancel, toAccount)
			ch.heartbeatLock.Unlock()
		}()

		ticker := time.NewTicker(replyHeartbeatInterval)
		defer ticker.Stop()

		maxIdle := time.NewTimer(replyHeartbeatMaxIdle)
		defer maxIdle.Stop()

		for {
			select {
			case <-hbCtx.Done():
				return
			case <-maxIdle.C:
				// 超时自动停止
				ch.sendHeartbeat(hbCtx, toAccount, botID, pb.EnumHeartbeat_HEARTBEAT_FINISH)
				return
			case <-ticker.C:
				if err := ch.sendHeartbeat(hbCtx, toAccount, botID, pb.EnumHeartbeat_HEARTBEAT_RUNNING); err != nil {
					return
				}
			}
		}
	}()
}

// stopReplyHeartbeat 停止回复心跳
func (ch *YuanbaoBot) stopReplyHeartbeat(toAccount string) {
	ch.heartbeatLock.Lock()
	defer ch.heartbeatLock.Unlock()

	if cancel, ok := ch.heartbeatCancel[toAccount]; ok {
		cancel()
		delete(ch.heartbeatCancel, toAccount)
	}
}

// sendHeartbeat 发送回复心跳
func (ch *YuanbaoBot) sendHeartbeat(ctx context.Context, toAccount, botID string, heartbeat pb.EnumHeartbeat) error {
	conn, err := ch.getConn()
	if err != nil {
		return err
	}

	return conn.SendPBNoWait(CmdSendPrivateHeartbeat, ModuleBiz, &pb.SendPrivateHeartbeatReq{
		FromAccount: botID,
		ToAccount:   toAccount,
		Heartbeat:   heartbeat,
	})
}

// sleepWithBackoff 指数退避等待
func (ch *YuanbaoBot) sleepWithBackoff(ctx context.Context, attempt int) {
	idx := attempt
	if idx >= len(reconnectDelays) {
		idx = len(reconnectDelays) - 1
	}
	delay := reconnectDelays[idx]

	select {
	case <-ctx.Done():
	case <-time.After(delay):
	}
}
