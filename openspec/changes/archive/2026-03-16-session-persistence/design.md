## Context

NFA 当前使用 ACP 协议在 UI (Client) 和 Agent 之间通信。会话历史存储在 Agent 侧的 `Session.history []*ai.Message` 中，每次程序退出后数据丢失。

关键约束：
- `ai.Message` 类型来自 genkit，具有 JSON 标签可直接序列化
- ACP 协议已定义 `LoadSession` 方法和 `AgentLoader` 接口
- Agent 当前只实现了 `acp.Agent` 基础接口

## Goals / Non-Goals

**Goals:**
- 实现会话持久化：每轮 Prompt 结束后保存会话历史
- 支持 `--resume <sessionId>` 恢复已有会话
- 退出时打印恢复提示

**Non-Goals:**
- 异常退出（崩溃、kill -9）的数据保护
- 会话清理/过期机制
- 会话元数据（创建时间、模型配置等）

## Decisions

### 1. 持久化位置：Agent 侧

**选择 Agent 侧持久化**，原因：
- 历史消息本就在 Agent 的 Session 中
- 无需额外传输数据
- 符合 "数据在哪，操作在哪" 原则

**替代方案**：UI 侧持久化 — 需要从 Agent 获取历史消息才能保存，增加复杂度。

### 2. 持久化时机：每轮 Prompt 结束后

**选择每轮 Prompt 结束后保存**，原因：
- 保证数据一致性：恢复后可以继续完整对话
- 简化实现：无需处理异常退出场景

**替代方案**：退出时保存 — 无法处理异常退出，用户可能丢失数据。

### 3. 会话恢复：使用 ACP LoadSession 方法

**选择实现 `acp.AgentLoader` 接口**，原因：
- ACP 协议已定义标准方法
- UI 侧只需调用 `conn.LoadSession()`
- 通过 `AgentCapabilities.LoadSession = true` 声明支持

**替代方案**：扩展 NewSession — 语义不清晰，不符合协议设计。

### 4. 存储格式：简单 JSON 文件

**选择简单 JSON 文件**，结构：
```json
{
  "messages": [
    {"role": "user", "content": [...]},
    {"role": "model", "content": [...]}
  ]
}
```

原因：当前只需要消息历史，避免过度设计。

## Risks / Trade-offs

- [会话文件累积] → 后续可添加清理机制（当前不在范围内）
- [磁盘空间占用] → 暂不处理，等待实际使用反馈
- [JSON 序列化兼容性] → `ai.Message` 已有稳定 JSON 标签，风险可控
