# models list 命令输出优化

## Why

当前 `models list` 命令仅以简单列表形式输出模型标识符，缺乏对模型类型、提供商和用途的直观展示。用户难以快速了解模型的基本信息，也不清楚哪些是主模型、快速模型或视觉模型。

## What Changes

- 增强模型列表输出格式，显示模型的详细信息
- 为不同用途的模型添加标识符（main/fast/vision）
- 显示模型所属的提供商
- 支持更友好的表格化输出
- 保持向后兼容的简单列表格式选项

## Capabilities

### Modified Capabilities
- `model-selection`: 当前模型选择规格需要更新，添加命令行列表展示的格式要求

### New Capabilities
- 无新增能力，仅为现有功能的体验优化

## Impact

- 受影响的代码：`pkg/commands/models.go` 中的 `newModelsListCommand` 函数
- 可能需要更新：`pkg/models/model_routing.go` 以暴露模型元数据
- 用户体验改进：用户可以更直观地了解可用模型及其用途
- 文档更新：可能需要更新命令行使用指南