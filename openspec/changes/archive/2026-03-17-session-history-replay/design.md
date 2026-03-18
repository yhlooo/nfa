# 设计文档

## 数据转换规则

将 `ai.Message` 转换为 `SessionUpdate`：

```
┌─────────────────────────────────────────────────────────────────────────┐
│               ai.Message → SessionUpdate 转换规则                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  用户消息 (Role == "user")                                               │
│  ├── Text Part    → acp.UpdateUserMessageText(text)                    │
│  └── 其他 Part    → acp.UpdateUserMessage(contentBlock)                │
│                                                                         │
│  模型消息 (Role == "model")                                              │
│  ├── Text Part         → acp.UpdateAgentMessageText(text)              │
│  ├── Reasoning Part    → acp.UpdateAgentThoughtText(text)              │
│  ├── ToolRequest Part  → acp.StartToolCall(ref, name, input)           │
│  │                        状态: in_progress                             │
│  └── ToolResponse Part → acp.UpdateToolCall(ref, output)               │
│                           状态: completed                               │
│                                                                         │
│  系统消息 (Role == "system")                                             │
│  └── 跳过，不回放                                                        │
│                                                                         │
│  工具消息 (Role == "tool")                                               │
│  └── 通常嵌入在 model 消息的 ToolResponse Part 中，单独处理              │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

## 工具调用回放顺序

工具调用需要两条 `SessionUpdate`：

1. **StartToolCall**: 状态 `in_progress`，包含工具名和输入参数
2. **UpdateToolCall**: 状态 `completed`，包含输出结果

```
时间线：
  StartToolCall(status=in_progress)
       ↓
  UpdateToolCall(status=completed)
```

## 错误处理

- 如果任一 `SessionUpdate` 发送失败，立即中断恢复过程
- `LoadSession` 返回错误，UI 侧收到错误并显示

## 实现方案

### 修改 LoadSession 方法

```go
func (a *NFAAgent) LoadSession(ctx context.Context, params acp.LoadSessionRequest) (acp.LoadSessionResponse, error) {
    a.lock.Lock()
    defer a.lock.Unlock()

    // 1. 加载会话数据
    data, err := LoadSessionData(a.sessionsDir, params.SessionId)
    if err != nil {
        return acp.LoadSessionResponse{}, fmt.Errorf("load session data error: %w", err)
    }

    // 2. 创建会话
    a.sessions[params.SessionId] = &Session{
        id:      params.SessionId,
        history: data.Messages,
    }

    // 3. 回放历史消息
    if err := a.replayHistory(ctx, params.SessionId, data.Messages); err != nil {
        delete(a.sessions, params.SessionId) // 清理已创建的会话
        return acp.LoadSessionResponse{}, fmt.Errorf("replay history error: %w", err)
    }

    return acp.LoadSessionResponse{}, nil
}
```

### 新增 replayHistory 方法

```go
func (a *NFAAgent) replayHistory(ctx context.Context, sessionID acp.SessionId, messages []*ai.Message) error {
    for _, msg := range messages {
        if err := a.replayMessage(ctx, sessionID, msg); err != nil {
            return err
        }
    }
    return nil
}

func (a *NFAAgent) replayMessage(ctx context.Context, sessionID acp.SessionId, msg *ai.Message) error {
    switch msg.Role {
    case ai.RoleUser:
        return a.replayUserMessage(ctx, sessionID, msg)
    case ai.RoleModel:
        return a.replayModelMessage(ctx, sessionID, msg)
    case ai.RoleTool:
        return a.replayToolMessage(ctx, sessionID, msg)
    case ai.RoleSystem:
        // 系统消息不回放
        return nil
    default:
        return nil
    }
}
```

## 性能考虑

- 历史消息逐条发送，不批量处理
- 如果历史很长，用户会看到消息逐条出现
- 后续可考虑优化（如异步加载、折叠显示），当前版本暂不实现
