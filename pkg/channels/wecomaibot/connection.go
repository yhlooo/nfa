package wecomaibot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
)

// Dial 建立连接
func Dial(ctx context.Context, url string, handler Handler) (*Connection, error) {
	conn, resp, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		if resp != nil {
			raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
			_ = resp.Body.Close()
			return nil, fmt.Errorf(
				"connect websocket %q error: %w, status code: %d, body: %s",
				url, err, resp.StatusCode, string(raw),
			)
		}
		return nil, fmt.Errorf("connect websocket %q error: %w", url, err)
	}

	c := &Connection{
		ctx:       ctx,
		conn:      conn,
		handler:   handler,
		done:      make(chan struct{}),
		logger:    logr.FromContextOrDiscard(ctx),
		responses: make(map[string]chan Response),
	}
	go c.receiveLoop()
	go c.pingLoop()

	return c, nil
}

// Handler 回调处理器
type Handler interface {
	// MessageCallback 处理消息回调
	MessageCallback(ctx context.Context, req *MessageCallbackRequest) error
	// EventCallback 处理事件回调
	EventCallback(ctx context.Context, req *EventCallbackRequest) error
}

// Connection 连接
type Connection struct {
	ctx     context.Context
	conn    *websocket.Conn
	handler Handler
	done    chan struct{}
	err     error
	logger  logr.Logger

	responsesLock sync.RWMutex
	responses     map[string]chan Response
}

// receiveLoop 运行接收消息的循环
func (c *Connection) receiveLoop() {
	defer close(c.done)
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		_, data, err := c.conn.ReadMessage()
		if err != nil {
			c.logger.Error(err, "read message error")
			c.err = err
			return
		}

		// 反序列化
		meta := MessageMeta{}
		if err := json.Unmarshal(data, &meta); err != nil {
			c.logger.Info(fmt.Sprintf("WARN ignore invalid message: %s, err %s", string(data), err.Error()))
			continue
		}
		c.logger.V(1).Info(fmt.Sprintf("received message: %s", string(data)))

		if meta.Cmd != "" {
			// 请求
			if err := c.handleRequest(c.ctx, meta.ToRequestMeta(), data); err != nil {
				c.logger.Error(err, "handle request error")
			}
		} else {
			// 响应
			c.responsesLock.Lock()
			respCh, ok := c.responses[meta.Headers.RequestID]
			delete(c.responses, meta.Headers.RequestID)
			c.responsesLock.Unlock()
			if ok {
				select {
				case respCh <- meta.ToResponse():
				default:
				}
			}
		}
	}
}

// pingLoop 运行定时发送 ping 的循环
func (c *Connection) pingLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.done:
			return
		case <-ticker.C:
		}

		_, err := c.Send(c.ctx, RequestMeta{
			Cmd: CmdPing,
			Headers: Headers{
				RequestID: fmt.Sprintf("%x", rand.Uint64()),
			},
		})
		if err != nil {
			c.logger.Error(err, "send ping error")
		}
	}
}

// handleRequest 处理请求
func (c *Connection) handleRequest(ctx context.Context, meta RequestMeta, raw []byte) error {
	switch meta.Cmd {
	case CmdMessageCallback:
		req := MessageCallbackRequest{}
		if err := json.Unmarshal(raw, &req); err != nil {
			return fmt.Errorf("invalid message callback request: %s, error: %w", string(raw), err)
		}
		return c.handler.MessageCallback(ctx, &req)
	case CmdEventCallback:
		req := EventCallbackRequest{}
		if err := json.Unmarshal(raw, &req); err != nil {
			return fmt.Errorf("invalid event callback request: %s, error: %w", string(raw), err)
		}
		return c.handler.EventCallback(ctx, &req)
	default:
		return fmt.Errorf("unknown command: %s", meta.Cmd)
	}
}

// Send 发送请求并接受响应
func (c *Connection) Send(ctx context.Context, req Request) (*Response, error) {
	reqRaw, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request to json error: %w", err)
	}

	// 设置接受响应的 channel
	reqID := req.GetMeta().Headers.RequestID
	respCh := make(chan Response, 1)
	c.responsesLock.Lock()
	c.responses[reqID] = respCh
	c.responsesLock.Unlock()
	defer func() {
		c.responsesLock.Lock()
		delete(c.responses, reqID)
		c.responsesLock.Unlock()
		close(respCh)
	}()

	// 发送请求
	if err := c.conn.WriteMessage(websocket.TextMessage, reqRaw); err != nil {
		return nil, fmt.Errorf("write message to websocket error: %w", err)
	}

	// 等待响应
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.done:
		return nil, fmt.Errorf("connection closed")
	case resp := <-respCh:
		return &resp, err
	}
}

// Done 返回连接结束通知通道
func (c *Connection) Done() <-chan struct{} {
	return c.done
}

// Err 返回导致连接关闭的错误
func (c *Connection) Err() error {
	return c.err
}

// Close 关闭连接
func (c *Connection) Close() error {
	return c.conn.Close()
}
