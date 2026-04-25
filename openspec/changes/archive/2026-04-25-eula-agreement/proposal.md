## Why

NFA 作为一个面向终端用户的金融 AI Agent 应用，需要在首次使用前与用户建立明确的法律关系。用户协议（EULA）确认是行业标准做法，确保用户知晓并同意使用条款后才能开始使用，同时支持协议版本变更后的重新确认。

## What Changes

- 新增 `pkg/eula` 模块，内嵌 `eula.md` 协议文件
- 首次启动时展示协议内容，要求用户明确同意后方可使用
- 用户同意后记录协议版本的 SHA256 签名，后续启动自动跳过
- 协议内容变更时检测 SHA256 不匹配，要求用户重新确认
- 每次启动将当前协议写入 `~/.nfa/eula.md`，方便用户随时查阅
- EULA 检查仅对 agent 主命令生效，管理子命令不受影响

## Capabilities

### New Capabilities

- `eula-agreement`: 用户协议确认机制，包括协议展示、用户确认、版本追踪和变更检测

### Modified Capabilities

（无）

## Impact

- `pkg/eula/` — 新增模块（eula 逻辑、嵌入式协议文件、i18n 消息）
- `pkg/commands/root.go` — `PersistentPreRunE` 中插入 EULA 检查调用
- `~/.nfa/eula.md` — 每次启动写入的协议文件
- `~/.nfa/eula_signed.sha256sum` — 用户签署记录文件
