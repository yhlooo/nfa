## Requirements

### Requirement: Channel 接口实现
YuanbaoBot 结构体 SHALL 实现 `channels.Channel` 接口（`Receive()` / `Send()` / `Err()`），并作为 channel 注册到应用中。

#### Scenario: 正常启动和消息收发
- **WHEN** 配置中 `channels.enabled` 为 `true` 且包含 `yuanbaoBot` 字段，appID 和 appSecret 有效
- **THEN** 应用启动时创建 YuanbaoBot 实例，调用 `Start()` 建立连接，通过 `Receive()` 返回用户消息信道

#### Scenario: 运行错误报告
- **WHEN** 连接过程中发生不可恢复的错误（如收到不可重连的 WebSocket 关闭码）
- **THEN** `Err()` SHALL 返回对应的错误，`Receive()` 信道 SHALL 关闭

### Requirement: HTTP Sign-Token 认证
YuanbaoBot SHALL 通过 HTTP POST 请求获取认证 Token，签名算法为 HMAC-SHA256。sign-token 响应中的 `source` 字段必须保留，用于后续 auth-bind 认证。

#### Scenario: 成功获取 Token
- **WHEN** 使用有效的 appID 和 appSecret 请求 sign-token
- **THEN** 返回包含 `token`、`bot_id`、`source`、`duration` 的响应，Token 被缓存用于后续 WebSocket 认证

#### Scenario: Token 刷新
- **WHEN** Token 即将在 `duration - 5 分钟` 后过期
- **THEN** 自动发起 sign-token 请求刷新 Token，刷新失败时记录错误日志

### Requirement: 混合编码协议
所有 WebSocket 消息统一使用 Protobuf `ConnMsg{head, data}` 包装。`Head.cmdType` 区分消息类型：`0`=请求、`1`=响应、`2`=服务端推送。客户端发送的所有消息使用 Protobuf 编码，服务端推送的 `inbound_message` 的 `data` 字段为 JSON 字符串（需 `json.Unmarshal` 解码）。

#### Scenario: 发送请求
- **WHEN** 发送 auth-bind、ping、send_c2c_message 等命令
- **THEN** 使用 Protobuf 编码，通过 `websocket.BinaryMessage` 发送

#### Scenario: 接收响应
- **WHEN** 收到 `cmdType=1` 的消息
- **THEN** 通过 `msgId` 路由到对应的请求等待者，使用 Protobuf 解码 `data`

#### Scenario: 接收入站消息推送
- **WHEN** 收到 `cmdType=2` 且 `cmd=inbound_message` 的推送
- **THEN** 将 `data` bytes 作为 JSON 字符串解析为 `InboundMessageJSON` 结构体

### Requirement: WebSocket 连接管理
YuanbaoBot SHALL 连接到 WebSocket 网关，连接建立后立即发送 `auth-bind` 命令进行认证绑定。auth-bind 中 `authInfo.source` 必须与 sign-token 返回的 `source` 一致，否则会返回错误码 41103。

#### Scenario: 连接和认证
- **WHEN** WebSocket 连接成功建立
- **THEN** 发送 auth-bind 消息（包含 bizId、authInfo（uid=botID, source, token）、deviceInfo），收到成功响应后开始接收消息

#### Scenario: 认证失败重试
- **WHEN** auth-bind 返回 Token 过期相关错误码
- **THEN** 先强制刷新 Token，再重试连接和认证

### Requirement: 连接层心跳
YuanbaoBot SHALL 通过 WebSocket 发送 `ping` 命令（module: `conn_access`）维持连接活跃。

#### Scenario: 正常心跳
- **WHEN** 连接处于活跃状态
- **THEN** 定期发送 ping 命令

### Requirement: 断线重连
YuanbaoBot SHALL 在 WebSocket 连接意外断开时自动重连，使用指数退避策略（1s, 2s, 5s, 10s, 30s, 60s）。

#### Scenario: 连接断开后重连
- **WHEN** WebSocket 连接意外断开（非主动关闭）
- **THEN** 按指数退避延迟重新执行 signToken → Dial → authBind 流程

#### Scenario: 不重连的关闭码
- **WHEN** WebSocket 收到关闭码 4012/4013/4014/4018/4019/4021
- **THEN** 不触发重连，设置错误并关闭 Receive 信道

### Requirement: 私聊文本消息接收
YuanbaoBot SHALL 接收 `C2C.CallbackAfterSendMsg` 类型的消息推送，解析 JSON `msg_body` 中的文本内容，转为 `UserMessage` 发送到 Receive 信道。

#### Scenario: 收到文本消息
- **WHEN** 收到 `callback_command` 为 `C2C.CallbackAfterSendMsg` 且 `msg_body` 包含文本类型消息
- **THEN** 提取文本内容，构造 `UserMessage{Meta: {replyToAccount, replyMsgID, botID}, Prompt: [TextBlock(text)]}` 发送到 Receive 信道，同时启动回复心跳

#### Scenario: 收到不支持的消息类型
- **WHEN** 收到的消息 `msg_body` 不包含文本内容
- **THEN** 记录日志并忽略该消息

### Requirement: 私聊文本消息发送（非流式）
元宝不支持流式输出，每次 `send_c2c_message` 是一条独立消息。`Send()` SHALL 在 `end=false` 时仅将内容追加到缓冲区，在 `end=true` 时发送拼接好的完整内容并清空缓冲区。

#### Scenario: 累积内容
- **WHEN** `Send()` 被调用且 `end=false`
- **THEN** 将内容追加到 `replyBuff[msgID]`，不发送任何消息

#### Scenario: 发送完整消息
- **WHEN** `Send()` 被调用且 `end=true`
- **THEN** 拼接缓冲区中该 msgID 的所有内容，通过 `send_c2c_message` 一次发送，并清空该消息的缓冲区

### Requirement: 回复心跳
YuanbaoBot SHALL 在收到用户消息后、`Send()` 调用前，每 2 秒向用户发送 RUNNING 心跳；在 `Send()` 调用时发送 FINISH 心跳并停止。通过 `map[string]context.CancelFunc`（key 为用户 ID）管理。

#### Scenario: 启动回复心跳
- **WHEN** 收到用户消息且该用户没有正在运行的心跳
- **THEN** 启动定时器，每 2 秒通过 `send_private_heartbeat` 命令发送 heartbeat=RUNNING

#### Scenario: 停止回复心跳
- **WHEN** `Send()` 被调用
- **THEN** 停止该用户的心跳定时器

#### Scenario: 心跳超时自动停止
- **WHEN** 回复心跳运行超过 30 秒仍未收到 `Send()` 调用
- **THEN** 自动发送 FINISH 并停止心跳

### Requirement: 配置集成
`configs.Channel` 结构体 SHALL 新增 `YuanbaoBot *YuanbaoBotOptions` 字段。配置包含 `appID`、`appSecret` 和可选的 `baseURL`、`websocketURL`。`configs.ChannelsConfig` 包含 `enabled` 布尔字段控制是否启用 channel 功能。

#### Scenario: JSON 配置
- **WHEN** 配置文件中包含 `{"channels": {"enabled": true, "channels": [{"yuanbaoBot": {"appID": "...", "appSecret": "..."}}]}}`
- **THEN** 应用启动时创建 YuanbaoBot 实例并注册为 channel

#### Scenario: 默认 URL
- **WHEN** 配置中未指定 `baseURL` 或 `websocketURL`
- **THEN** `baseURL` 默认为 `https://bot.yuanbao.tencent.com`，`websocketURL` 默认为 `wss://bot-wss.yuanbao.tencent.com/wss/connection`

### Requirement: 元数据透传
消息元数据通过 `map[string]any` 在 channel ↔ chat ↔ agent 间透传，用于关联回复目标。

#### Scenario: 元数据 keys
- **WHEN** 收到用户消息
- **THEN** 在 Meta 中设置 `yuanbaoBotReplyToAccount`（回复目标用户 ID）、`yuanbaoBotReplyMsgID`（消息 ID，用于缓冲区 key）、`yuanbaoBotBotID`（机器人 ID）
