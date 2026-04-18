## Why

NFA 目前仅支持企业微信智能机器人作为消息 channel。为了扩展用户触达渠道，需要新增对腾讯元宝智能机器人的支持，让用户可以通过元宝平台与 NFA Agent 交互。

## What Changes

- 新增 `pkg/channels/yuanbaobot` 包，实现 `channels.Channel` 接口，对接元宝机器人 WebSocket API
- 包含 HTTP Sign-Token 认证、WebSocket 连接管理、连接层心跳（5s）、回复心跳（2s）和断线重连
- 支持私聊文本消息的收发和流式回复
- 在 `pkg/configs` 中新增 `YuanbaoBotOptions` 配置结构
- 在 `pkg/commands/root.go` 中注册 yuanbaobot channel 工厂

## Capabilities

### New Capabilities

- `yuanbaobot-channel`: 元宝机器人 channel 实现，包括 WebSocket 连接、认证、消息收发、心跳和重连机制

### Modified Capabilities

（无需修改现有 capability 的需求规格，仅扩展配置和工厂代码）

## Impact

- 新增包：`pkg/channels/yuanbaobot/`（约 5 个文件）
- 修改文件：`pkg/configs/config.go`（新增 `YuanbaoBotOptions` 和 `Channel` 字段）、`pkg/commands/root.go`（新增工厂分支）
- 新增依赖：`github.com/gorilla/websocket`（已有间接依赖，企微 channel 已使用）、`crypto/hmac`（标准库）
- 参考 API 文档：`stories/yuanbao-bot.md`