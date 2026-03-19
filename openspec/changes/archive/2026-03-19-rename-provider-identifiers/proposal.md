## Why

当前模型供应商的标识符使用公司名称（aliyun、zhipu），而非用户认知的品牌名称。这导致：

1. **品牌辨识度低**：用户看到 `aliyun/qwen3-max` 时，更熟悉的是"通义千问/Qwen"品牌而非"阿里云"
2. **命名不一致**：其他供应商如 `ollama`、`deepseek` 直接使用品牌名，而阿里云和智谱使用公司名

为提高品牌辨识度和命名一致性，需要将供应商标识符改为品牌名。

## What Changes

- **修改**：`aliyun` → `qwen`（通义千问品牌）
- **修改**：`zhipu` → `z-ai`（智谱 AI 品牌）
- **重命名**：代码文件、类型名、字段名、常量名
- **更新**：文档和活跃 spec 文件
- **BREAKING**：配置文件中的供应商键名和模型前缀需更新

## Capabilities

### Modified Capabilities

- `model-config`: 供应商配置键名从 `aliyun`/`zhipu` 改为 `qwen`/`z-ai`
- `model-selection`: 模型标识格式从 `aliyun/*` 和 `zhipu/*` 改为 `qwen/*` 和 `z-ai/*`

## Impact

**代码影响范围**：
- `pkg/models/providers.go` - 字段名和 JSON tag
- `pkg/models/aliyun_dashscope.go` → `pkg/models/qwen.go` - 文件重命名和内部命名
- `pkg/models/zhipu_bigmodel.go` → `pkg/models/zai.go` - 文件重命名和内部命名
- `pkg/agents/genkit.go` - 字段引用和日志消息

**配置影响**：
- 用户需更新配置文件中的供应商键名
- 模型标识前缀变更（如 `aliyun/qwen3-max` → `qwen/qwen3-max`）

**文档影响**：
- `docs/guides/model-config.md`
- `docs/guides/command-line.md`
- `openspec/specs/model-config/spec.md`
- `openspec/specs/model-selection/spec.md`

**命名映射**：

| 当前 | 改为 |
|------|------|
| `aliyun` (JSON) | `qwen` |
| `zhipu` (JSON) | `z-ai` |
| `Aliyun` (Go field) | `Qwen` |
| `Zhipu` (Go field) | `ZAI` |
| `DashScopeProviderName` | `QwenProviderName` |
| `DashScopeOptions` | `QwenOptions` |
| `DashScopeModels` | `QwenModels` |
| `BigModelProviderName` | `ZAIProviderName` |
| `BigModelOptions` | `ZAIOptions` |
| `BigModelModels` | `ZAIModels` |
| `aliyun_dashscope.go` | `qwen.go` |
| `zhipu_bigmodel.go` | `zai.go` |
