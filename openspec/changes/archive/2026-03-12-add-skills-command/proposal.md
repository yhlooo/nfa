## Why

当前 NFA 系统已支持技能系统（builtin 和 local 技能），但用户无法在运行时查看已加载的技能列表。用户需要了解当前有哪些技能可用，才能更好地使用 Agent 的技能调用功能。

## What Changes

- 新增 `/skills` 斜杠命令，在 UI 中展示已加载的技能列表
- 技能列表按来源分组显示（Builtin skills / Local skills）
- 显示技能名称和描述信息
- Agent 在初始化时通过 Meta 传递技能列表到 UI

## Capabilities

### New Capabilities
- `skills-list`: 在运行时查看已加载技能列表的能力

### Modified Capabilities
(无现有能力的需求级别变更)

## Impact

**Agent 层** (`pkg/agents/`)
- `meta.go`: 新增 `MetaKeySkills` 常量和 `GetMetaSkillsValue` 工具函数
- `acp.go`: `Initialize` 方法在返回的 Meta 中添加技能列表

**UI 层** (`pkg/ui/chat/`)
- `ui.go`: 新增 `skills` 字段，添加 `/skills` 命令处理，添加 `printSkillsList` 方法
- `input.go`: 在命令选项中添加 `skills`

**依赖关系**: 无新增外部依赖

**向后兼容性**: 完全兼容，现有行为不变
