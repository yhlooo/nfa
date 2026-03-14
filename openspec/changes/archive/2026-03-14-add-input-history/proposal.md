## Why

当前 NFA 系统的聊天 UI 不支持历史输入记录功能。用户在多次交互中输入的内容无法被记录和复用，每次都需要重新输入。这降低了交互效率，特别是当用户需要重复发送类似内容时。

参考 shell 的历史记录功能（如 bash 的 `~/.bash_history`），为 NFA 添加类似的历史输入记录功能可以显著提升用户体验。

## What Changes

- 新增历史输入记录功能，将用户输入存储到 `~/.nfa/history.json`
- 历史记录格式：`[{"ts": <unix nano>, "content": "..."}, ...]`
- 保留最近 100 条记录
- 单行模式下支持 ↑/↓ 键浏览历史输入
- 多行模式下禁用历史浏览，保留原有光标移动行为
- 用户提交非空输入后立即保存历史
- 空输入不记录

## Capabilities

### New Capabilities
- `input-history`: 记录和浏览用户历史输入的能力

### Modified Capabilities
(无现有能力的需求级别变更)

## Impact

**新增包** (`pkg/history/`)
- `history.go`: 定义 `History` 和 `Entry` 结构体，实现 Add/Up/Down 等方法
- `load.go`: 实现 `LoadHistory` / `SaveHistory` 函数

**UI 层** (`pkg/ui/chat/`)
- `input.go`: 在 `InputBox` 中添加历史记录支持和上下键处理
- `ui.go`: 初始化时加载历史，提交输入时保存历史

**数据文件**
- 新增 `~/.nfa/history.json` 存储历史记录

**依赖关系**: 无新增外部依赖

**向后兼容性**: 完全兼容，现有行为不变
