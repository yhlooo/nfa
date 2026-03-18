# session-persistence Specification (Delta)

## Purpose

扩展会话恢复功能，在 UI 上展示历史消息。

## Requirements

### Requirement: 会话历史回放

Agent SHALL replay conversation history via SessionUpdate during LoadSession.

#### Scenario: 回放用户消息

- **WHEN** Agent 加载会话
- **THEN** Agent MUST 通过 `SessionUpdate` 发送所有用户消息
- **THEN** `SessionUpdate.Update.UserMessageChunk` MUST 包含用户输入文本

#### Scenario: 回放 Agent 消息

- **WHEN** Agent 加载会话
- **THEN** Agent MUST 通过 `SessionUpdate` 发送所有 Agent 回复消息
- **THEN** `SessionUpdate.Update.AgentMessageChunk` MUST 包含 Agent 回复文本

#### Scenario: 回放思考过程

- **WHEN** Agent 加载会话
- **THEN** Agent MUST 通过 `SessionUpdate` 发送所有思考过程
- **THEN** `SessionUpdate.Update.AgentThoughtChunk` MUST 包含思考内容

#### Scenario: 回放工具调用

- **WHEN** Agent 加载会话且历史包含工具调用
- **THEN** Agent MUST 先发送 `SessionUpdate.Update.ToolCall` 状态为 `in_progress`
- **THEN** Agent MUST 后发送 `SessionUpdate.Update.ToolCallUpdate` 状态为 `completed`

#### Scenario: 回放失败处理

- **WHEN** 任一 `SessionUpdate` 发送失败
- **THEN** Agent MUST 中断恢复过程
- **THEN** `LoadSession` MUST 返回错误
- **THEN** UI MUST 显示错误信息
