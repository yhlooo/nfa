## ADDED Requirements

### Requirement: 会话自动保存

Agent SHALL save session history to persistent storage after each successful Prompt completion.

#### Scenario: Prompt 完成后保存会话

- **WHEN** 用户发起一轮 Prompt 并成功完成
- **THEN** Agent MUST 将当前会话的历史消息保存到 `~/.nfa/sessions/{sessionId}/session.json`

#### Scenario: 保存文件格式

- **WHEN** Agent 保存会话
- **THEN** 文件 MUST 是有效的 JSON 格式，包含 `messages` 数组

### Requirement: 会话恢复

System SHALL support resuming existing sessions via `--resume <sessionId>` command line argument.

#### Scenario: 使用有效 sessionId 恢复会话

- **WHEN** 用户使用 `nfa --resume <valid-session-id>` 启动
- **THEN** UI MUST 调用 `LoadSession` 方法加载会话
- **THEN** Agent MUST 从 `~/.nfa/sessions/{sessionId}/session.json` 加载历史消息
- **THEN** 用户可以继续之前的对话

#### Scenario: 使用无效 sessionId 启动

- **WHEN** 用户使用 `nfa --resume <invalid-session-id>` 启动
- **THEN** 系统 MUST 返回错误并退出

### Requirement: 退出提示

Program SHALL print session resume hint on normal exit.

#### Scenario: 显示恢复提示

- **WHEN** 用户正常退出程序（/exit 或 Ctrl+C）
- **THEN** 系统 MUST 打印：
  ```
  Resume this session with:
  nfa --resume <sessionId>
  ```

### Requirement: Agent 能力声明

Agent SHALL declare session loading capability during initialization.

#### Scenario: 声明 LoadSession 能力

- **WHEN** UI 调用 `Initialize` 方法
- **THEN** Agent MUST 返回 `AgentCapabilities.LoadSession = true`