## 1. Models Layer - 数据结构扩展

- [x] 1.1 在 `pkg/models/providers.go` 的 `ModelConfig` 结构体中添加 `Description string` 字段
- [x] 1.2 更新模型提供商文档，说明 Description 字段的可选性和用途

## 2. Config Layer - 配置持久化

- [x] 2.1 在 `pkg/configs/load.go` 中实现 `SaveConfig(path string, cfg Config) error` 函数
- [x] 2.2 确保保存的 JSON 使用 2 空格缩进（`json.MarshalIndent`）
- [x] 2.3 添加配置保存单元测试

## 3. Agent Layer - 模型描述收集

- [x] 3.1 在 `pkg/agents/meta.go` 中添加常量 `MetaKeyModelDescriptions = "modelDescriptions"`
- [x] 3.2 在 `pkg/agents/meta.go` 中添加 `GetMetaStringMapValue` 工具函数
- [x] 3.3 在 `pkg/agents/meta.go` 中添加 `MetaKeyFastModel = "fastModel"` 常量
- [x] 3.4 在 `pkg/agents/meta.go` 中添加 `MetaKeyVisionModel = "visionModel"` 常量
- [x] 3.5 修改 `pkg/agents/genkit.go` 的 `NewGenkitWithModels` 函数，返回值增加 `map[string]string`（模型描述）
- [x] 3.6 在 `NewGenkitWithModels` 中实现逻辑：遍历所有 Provider，从 `ModelConfig.Description` 收集描述信息
- [x] 3.7 修改 `NFAAgent` 结构体，添加 `modelDescriptions map[string]string` 字段
- [x] 3.8 修改 `NFAAgent.Initialize` 方法，将模型描述放入返回的 Meta 中
- [x] 3.9 修改 `NFAAgent.Prompt` 方法，从 Meta 读取 `fastModel` 和 `visionModel`，更新 `m.Fast` 和 `m.Vision`

## 4. Commands Layer - Context 传递

- [x] 4.1 在 `pkg/commands/root.go` 中定义 context key：`type cfgPathContextKey struct{}`
- [x] 4.2 添加 `ContextWithCfgPath` 和 `CfgPathFromContext` 辅助函数
- [x] 4.3 在 `PersistentPreRunE` 中将 `cfgPath` 放入 context
- [x] 4.4 修改 `NewChatUI` 调用，传递 context（确保 UI 可以从 context 获取 cfgPath）

## 5. UI Layer - 基础数据结构

- [x] 5.1 在 `pkg/ui/chat/ui.go` 的 `ChatUI` 结构体中添加 `cfgPath string` 字段
- [x] 5.2 在 `ChatUI` 结构体中添加 `cfg configs.Config` 字段
- [x] 5.3 在 `ChatUI` 结构体中添加 `selectedModels models.Models` 字段
- [x] 5.4 在 `ChatUI` 结构体中添加 `modelDescriptions map[string]string` 字段
- [x] 5.5 定义 `type viewState string` 和常量 `viewStateInput`, `viewStateModelSelect`
- [x] 5.6 在 `ChatUI` 结构体中添加 `viewState viewState` 字段
- [x] 5.7 定义 `type ModelType string` 和常量 `ModelTypeMain`, `ModelTypeFast`, `ModelTypeVision`
- [x] 5.8 在 `ChatUI` 结构体中添加 `modelSelector *ModelSelector` 字段
- [x] 5.9 在 `ChatUI` 结构体中添加 `selectorTargetModel ModelType` 字段

## 6. UI Layer - ModelSelector 组件

- [x] 6.1 创建新文件 `pkg/ui/chat/model_selector.go`
- [x] 6.2 定义 `ModelItem` 结构体（ID 和 Desc 字段）
- [x] 6.3 定义 `ModelSelector` 结构体，实现 `tea.Model` 接口
- [x] 6.4 实现 `NewModelSelector` 构造函数，接收 target、available、descriptions、current 参数
- [x] 6.5 实现 `ModelSelector.Init()` 方法
- [x] 6.6 实现 `ModelSelector.Update()` 方法，处理上下键和窗口大小消息
- [x] 6.7 实现 `ModelSelector.View()` 方法，渲染无框列表样式（"❯ 序号. model - 描述"）
- [x] 6.8 实现描述截断逻辑：超过 80 字符使用 "..." 替代
- [x] 6.9 实现 `ModelSelector.Selected()` 方法，返回当前选中的模型 ID

## 7. UI Layer - 状态管理和视图切换

- [x] 7.1 修改 `ChatUI.initAgent()`，从 context 获取 cfgPath 和 cfg
- [x] 7.2 在 `initAgent` 中接收 `modelDescriptions` Meta，存储到 `ui.modelDescriptions`
- [x] 7.3 初始化 `ui.selectedModels = cfg.DefaultModels`
- [x] 7.4 实现 `enterModelSelectMode(target ModelType)` 方法
- [x] 7.5 实现 `exitModelSelectMode()` 方法
- [x] 7.6 实现 `applyModelAndSave(modelType ModelType, modelID string) error` 方法
- [x] 7.7 在 `applyModelAndSave` 中更新 `selectedModels` 和 `cfg.DefaultModels`
- [x] 7.8 在 `applyModelAndSave` 中调用 `configs.SaveConfig` 保存配置
- [x] 7.9 修改 `ChatUI.Update()` 方法，根据 `viewState` 路由消息到 `updateInInputState` 或 `updateInModelSelectState`
- [x] 7.10 实现 `updateInInputState` 方法，处理正常输入状态
- [x] 7.11 实现 `updateInModelSelectState` 方法，处理选择菜单状态

## 8. UI Layer - 命令解析和处理

- [x] 8.1 在 `updateInInputState` 的 `tea.KeyEnter` 分支中添加 `/model` 命令检测
- [x] 8.2 实现 `/model` 命令处理，调用 `enterModelSelectMode(ModelTypeMain)`
- [x] 8.3 实现 `/model :main` 命令处理
- [x] 8.4 实现 `/model :fast` 命令处理
- [x] 8.5 实现 `/model :vision` 命令处理
- [x] 8.6 实现 `handleDirectModelSet(content string) (handled bool)` 方法
- [x] 8.7 在 `handleDirectModelSet` 中解析命令格式：`/model [:target] <provider>/<name>`
- [x] 8.8 实现目标类型验证（main, fast, vision）
- [x] 8.9 在解析成功后调用 `applyModelAndSave` 并显示成功/失败消息
- [x] 8.10 在解析失败时显示错误提示

## 9. UI Layer - 视图渲染

- [x] 9.1 修改 `ChatUI.View()` 方法，根据 `viewState` 条件渲染
- [x] 9.2 在 `viewStateInput` 时渲染 InputBox（现有逻辑）
- [x] 9.3 在 `viewStateModelSelect` 时渲染 ModelSelector
- [x] 9.4 确保两种状态下都渲染 MessageViewport 和 Token Usage

## 10. UI Layer - PromptRequest Meta 传递

- [x] 10.1 修改 `ChatUI.newPrompt()` 方法
- [x] 10.2 在构造 `PromptRequest` 时添加 Meta 字段
- [x] 10.3 如果 `selectedModels.Main` 非空，设置 `Meta["modelName"]`
- [x] 10.4 如果 `selectedModels.Fast` 非空，设置 `Meta["fastModel"]`
- [x] 10.5 如果 `selectedModels.Vision` 非空，设置 `Meta["visionModel"]`

## 11. UI Layer - 键盘交互细节

- [x] 11.1 在 `updateInModelSelectState` 中处理 `tea.KeyEnter`：确认选择
- [x] 11.2 在确认后调用 `applyModelAndSave`，显示 "✓ {type} model set to: {model}"
- [x] 11.3 在确认后调用 `exitModelSelectMode` 返回输入状态
- [x] 11.4 在 `updateInModelSelectState` 中处理 `tea.KeyEsc`：取消选择
- [x] 11.5 在取消后调用 `exitModelSelectMode`，不显示消息
- [x] 11.6 确保 Agent 处理过程中也能打开选择菜单（不阻塞）

## 12. 测试和验证

- [x] 12.1 手动测试：输入 `/model` 打开主模型选择菜单
- [x] 12.2 手动测试：使用上下键导航，按回车确认
- [x] 12.3 手动测试：按 ESC 取消选择
- [x] 12.4 手动测试：输入 `/model :fast` 打开快速模型选择菜单
- [x] 12.5 手动测试：输入 `/model :vision` 打开视觉模型选择菜单
- [x] 12.6 手动测试：输入 `/model ollama/llama3.2` 直接设置主模型
- [x] 12.7 手动测试：输入 `/model :fast ollama/qwen3:14b` 直接设置快速模型
- [x] 12.8 验证配置文件 `~/.nfa/nfa.json` 是否正确保存
- [x] 12.9 验证保存的配置在重启后正确加载
- [x] 12.10 验证 Agent 使用了正确的模型（检查 PromptRequest.Meta）
- [x] 12.11 测试描述超过 80 字符时的截断显示
- [x] 12.12 测试无描述的模型显示
- [x] 12.13 测试无效命令格式的错误提示
- [x] 12.14 测试 Agent 思考过程中切换模型

## 13. 文档和清理

- [x] 13.1 更新 README.md，添加 `/model` 命令使用说明
- [x] 13.2 更新命令帮助信息（如需要）
- [x] 13.3 添加配置文件示例，展示 ModelConfig.Description 字段
- [x] 13.4 代码审查和优化
- [x] 13.5 提交代码到版本控制
