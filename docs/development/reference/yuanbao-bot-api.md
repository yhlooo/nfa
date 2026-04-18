# 元宝机器人 API 文档

以下内容基于 [openclaw-plugin-yuanbao](https://www.npmjs.com/package/openclaw-plugin-yuanbao) 源码逻辑整理

---

## 目录

1. [概述](#概述)
2. [连接与认证](#连接与认证)
3. [WebSocket 协议](#websocket-协议)
4. [消息格式](#消息格式)
5. [消息类型](#消息类型)
6. [发送消息](#发送消息)
7. [流式响应与心跳](#流式响应与心跳)
8. [媒体资源](#媒体资源)
9. [群组管理](#群组管理)
10. [错误处理与重连](#错误处理与重连)
11. [配置项](#配置项)

---

## 概述

元宝机器人通过 **WebSocket** 长连接进行实时消息收发，通过 **HTTP API** 进行认证和媒体资源管理。

- WebSocket 网关：`wss://bot-wss.yuanbao.tencent.com/wss/connection`
- HTTP API 域名：`bot.yuanbao.tencent.com`

### 协议编码

WebSocket 通信使用 **Protobuf 二进制编码**，所有消息统一包装在 `ConnMsg` 结构中：

```protobuf
message ConnMsg {
    Head  head = 1;  // 消息头
    bytes data = 2;  // 消息体（不同命令编码方式不同）
}

message Head {
    uint32 cmdType = 1;  // 0=请求, 1=响应, 2=服务端推送
    string cmd     = 2;  // 命令名
    uint32 seqNo   = 3;  // 序列号
    string msgId   = 4;  // 消息 ID（用于请求-响应关联）
    string module  = 5;  // 模块名
    int32  status  = 10; // 状态码
}
```

**注意**：入站消息推送（`inbound_message`）的 `data` 字段是 **JSON 字符串**（非 Protobuf），而连接层和业务层的请求/响应使用 Protobuf 编码。

---

## 连接与认证

### 1. 获取 Token（Sign-Token）

通过 HTTP API 获取认证 Token，用于后续 WebSocket 认证。

**请求：**

```
POST https://bot.yuanbao.tencent.com/api/v5/robotLogic/sign-token
```

**请求体：**

```json
{
  "app_key": "<APP_KEY>",
  "nonce": "<随机 16 字节 hex 字符串>",
  "signature": "<HMAC-SHA256 签名>",
  "timestamp": "2025-01-01T12:00:00+08:00"
}
```

**签名计算：**

```
plain = nonce + timestamp + appKey + appSecret
signature = HMAC-SHA256(appSecret, plain)
```

**响应：**

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "bot_id": "<机器人用户 ID>",
    "token": "<认证 Token>",
    "duration": 2592000,
    "product": "bot",
    "source": "web",
    "create_type": 2
  }
}
```

| 字段 | 说明 |
|------|------|
| `bot_id` | 机器人用户 ID |
| `token` | WebSocket 认证令牌 |
| `duration` | Token 有效期（秒） |
| `source` | 来源标识（实际返回 `"web"`，auth-bind 时必须使用此值） |
| `create_type` | `1` = 一键创建，`2` = 关联创建 |

### 2. Token 缓存与刷新

- Token 缓存在内存中，有效期内复用
- 在过期前 5 分钟自动调度刷新
- 遇到认证失败（错误码 41103/41104/41108）时立即刷新

---

## WebSocket 协议

### 连接命令（Connection Module）

Module 名称：`conn_access`

| 命令 | 说明 | data 编码 |
|------|------|-----------|
| `auth-bind` | 认证绑定 | Protobuf (`AuthBindReq`) |
| `ping` | 心跳 ping | Protobuf (`PingReq`，空消息) |
| `kickout` | 被踢下线通知（服务端推送） | Protobuf (`KickoutMsg`) |
| `update-meta` | 元信息更新（服务端推送） | - |

### 业务命令（Business Module）

Module 名称：`yuanbao_openclaw_proxy`

| 命令 | 说明 | data 编码 |
|------|------|-----------|
| `inbound_message` | 入站消息推送（服务端推送） | **JSON 字符串** |
| `send_c2c_message` | 发送私聊消息 | Protobuf (`SendC2CMessageReq`) |
| `send_group_message` | 发送群组消息 | Protobuf (`SendGroupMessageReq`) |
| `send_private_heartbeat` | 发送私聊回复心跳 | Protobuf (`SendPrivateHeartbeatReq`) |
| `send_group_heartbeat` | 发送群聊回复心跳 | Protobuf (`SendGroupHeartbeatReq`) |
| `query_group_info` | 查询群组信息 | Protobuf |
| `get_group_member_list` | 获取群成员列表 | Protobuf |
| `sync_information` | 同步信息 | Protobuf |

### cmdType 说明

| 值 | 含义 | 方向 |
|----|------|------|
| `0` | 请求 | 客户端 → 服务端 |
| `1` | 响应 | 服务端 → 客户端（对应请求的 msgId） |
| `2` | 服务端推送 | 服务端 → 客户端 |

### 认证绑定（AuthBind）

连接建立后发送 `ConnMsg`，head 的 cmd 为 `auth-bind`，data 为 Protobuf 编码的 `AuthBindReq`：

```protobuf
message AuthBindReq {
    string     bizId      = 1;  // 固定 "ybBot"
    AuthInfo   authInfo   = 2;
    DeviceInfo deviceInfo = 3;
    string     envName    = 5;  // 可选
}

message AuthInfo {
    string uid    = 1;  // sign-token 返回的 bot_id
    string source = 2;  // 必须与 sign-token 返回的 source 一致（如 "web"）
    string token  = 3;  // sign-token 返回的 token
}

message DeviceInfo {
    string appVersion         = 1;
    string appOperationSystem = 2;
    string instanceId         = 10; // 固定 "16"
    string botVersion         = 24;
}
```

**响应**（Protobuf `AuthBindRsp`）：

```protobuf
message AuthBindRsp {
    int32  code    = 1;
    string message = 2;
}
```

### 认证结果码

| 错误码 | 含义 |
|--------|------|
| 41101 | 已认证（成功，重复绑定） |
| 41103 | Token 无效（source 不匹配也会触发） |
| 41104 | Token 过期 |
| 41108 | Token 被强制过期 |
| 50400 | 服务内部错误（可重试） |
| 50503 | 过载控制（可重试） |
| 90001 | 网络故障（可重试） |
| 90003 | 后端返回失败（可重试） |

---

## 消息格式

### 入站消息（服务端推送）

cmd 为 `inbound_message`，cmdType 为 `2`，`data` 字段为 **JSON 字符串**（不是 Protobuf）。

**JSON 结构：**

```json
{
  "callback_command": "C2C.CallbackAfterSendMsg",
  "from_account": "<发送者 ID>",
  "to_account": "<接收者 ID>",
  "sender_nickname": "<发送者昵称>",
  "msg_seq": 2705291063,
  "msg_time": 1776488290,
  "msg_key": "<消息 key>",
  "msg_id": "<消息 ID>",
  "msg_body": [
    {
      "msg_type": "TIMTextElem",
      "msg_content": {
        "text": "消息文本"
      }
    }
  ],
  "cloud_custom_data": "{}",
  "bot_owner_id": "<所有者 ID>",
  "claw_msg_type": 2,
  "log_ext": {
    "trace_id": "<追踪 ID>"
  }
}
```

### 回调命令类型

| 命令 | 说明 |
|------|------|
| `C2C.CallbackAfterSendMsg` | 私聊消息 |
| `Group.CallbackAfterSendMsg` | 群组消息 |
| `C2C.CallbackAfterMsgWithDraw` | 私聊消息撤回 |
| `Group.CallbackAfterRecallMsg` | 群组消息撤回 |

---

## 消息类型

### 文本消息（TIMTextElem）

```json
{
  "msg_type": "TIMTextElem",
  "msg_content": {
    "text": "消息文本内容"
  }
}
```

### 图片消息（TIMImageElem）

```json
{
  "msg_type": "TIMImageElem",
  "msg_content": {
    "uuid": "<图片 UUID>",
    "image_format": 0,
    "image_info_array": [
      {
        "type": 1,
        "size": 12345,
        "width": 800,
        "height": 600,
        "url": "https://..."
      }
    ]
  }
}
```

### 文件消息（TIMFileElem）

```json
{
  "msg_type": "TIMFileElem",
  "msg_content": {
    "uuid": "<文件 UUID>",
    "url": "https://...",
    "file_size": 12345,
    "file_name": "document.pdf"
  }
}
```

---

## 发送消息

### 发送私聊消息（C2C）

发送 `ConnMsg`，head 的 cmd 为 `send_c2c_message`，data 为 Protobuf 编码的 `SendC2CMessageReq`：

```protobuf
message SendC2CMessageReq {
    string                msgId       = 1;
    string                toAccount   = 2;
    string                fromAccount = 3;
    uint32                msgRandom   = 4;
    repeated MsgBodyElement msgBody   = 5;
    uint64                msgSeq      = 7;
    LogInfoExt            logExt      = 8;
}

message MsgBodyElement {
    string      msgType    = 1;  // "TIMTextElem"
    MsgContent  msgContent = 2;
}

message MsgContent {
    string text = 1;
}
```

**响应**（Protobuf `SendC2CMessageRsp`）：

```protobuf
message SendC2CMessageRsp {
    int32  code    = 1;  // 0 = 成功
    string message = 2;
}
```

### 不支持流式输出

元宝不支持流式输出，每次 `send_c2c_message` 是一条独立消息。应在收到完整回复后一次性发送，不要分块发送中间内容。

---

## 流式响应与心跳

### 回复心跳

在 AI 处理时间较长时，通过心跳消息告知客户端正在处理。

**请求**（Protobuf `SendPrivateHeartbeatReq`）：

```protobuf
enum EnumHeartbeat {
    HEARTBEAT_RUNNING = 1;
    HEARTBEAT_FINISH  = 2;
}

message SendPrivateHeartbeatReq {
    string         fromAccount = 1;
    string         toAccount   = 2;
    EnumHeartbeat  heartbeat   = 3;
}
```

- 默认间隔：2 秒
- 最大空闲时间：30 秒（超时自动停止）

---

## 媒体资源

### 获取上传信息

```
POST https://bot.yuanbao.tencent.com/api/resource/genUploadInfo
```

### 获取下载链接

```
GET https://bot.yuanbao.tencent.com/api/resource/v1/download?resourceId=<resource_id>
```

### 媒体限制

- 最大文件大小：20 MB

---

## 群组管理

### 查询群组信息

业务命令：`query_group_info`（Protobuf `QueryGroupInfoReq`）

### 获取群成员列表

业务命令：`get_group_member_list`（Protobuf `GetGroupMemberListReq`）

---

## 错误处理与重连

### WebSocket 重连策略

**指数退避延迟：** 1s, 2s, 5s, 10s, 30s, 60s

**不重连的关闭码：** 4012, 4013, 4014, 4018, 4019, 4021

### 连接层心跳

- Ping 间隔：5 秒
- Ping 请求：`ConnMsg{head: {cmd: "ping", module: "conn_access"}, data: PingReq（空）}`
- Ping 响应：Protobuf `PingRsp`（含 `heartInterval` 和 `timestamp`）

---

## 配置项

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `wsGatewayUrl` | `wss://bot-wss.yuanbao.tencent.com/wss/connection` | WebSocket 网关 |
| `apiDomain` | `bot.yuanbao.tencent.com` | HTTP API 域名 |
| `heartbeatIntervalS` | 5 | 连接心跳间隔（秒） |
| `maxReconnectAttempts` | 100 | 最大重连次数 |
| `mediaMaxMb` | 20 | 媒体文件最大 MB |

---

## Protobuf 定义文件

Protobuf 定义已转换为 `.proto` 文件并生成 Go 代码：

- `pkg/channels/yuanbaobot/proto/conn.proto` — 连接层消息（ConnMsg, Head, AuthBindReq 等）
- `pkg/channels/yuanbaobot/proto/biz.proto` — 业务层消息（SendC2CMessageReq, InboundMessagePush 等）
