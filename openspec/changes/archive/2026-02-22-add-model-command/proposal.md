## Why

当前 NFA 系统中，用户只能在启动时通过配置文件 `~/.nfa/nfa.json` 或命令行参数 `--model` 设置模型，无法在运行时动态切换主模型、快速模型和视觉模型。这导致用户在需要尝试不同模型或根据任务类型切换模型时必须重启应用，影响使用体验。

## What Changes

- 新增交互式 `/model` 命令，支持在运行时切换模型
- 支持多种命令语法：
  - `/model` - 打开主模型选择菜单
  - `/model :fast` - 打开快速模型选择菜单
  - `/model :vision` - 打开视觉模型选择菜单
  - `/model <provider>/<name>` - 直接设置主模型
  - `/model :fast <provider>/<name>` - 直接设置快速模型
  - `/model :vision <provider>/<name>` - 直接设置视觉模型
- 新增模型选择器 UI 组件，支持上下键选择、回车确认、ESC 取消
- 模型列表显示模型描述信息（限制 80 字符，超出用 "..." 替代）
- 选择后自动保存配置到 `~/.nfa/nfa.json`
- Agent 支持从 PromptRequest.Meta 读取 fast 和 vision 模型配置

## Capabilities

### New Capabilities
- `model-selection`: 交互式模型选择和切换能力，支持运行时动态更改主模型、快速模型和视觉模型

### Modified Capabilities
(无现有能力的需求级别变更，仅实现层面的调整)

## Impact

**UI 层** (`pkg/ui/chat/`)
- `ui.go`: ChatUI 新增视图状态管理（input/model_select）、selectedModels 字段、modelDescriptions 字段、配置保存逻辑
- `model_selector.go` (新增): ModelSelector 组件实现选择菜单
- `acp.go`: initAgent 接收模型描述信息，newPrompt 传递模型配置到 Meta

**Agent 层** (`pkg/agents/`)
- `genkit.go`: 返回模型描述信息映射
- `acp.go`: Prompt 方法支持从 Meta 读取 fastModel 和 visionModel
- `meta.go`: 新增 MetaKeyFastModel、MetaKeyVisionModel、MetaKeyModelDescriptions 及相关工具函数

**Config 层** (`pkg/configs/`)
- `load.go`: 新增 SaveConfig 函数实现配置持久化

**Models 层** (`pkg/models/`)
- `providers.go`: ModelConfig 结构体新增 Description 字段

**Commands 层** (`pkg/commands/`)
- `root.go`: context 传递配置文件路径

**依赖关系**: 无新增外部依赖

**向后兼容性**: 完全兼容，现有配置文件格式和行为不变