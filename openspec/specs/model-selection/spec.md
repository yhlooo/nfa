# model-selection Specification

## Purpose
TBD - created by archiving change add-model-command. Update Purpose after archive.
## Requirements
### Requirement: Interactive model selection menu
用户通过 `/model` 命令进入模型选择菜单时，系统 MUST 显示可用模型列表，支持键盘导航和选择。

#### Scenario: Open primary model selection menu
- **WHEN** 用户输入 `/model` 并按回车
- **THEN** 系统隐藏输入框
- **AND** 显示主模型选择菜单
- **AND** 菜单标题为 "Select primary model"
- **AND** 菜单显示所有可用模型列表

#### Scenario: Open light model selection menu
- **WHEN** 用户输入 `/model :light` 并按回车
- **THEN** 系统隐藏输入框
- **AND** 显示轻量模型选择菜单
- **AND** 菜单标题为 "Select light model"

#### Scenario: Open vision model selection menu
- **WHEN** 用户输入 `/model :vision` 并按回车
- **THEN** 系统隐藏输入框
- **AND** 显示视觉模型选择菜单
- **AND** 菜单标题为 "Select vision model"

### Requirement: Model list display with descriptions
选择菜单 MUST 显示模型列表，每个模型项包含序号、模型名称和可选的描述信息。

#### Scenario: Display models with descriptions
- **WHEN** 模型选择菜单打开
- **THEN** 每个模型项显示为 "序号. provider/modelName - 描述"
- **AND** 描述信息限制在 80 字符以内
- **AND** 超过 80 字符的描述使用 "..." 截断
- **AND** 当前选中的模型前有 "❯" 指示符
- **AND** 其他模型前有 "  " 空格

#### Scenario: Display models without descriptions
- **WHEN** 模型没有描述信息
- **THEN** 模型项仅显示 "序号. provider/modelName"
- **AND** 不显示描述分隔符和内容

### Requirement: Keyboard navigation in model selector
用户 MUST 能够使用键盘在模型列表中导航和选择。

#### Scenario: Navigate with arrow keys
- **WHEN** 用户按向下键或 Tab 键
- **THEN** 选择光标移动到下一个模型
- **AND** 选择指示符 "❯" 更新到新位置

#### Scenario: Navigate with arrow keys up
- **WHEN** 用户按向上键或 Shift+Tab
- **THEN** 选择光标移动到上一个模型
- **AND** 选择指示符 "❯" 更新到新位置

#### Scenario: Confirm model selection
- **WHEN** 用户按回车键
- **THEN** 系统应用当前选中的模型到对应类型（primary/light/vision）
- **AND** 保存配置到 `~/.nfa/nfa.json`
- **AND** 显示成功消息 "✓ {type} model set to: {model}"
- **AND** 返回输入框视图

#### Scenario: Cancel model selection
- **WHEN** 用户按 ESC 键
- **THEN** 系统取消选择操作
- **AND** 返回输入框视图
- **AND** 不修改任何模型配置

### Requirement: Direct model setting via command
用户 MUST 能够通过命令直接指定模型，无需进入选择菜单。

#### Scenario: Set primary model directly
- **WHEN** 用户输入 `/model ollama/llama3.2` 并按回车
- **THEN** 系统设置主模型为 "ollama/llama3.2"
- **AND** 保存配置到 `~/.nfa/nfa.json`
- **AND** 显示成功消息 "✓ Primary model set to: ollama/llama3.2"

#### Scenario: Set light model with explicit target
- **WHEN** 用户输入 `/model :light ollama/qwen3:14b` 并按回车
- **THEN** 系统设置轻量模型为 "ollama/qwen3:14b"
- **AND** 保存配置
- **AND** 显示成功消息

#### Scenario: Set vision model directly
- **WHEN** 用户输入 `/model :vision aliyun/qwen3-vl-plus` 并按回车
- **THEN** 系统设置视觉模型为 "aliyun/qwen3-vl-plus"
- **AND** 保存配置
- **AND** 显示成功消息

### Requirement: Model configuration persistence
系统 MUST 在用户选择或设置模型后立即保存配置到文件。

#### Scenario: Save configuration after model selection
- **WHEN** 用户在选择菜单中确认模型选择
- **THEN** 系统更新 `~/.nfa/nfa.json` 中的 `defaultModels` 字段
- **AND** 对应模型类型（primary/light/vision）的值更新为新选择的模型
- **AND** 配置文件使用 2 空格缩进
- **AND** 其他配置项保持不变

#### Scenario: Save configuration after direct model setting
- **WHEN** 用户通过命令直接设置模型
- **THEN** 系统立即保存配置到 `~/.nfa/nfa.json`
- **AND** 更新对应的模型类型字段

### Requirement: Agent model configuration from UI
系统 MUST 将用户选择的模型配置传递给 Agent，在每次对话时使用正确的模型。

#### Scenario: Pass selected models to agent
- **WHEN** UI 发送新的 PromptRequest 到 Agent
- **THEN** PromptRequest.Meta 包含当前选择的模型配置
- **AND** Meta 中包含 "modelName" 字段（主模型）
- **AND** Meta 中包含 "lightModel" 字段（轻量模型，如果已设置）
- **AND** Meta 中包含 "visionModel" 字段（视觉模型，如果已设置）

#### Scenario: Agent uses provided models
- **WHEN** Agent 接收到包含模型配置的 PromptRequest
- **THEN** Agent 使用 Meta 中指定的模型
- **AND** 如果 Meta 中未指定某个模型类型，使用配置文件中的默认值

### Requirement: Model description availability
系统 MUST 在初始化时从 Agent 获取模型的描述信息并在 UI 中使用。

#### Scenario: Agent provides model descriptions
- **WHEN** Agent 初始化
- **THEN** Agent 返回可用模型列表
- **AND** Agent 返回模型描述映射（map[string]string）
- **AND** 描述映射的键为 "provider/modelName" 格式

#### Scenario: UI receives and stores descriptions
- **WHEN** UI 初始化 Agent
- **THEN** UI 接收模型描述映射
- **AND** 存储在 ChatUI.modelDescriptions 字段
- **AND** 传递给 ModelSelector 用于显示

### Requirement: Current model highlighting
选择菜单 MUST 标识当前正在使用的模型。

#### Scenario: Highlight current primary model
- **WHEN** 打开主模型选择菜单
- **THEN** 选择光标默认定位到当前使用的主模型
- **AND** 当前模型前显示 "❯" 指示符

#### Scenario: Highlight current light model
- **WHEN** 打开轻量模型选择菜单
- **THEN** 选择光标默认定位到当前使用的轻量模型

### Requirement: Command syntax error handling
系统 MUST 对无效的命令语法提供清晰的错误提示。

#### Scenario: Invalid model format
- **WHEN** 用户输入 `/model invalid-format` 并按回车
- **THEN** 系统显示错误消息
- **AND** 错误消息提示正确的格式 "Expected: provider/name"

#### Scenario: Invalid target prefix
- **WHEN** 用户输入 `/model :invalid-target ollama/llama3.2`
- **THEN** 系统显示错误消息
- **AND** 错误消息提示有效的目标选项 "Expected: :primary, :light, :vision"

### Requirement: Model selection during agent processing
系统 MUST 允许在 Agent 思考过程中切换模型，新模型在下一次对话时生效。

#### Scenario: Switch model while agent is thinking
- **WHEN** Agent 正在处理请求
- **AND** 用户打开模型选择菜单
- **THEN** 系统允许切换模型
- **AND** 新模型设置在当前 Agent 请求完成后生效
- **AND** 下一次对话使用新选择的模型

