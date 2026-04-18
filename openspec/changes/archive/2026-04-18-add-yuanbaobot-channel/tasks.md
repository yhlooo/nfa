## 1. 消息协议定义

- [x] 1.1 创建 `pkg/channels/yuanbaobot/messages.go`，定义消息元信息结构（`MessageMeta`、`RequestMeta`、`Response`），包含 `cmd` 和 `module` 字段
- [x] 1.2 定义连接层命令常量（`CmdAuthBind`、`CmdPing`、`CmdKickout`、`CmdUpdateMeta`）和业务命令常量（`CmdSendC2CMessage`、`CmdSendPrivateHeartbeat`），以及对应的 module 常量
- [x] 1.3 定义 AuthBind 请求/响应结构（`AuthBindRequest`、`AuthBindResponse`）
- [x] 1.4 定义入站消息结构（`CallbackMessage`、`MsgBody`、`MsgContent`），支持 `TIMTextElem` 等消息类型
- [x] 1.5 定义出站消息结构（`SendC2CMessageRequest`、`SendC2CMessageResponse`），包含 `msgBody` 和 `TIMTextElem`
- [x] 1.6 定义回复心跳消息结构（`ReplyHeartbeat`），包含 `heartbeat` 字段（1=RUNNING, 2=FINISH）
- [x] 1.7 创建 `pkg/channels/yuanbaobot/errors.go`，定义错误常量

## 2. HTTP Sign-Token 认证

- [x] 2.1 创建 `pkg/channels/yuanbaobot/auth.go`，实现 `SignTokenRequest` 和 `SignTokenResponse` 结构体
- [x] 2.2 实现 HMAC-SHA256 签名计算函数：`plain = nonce + timestamp + appKey + appSecret`
- [x] 2.3 实现 `signToken()` HTTP 请求方法，POST 到 `/api/v5/robotLogic/sign-token`
- [x] 2.4 实现 Token 缓存和自动刷新逻辑（过期前 5 分钟刷新）

## 3. WebSocket 连接管理

- [x] 3.1 创建 `pkg/channels/yuanbaobot/connection.go`，实现 `Dial()` 函数建立 WebSocket 连接
- [x] 3.2 实现 `Connection` 结构体，包含 receiveLoop（消息分发）和 sendAndWait（请求-响应关联）
- [x] 3.3 实现 `authBind()` 方法，连接建立后发送认证绑定消息
- [x] 3.4 实现 `pingLoop()`，每 5 秒发送连接层心跳
- [x] 3.5 实现消息分发逻辑：根据 `cmd`+`module` 区分推送消息和响应，分发到对应 handler

## 4. 主结构体与 Channel 接口

- [x] 4.1 创建 `pkg/channels/yuanbaobot/yuanbao_bot.go`，定义 `YuanbaoBot` 结构体（AppKey、AppSecret、URL、连接、缓冲区、心跳管理）
- [x] 4.2 实现 `Start()` 方法：启动 connectLoop goroutine（signToken → Dial → authBind → 等待断开 → 重连）
- [x] 4.3 实现 `Receive()` 方法：返回 `receiveChan`
- [x] 4.4 实现消息回调处理：解析入站消息，构造 `UserMessage`（Meta 含 toAccount 和 msgID），发送到 `receiveChan`
- [x] 4.5 实现 `Send()` 方法：从 meta 获取 toAccount 和 msgID，缓冲内容，通过 `send_c2c_message` 发送
- [x] 4.6 实现回复心跳管理：`startReplyHeartbeat()` / `stopReplyHeartbeat()`，使用 `map[string]context.CancelFunc` 管理
- [x] 4.7 实现 `Err()` 方法

## 5. 配置与注册集成

- [x] 5.1 在 `pkg/configs/config.go` 中新增 `YuanbaoBotOptions` 结构体（AppKey、AppSecret、URL）和 `Channel.YuanbaoBot` 字段
- [x] 5.2 在 `pkg/commands/root.go` 中新增 `case chOpts.YuanbaoBot != nil` 分支，创建并注册 YuanbaoBot 实例

## 6. 代码质量验证

- [x] 6.1 执行 `go fmt ./...` 确保代码格式化
- [x] 6.2 执行 `go vet ./...` 确保静态分析通过
- [x] 6.3 执行 `go test ./...` 确保所有测试通过