package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
)

const (
	// 心跳间隔
	pingInterval = 10 * time.Second
	// 最大重连等待时间
	maxReconnectDelay = 30 * time.Second
)

// ConnectionState 连接状态
type ConnectionState struct {
	Connected    bool
	LastUpdate   time.Time
	Reconnecting bool
}

// MarketEvent 市场事件
type MarketEvent struct {
	Type      string      // "book", "price_change", "last_trade_price"
	AssetID   string      // 适用的 asset ID
	Data      interface{} // 具体事件数据
	Timestamp time.Time
}

// Watcher 市场监听器
type Watcher struct {
	ctx          context.Context
	client       *Client
	assetIDs     []string
	conn         *websocket.Conn
	connMu       sync.Mutex
	eventCh      chan MarketEvent
	stateCh      chan ConnectionState
	stopCh       chan struct{}
	wg           sync.WaitGroup
	reconnectIdx int // 重连指数退避计数
}

// NewWatcher 创建市场监听器
func NewWatcher(client *Client, assetIDs []string) *Watcher {
	return &Watcher{
		client:   client,
		assetIDs: assetIDs,
		eventCh:  make(chan MarketEvent, 100),
		stateCh:  make(chan ConnectionState, 10),
		stopCh:   make(chan struct{}),
	}
}

// Start 启动监听
func (w *Watcher) Start(ctx context.Context) error {
	w.ctx = ctx
	logger := logr.FromContextOrDiscard(ctx)

	// 建立初始连接
	if err := w.connect(ctx); err != nil {
		return fmt.Errorf("initial connection error: %w", err)
	}

	// 启动读取和心跳 goroutine
	w.wg.Add(2)
	go w.readLoop(ctx)
	go w.heartbeatLoop(ctx)

	logger.V(1).Info("watcher started")
	return nil
}

// Stop 停止监听
func (w *Watcher) Stop() error {
	close(w.stopCh)
	w.wg.Wait()

	w.connMu.Lock()
	if w.conn != nil {
		_ = w.conn.Close()
	}
	w.connMu.Unlock()

	close(w.eventCh)
	close(w.stateCh)
	return nil
}

// Events 返回事件通道
func (w *Watcher) Events() <-chan MarketEvent {
	return w.eventCh
}

// ConnectionState 返回连接状态通道
func (w *Watcher) ConnectionState() <-chan ConnectionState {
	return w.stateCh
}

// connect 建立连接并发送订阅
func (w *Watcher) connect(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx)

	conn, err := w.client.ConnectMarketChannel(ctx)
	if err != nil {
		return err
	}

	w.connMu.Lock()
	w.conn = conn
	w.connMu.Unlock()

	// 发送订阅请求
	subReq := SubscriptionRequest{
		AssetIDs:    w.assetIDs,
		Type:        "market",
		InitialDump: true,
		Level:       2,
	}

	if err := conn.WriteJSON(subReq); err != nil {
		_ = conn.Close()
		return fmt.Errorf("send subscription error: %w", err)
	}

	logger.V(1).Info("websocket connected and subscribed")

	// 通知连接状态
	w.stateCh <- ConnectionState{
		Connected:    true,
		LastUpdate:   time.Now(),
		Reconnecting: false,
	}

	// 重置重连计数
	w.reconnectIdx = 0

	return nil
}

// readLoop 读取消息循环
func (w *Watcher) readLoop(ctx context.Context) {
	defer w.wg.Done()
	logger := logr.FromContextOrDiscard(ctx)

	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		default:
		}

		w.connMu.Lock()
		conn := w.conn
		w.connMu.Unlock()

		if conn == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		messageType, message, err := conn.ReadMessage()
		if err != nil {
			logger.V(1).Info(fmt.Sprintf("read error: %v", err))
			// 触发重连
			w.handleDisconnect(ctx)
			continue
		}

		// 处理消息
		if messageType == websocket.TextMessage {
			w.handleMessage(message)
		}
	}
}

// handleMessage 处理消息
func (w *Watcher) handleMessage(message []byte) {
	logger := logr.FromContextOrDiscard(w.ctx)

	// 检查是否是 PONG
	if string(message) == "PONG" {
		logger.V(2).Info("received PONG")
		return
	}

	// 解析事件类型
	var baseEvent struct {
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal(message, &baseEvent); err != nil {
		logger.V(1).Info(fmt.Sprintf("parse event type error: %v", err))
		return
	}

	now := time.Now()

	switch baseEvent.EventType {
	case "book":
		var event BookEvent
		if err := json.Unmarshal(message, &event); err != nil {
			logger.V(1).Info(fmt.Sprintf("parse book event error: %v", err))
			return
		}
		w.eventCh <- MarketEvent{
			Type:      "book",
			AssetID:   event.AssetID,
			Data:      &event,
			Timestamp: now,
		}

	case "price_change":
		var event PriceChangeEvent
		if err := json.Unmarshal(message, &event); err != nil {
			logger.V(1).Info(fmt.Sprintf("parse price_change event error: %v", err))
			return
		}
		// 价格变化事件可能包含多个 asset 的变化
		for _, pc := range event.PriceChanges {
			w.eventCh <- MarketEvent{
				Type:      "price_change",
				AssetID:   pc.AssetID,
				Data:      &pc,
				Timestamp: now,
			}
		}

	case "last_trade_price":
		var event LastTradePriceEvent
		if err := json.Unmarshal(message, &event); err != nil {
			logger.V(1).Info(fmt.Sprintf("parse last_trade_price event error: %v", err))
			return
		}
		w.eventCh <- MarketEvent{
			Type:      "last_trade_price",
			AssetID:   event.AssetID,
			Data:      &event,
			Timestamp: now,
		}
	}
}

// heartbeatLoop 心跳循环
func (w *Watcher) heartbeatLoop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.connMu.Lock()
			conn := w.conn
			w.connMu.Unlock()

			if conn != nil {
				if err := conn.WriteMessage(websocket.TextMessage, []byte("PING")); err != nil {
					// 写入失败，触发重连
					w.handleDisconnect(ctx)
				}
			}
		}
	}
}

// handleDisconnect 处理断连
func (w *Watcher) handleDisconnect(ctx context.Context) {
	logger := logr.FromContextOrDiscard(ctx)

	w.connMu.Lock()
	if w.conn != nil {
		_ = w.conn.Close()
		w.conn = nil
	}
	w.connMu.Unlock()

	// 通知断连状态
	w.stateCh <- ConnectionState{
		Connected:    false,
		LastUpdate:   time.Now(),
		Reconnecting: true,
	}

	// 指数退避重连
	w.reconnectIdx++
	delay := w.reconnectDelay()
	logger.V(1).Info(fmt.Sprintf("disconnected, reconnecting in %v...", delay))

	select {
	case <-w.stopCh:
		return
	case <-ctx.Done():
		return
	case <-time.After(delay):
	}

	// 尝试重连
	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		default:
		}

		if err := w.connect(ctx); err != nil {
			logger.V(1).Info(fmt.Sprintf("reconnect error: %v", err))
			w.reconnectIdx++
			delay := w.reconnectDelay()
			logger.V(1).Info(fmt.Sprintf("retrying in %v...", delay))

			select {
			case <-w.stopCh:
				return
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}
		} else {
			logger.V(1).Info("reconnected successfully")
			return
		}
	}
}

// reconnectDelay 计算重连延迟
func (w *Watcher) reconnectDelay() time.Duration {
	// 指数退避: 1s, 2s, 4s, 8s, 16s, 30s, 30s, ...
	delay := time.Duration(1<<uint(w.reconnectIdx-1)) * time.Second
	if delay > maxReconnectDelay {
		delay = maxReconnectDelay
	}
	return delay
}
