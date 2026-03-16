## Why

当前每次启动 NFA 都是新会话，用户无法恢复之前的对话上下文。这在长时间使用场景下造成不便——用户可能需要重新描述背景、重新建立上下文。会话持久化功能允许用户在退出后恢复之前的对话，提升连续性体验。

## What Changes

- Agent 侧实现会话持久化：每轮 Prompt 结束后保存会话历史到文件
- 新增 `--resume <sessionId>` 命令行参数，支持恢复已有会话
- Agent 实现 `acp.AgentLoader` 接口的 `LoadSession` 方法
- 退出时打印会话恢复提示信息

## Capabilities

### New Capabilities

- `session-persistence`: 会话持久化能力，包括会话保存、加载和恢复

### Modified Capabilities

无

## Impact

- **pkg/agents/**: 新增会话存储逻辑，实现 `LoadSession` 方法
- **pkg/ui/chat/**: 支持通过 `--resume` 恢复会话
- **pkg/commands/**: 新增 `--resume` 命令行参数
- **~/.nfa/sessions/**: 新增会话存储目录
