package yuanbaobot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	pb "github.com/yhlooo/nfa/pkg/channels/yuanbaobot/proto"
)

// Handler 回调处理器
type Handler interface {
	// OnMessageJSON 处理 JSON 编码的用户消息
	OnMessageJSON(ctx context.Context, msg *InboundMessageJSON) error
}

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
		responses: make(map[string]chan *pb.ConnMsg),
	}
	go c.receiveLoop()
	go c.pingLoop()

	return c, nil
}

// Connection WebSocket 连接
type Connection struct {
	ctx     context.Context
	conn    *websocket.Conn
	handler Handler
	done    chan struct{}
	err     error
	logger  logr.Logger

	seqNo atomic.Uint32

	responsesLock sync.RWMutex
	responses     map[string]chan *pb.ConnMsg
}

// nextSeqNo 获取下一个序列号
func (c *Connection) nextSeqNo() uint32 {
	return c.seqNo.Add(1)
}

// receiveLoop 接收消息循环
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

		// 解码 ConnMsg
		connMsg := &pb.ConnMsg{}
		if err := proto.Unmarshal(data, connMsg); err != nil {
			c.logger.Error(err, "unmarshal ConnMsg error")
			continue
		}
		head := connMsg.Head
		if head == nil {
			c.logger.Info("ignore message without head")
			continue
		}

		c.logger.V(1).Info(fmt.Sprintf(
			"received message: cmd=%s, module=%s, msgId=%s, cmdType=%d, status=%d",
			head.Cmd, head.Module, head.MsgId, head.CmdType, head.Status,
		))

		// cmdType=0: 请求（客户端发送）, cmdType=1: 响应, cmdType=2: 服务端推送
		if head.CmdType == 2 {
			c.handlePush(c.ctx, head, connMsg.Data)
		} else if head.CmdType == 1 {
			// 响应消息：通过 msgId 分发到等待者
			msgID := head.MsgId
			c.responsesLock.Lock()
			respCh, ok := c.responses[msgID]
			delete(c.responses, msgID)
			c.responsesLock.Unlock()
			if ok {
				select {
				case respCh <- connMsg:
				default:
				}
			}
		}
	}
}

// handlePush 处理推送消息
func (c *Connection) handlePush(ctx context.Context, head *pb.Head, data []byte) {
	switch head.Cmd {
	case CmdKickout:
		kickMsg := &pb.KickoutMsg{}
		if err := proto.Unmarshal(data, kickMsg); err == nil {
			c.logger.Info(fmt.Sprintf("received kickout: status=%d, reason=%s", kickMsg.Status, kickMsg.Reason))
		}
	case CmdUpdateMeta:
		c.logger.Info("received update-meta")
	case "inbound_message":
		// data 是 JSON 编码的入站消息
		jsonData := string(data)
		c.logger.V(1).Info(fmt.Sprintf("inbound_message json: %s", jsonData))

		var cbMsg InboundMessageJSON
		if err := json.Unmarshal(data, &cbMsg); err != nil {
			c.logger.Error(err, "unmarshal inbound message json error")
			return
		}
		if cbMsg.CallbackCommand == "" {
			c.logger.Info("ignore inbound_message without callback_command")
			return
		}
		if err := c.handler.OnMessageJSON(ctx, &cbMsg); err != nil {
			c.logger.Error(err, "handle inbound message error")
		}
	default:
		c.logger.Info(fmt.Sprintf("ignore unknown push cmd: %s", head.Cmd))
	}
}

// pingLoop 心跳循环
func (c *Connection) pingLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.done:
			return
		case <-ticker.C:
		}

		head := &pb.Head{
			CmdType: 0,
			Cmd:     CmdPing,
			SeqNo:   c.nextSeqNo(),
			Module:  ModuleConnAccess,
		}
		pingData, _ := proto.Marshal(&pb.PingReq{})
		connMsg, _ := proto.Marshal(&pb.ConnMsg{Head: head, Data: pingData})

		if err := c.conn.WriteMessage(websocket.BinaryMessage, connMsg); err != nil {
			c.logger.Error(err, "send ping error")
		}
	}
}

// SendPB 发送 Protobuf 请求并等待响应
func (c *Connection) SendPB(ctx context.Context, cmd, module string, innerMsg proto.Message) (*pb.ConnMsg, error) {
	innerData, err := proto.Marshal(innerMsg)
	if err != nil {
		return nil, fmt.Errorf("marshal inner message error: %w", err)
	}

	msgID := fmt.Sprintf("%x", c.nextSeqNo())
	head := &pb.Head{
		CmdType: 0,
		Cmd:     cmd,
		SeqNo:   c.nextSeqNo(),
		MsgId:   msgID,
		Module:  module,
	}
	connMsg := &pb.ConnMsg{Head: head, Data: innerData}
	connData, err := proto.Marshal(connMsg)
	if err != nil {
		return nil, fmt.Errorf("marshal ConnMsg error: %w", err)
	}

	c.logger.V(1).Info(fmt.Sprintf("send request: cmd=%s, module=%s, msgId=%s", cmd, module, msgID))

	// 注册响应 channel
	respCh := make(chan *pb.ConnMsg, 1)
	c.responsesLock.Lock()
	c.responses[msgID] = respCh
	c.responsesLock.Unlock()
	defer func() {
		c.responsesLock.Lock()
		delete(c.responses, msgID)
		c.responsesLock.Unlock()
		close(respCh)
	}()

	// 发送二进制 Protobuf
	if err := c.conn.WriteMessage(websocket.BinaryMessage, connData); err != nil {
		return nil, fmt.Errorf("write message error: %w", err)
	}

	// 等待响应
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.done:
		return nil, fmt.Errorf("connection closed")
	case resp := <-respCh:
		return resp, nil
	}
}

// SendPBNoWait 发送 Protobuf 请求不等待响应
func (c *Connection) SendPBNoWait(cmd, module string, innerMsg proto.Message) error {
	innerData, err := proto.Marshal(innerMsg)
	if err != nil {
		return fmt.Errorf("marshal inner message error: %w", err)
	}

	head := &pb.Head{
		CmdType: 0,
		Cmd:     cmd,
		SeqNo:   c.nextSeqNo(),
		Module:  module,
	}
	connMsg := &pb.ConnMsg{Head: head, Data: innerData}
	connData, err := proto.Marshal(connMsg)
	if err != nil {
		return fmt.Errorf("marshal ConnMsg error: %w", err)
	}

	c.logger.V(1).Info(fmt.Sprintf("send request (no-wait): cmd=%s, module=%s", cmd, module))

	if err := c.conn.WriteMessage(websocket.BinaryMessage, connData); err != nil {
		return fmt.Errorf("write message error: %w", err)
	}
	return nil
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
