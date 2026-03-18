# 会话历史回放

## Summary

通过 `--resume` 恢复会话时，在 UI 上展示历史消息。遵循 ACP 协议规范，在 `LoadSession` 过程中通过 `SessionUpdate` 流式回放历史消息。

## Motivation

当前 `--resume` 功能仅恢复 Agent 内部的对话上下文，但 UI 侧没有任何历史消息展示。用户恢复会话后看到的是空白界面，无法了解之前的对话内容，体验不完整。

根据 ACP 协议文档，`session/load` 流程应包含历史消息回放：

```
Client->>Agent: session/load (sessionId)
Note over Agent: Restore session context
Note over Agent,Client: Replay conversation history...
Agent->>Client: session/update
Agent->>Client: session/update
Note over Agent,Client: All content streamed
Agent-->>Client: session/load response
```

## Capabilities

- Agent 在 `LoadSession` 过程中遍历历史消息并发送 `SessionUpdate`
- UI 侧无需修改，`MessageViewport` 正常接收并渲染历史消息
- 支持完整历史回放：用户消息、Agent 消息、思考过程、工具调用

## Impact

- **pkg/agents/acp.go**: 修改 `LoadSession` 方法，添加历史消息回放逻辑
- **pkg/agents/session_store.go**: 新增消息转换辅助函数

## Out of Scope

- UI 侧修改（复用现有渲染逻辑）
- 会话存储格式修改
- 历史消息分页/折叠（后续优化）
