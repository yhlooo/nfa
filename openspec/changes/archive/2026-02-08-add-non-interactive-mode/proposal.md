# Proposal: Add Non-Interactive Mode

## Why

nfa 目前只支持交互式终端模式，用户需要手动输入问题并等待回答。对于脚本自动化或快速查询场景，需要一种能够直接回答问题并退出的非交互式模式，提升 nfa 的灵活性和可用性。

## What Changes

- **New flag**: 添加 `-p` / `--print` 标志，表示回答后立即退出
- **Positional argument**: 支持通过位置参数传递问题（第一个参数）
- **Three modes**:
  - `nfa` - 交互式模式（现有行为）
  - `nfa '问题'` - 交互式模式 + 自动发送初始问题
  - `nfa '问题' -p` - 非交互式单次问答模式，回答后退出

## Capabilities

### New Capabilities
- `non-interactive-mode`: 非交互式单次问答功能，包括 CLI 参数解析、初始问题发送、自动退出逻辑

### Modified Capabilities
None. Changes are implementation-level CLI behavior modifications that do not affect core agent or system capabilities.

## Impact

- **Affected code**:
  - `pkg/commands/root.go` - 添加 `-p` flag，修改 `RunE` 逻辑
  - `pkg/ui/chat/ui.go` - 添加 `InitialPrompt` 和 `AutoExitAfterResponse` 选项，修改 `Update` 处理 `PromptResponse` 时自动退出
- **No breaking changes** - 所有现有行为保持不变
- **Dependencies** - 无新依赖，复用现有组件