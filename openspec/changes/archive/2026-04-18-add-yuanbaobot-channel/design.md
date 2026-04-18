## Context

NFA 已有企业微信智能机器人 channel（`pkg/channels/wecomaibot`），基于 WebSocket JSON 协议实现。现在需要新增元宝机器人 channel，同样基于 WebSocket，但协议和认证方式不同。

**元宝机器人 API 特点：**
- WebSocket 网关：`wss://bot-wss.yuanbao.tencent.com/wss/connection`
- HTTP API 基础 URL：`https://bot.yuanbao.tencent.com`
- 认证方式：先通过 HTTP API 获取 Sign-Token，再通过 WebSocket 发送 AuthBind 消息
- Token 有效期较长（实测约 30 天），仍需后台定期刷新
- **混合编码协议**：连接层和业务请求/响应使用 **Protobuf** 编码，入站消息推送的 data 为 **JSON 字符串**
- 有两层心跳：连接层 ping（5s）和回复心跳（RUNNING/FINISH，2s）
- **不支持流式输出**：每次 `send_c2c_message` 是一条独立消息，需攒完再发

**现有 channel 架构：**
- `channels.Channel` 接口：`Receive()` / `Send()` / `Err()`
- WeComAIBot 参考实现：连接管理（`connection.go`）、消息协议（`messages.go`）、主结构体（`wecom_aibot.go`）
- 工厂模式：`configs.ChannelsConfig` + `root.go` switch 分支
- 元数据通过 `map[string]any` 在 channel ↔ chat ↔ agent 间透传

## Goals / Non-Goals

**Goals:**

- 实现完整的元宝机器人 channel，遵循 `channels.Channel` 接口
- 支持私聊文本消息的收发
- 实现 HTTP Sign-Token 认证和自动刷新
- 实现 WebSocket 连接管理和断线重连
- 实现连接层心跳和回复心跳
- 配置与注册集成到现有体系

**Non-Goals:**

- 不支持群组消息（后续扩展）
- 不支持媒体消息（图片、文件、语音、视频）
- 不支持消息撤回处理
- 不支持消息队列策略（merge-text 等）

## Decisions

### 1. 文件结构

```
pkg/channels/yuanbaobot/
├── yuanbao_bot.go   # 主结构体，实现 Channel 接口
├── connection.go    # WebSocket 连接管理（Protobuf 编解码）
├── messages.go      # 常量定义 + JSON 入站消息结构体
├── auth.go          # Sign-Token HTTP 认证 + Token 刷新
├── errors.go        # 错误定义
└── proto/           # Protobuf 定义和生成的 Go 代码
    ├── conn.proto   # 连接层消息定义
    ├── biz.proto    # 业务层消息定义
    ├── conn.pb.go   # 生成的连接层代码
    └── biz.pb.go    # 生成的业务层代码
```

### 2. 混合编码协议

实际调试发现元宝使用混合编码：

| 方向 | 命令/消息 | 编码方式 |
|------|----------|---------|
| 客户端 → 服务端 | 所有请求（auth-bind, ping, send_c2c_message 等） | Protobuf |
| 服务端 → 客户端 | 响应 | Protobuf |
| 服务端 → 客户端 | 入站消息推送（`inbound_message`） | JSON 字符串（在 `ConnMsg.data` bytes 中） |

所有消息统一包装在 Protobuf `ConnMsg{head, data}` 中，`Head.cmdType` 区分消息类型：
- `0` = 请求（客户端发送）
- `1` = 响应（服务端返回）
- `2` = 服务端推送

入站消息 `cmd=inbound_message` 时，`ConnMsg.data` 是 JSON 字符串而非 Protobuf，需用 `json.Unmarshal` 解码。

### 3. 认证流程

```
signToken() ──HTTP POST──▶ 获取 token + source
      │
      ▼
Dial() ──WS 连接──▶ 建立连接
      │
      ▼
authBind() ──WS Protobuf──▶ 认证绑定（source 必须与 sign-token 返回值一致）
      │
      ▼
tokenRefreshLoop() ──定时刷新──▶ 过期前 5min 更新 token
```

关键点：sign-token 返回的 `source` 字段（实测为 `"web"`）必须原样传入 auth-bind 的 `authInfo.source`，否则会报 `token source not match`（错误码 41103）。

Token 刷新后直接重连。暂未实现 `update-meta` 热更新。

### 4. 连接管理

`Connection` 结构体负责 Protobuf 编解码和请求-响应关联：

- 所有发送使用 `websocket.BinaryMessage` + Protobuf 编码
- 请求通过 `Head.msgId` 关联响应
- `cmdType=2` 的消息为服务端推送，分发到 handler
- `cmdType=1` 的消息为响应，通过 `msgId` 路由到等待者

### 5. 消息发送（非流式）

元宝不支持流式输出，每次 `send_c2c_message` 是一条独立消息。实现策略：

- `Send()` 被调用且 `end=false` 时，仅将内容追加到 `replyBuff[msgID]`，不发送
- `Send()` 被调用且 `end=true` 时，发送拼接好的完整内容并清空缓冲区

### 6. 回复心跳

元宝独有的回复心跳机制——在 Agent 处理期间，每 2 秒向用户发送"正在处理"的心跳。

**实现方案**：YuanbaoBot 内部管理，不修改 Channel 接口。

- 收到用户消息时 `startReplyHeartbeat()`，启动定时器
- `Send()` 被调用时 `stopReplyHeartbeat()`，发送 FINISH 并停止
- 最大空闲 30 秒自动停止
- 通过 `map[string]context.CancelFunc`（key 为用户 ID）管理

### 7. 元数据

Meta keys:
- `yuanbaoBotReplyToAccount`: 回复目标用户 ID
- `yuanbaoBotReplyMsgID`: 消息 ID（用于缓冲区 key）
- `yuanbaoBotBotID`: 机器人 ID（用于 fromAccount）

### 8. 配置

```go
type YuanbaoBotOptions struct {
    AppID        string `json:"appID"`
    AppSecret    string `json:"appSecret"`
    BaseURL      string `json:"baseURL,omitempty"`       // HTTP API，默认 https://bot.yuanbao.tencent.com
    WebSocketURL string `json:"websocketURL,omitempty"`  // WebSocket，默认 wss://bot-wss.yuanbao.tencent.com/wss/connection
}
```

配置位于 `configs.ChannelsConfig.Channels` 中，需要 `channels.enabled: true` 开启。

## Risks / Trade-offs

- **[Token 刷新失败]** → 刷新失败会导致 WebSocket 认证失效。缓解：重连机制兜底，指数退避重试刷新。
- **[回复心跳 goroutine 泄漏]** → 如果 `Send()` 未被调用（agent 异常），心跳 goroutine 可能不会停止。缓解：设置最大空闲时间（30s），超时自动停止。
- **[API 非公开]** → 元宝 API 非公开文档，基于逆向整理和实际调试验证。缓解：增加详细日志，方便排查。
