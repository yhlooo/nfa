## Why

当前系统通过调用供应商 API 动态发现模型列表，这种方式存在两个主要问题：

1. **性能问题**：每次启动都需要发起 HTTP 请求列出模型，耗时较长
2. **元数据缺失**：API 仅返回模型名称，无法获取模型特性（是否支持思考、是否支持视觉、上下文窗口大小、价格等）

这导致系统无法：
- 根据模型能力做智能路由
- 计算 API 调用成本
- 在配置层面了解和约束模型行为

## What Changes

- **新增**：统一的 `ModelConfig` 结构体，包含模型名称和元数据字段
- **新增**：各 Provider (`OllamaOptions`, `DeepseekOptions`, `OpenAICompatibleOptions`) 添加 `Models []ModelConfig` 字段
- **修改**：模型注册逻辑，从"API 发现"改为"配置驱动"
- **BREAKING**：`OllamaOptions.Models` 从 `[]string` 改为 `[]ModelConfig`
- **移除**：未配置 `models` 时的 API 自动发现行为

## Capabilities

### New Capabilities
- `model-config`: 为每个模型供应商提供模型配置能力，支持配置模型名称及扩展元数据（推理能力、视觉能力、价格、上下文窗口等）

### Modified Capabilities
无现有能力变更，这是纯新增功能。

## Impact

**代码影响范围**：
- `pkg/models/providers.go` - 新增 `ModelConfig` 和 `ModelCost` 结构体
- `pkg/models/ollama.go` - `OllamaOptions.Models` 字段类型变更
- `pkg/models/deepseek.go` - `DeepseekOptions` 新增 `Models` 字段
- `pkg/models/openai_compatible.go` - `OpenAICompatibleOptions` 新增 `Models` 字段
- `pkg/agents/genkit.go` - 模型注册逻辑变更，移除 API 发现调用

**配置影响**：
- 用户需要在配置文件中显式配置 `models` 列表
- 未配置 `models` 的 provider 将不会注册任何模型

**文档影响**：
- `docs/guides/model-config.md` 需要更新配置示例
