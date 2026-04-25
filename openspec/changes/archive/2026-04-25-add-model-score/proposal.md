# 模型效果评分展示

## Why

当前 `ModelConfig` 没有效果评分的概念，用户无法直观了解每个模型的实际使用效果。为模型添加评分属性和命令行展示，帮助用户在选择模型时参考效果评分。

## What Changes

- `ModelConfig` 结构体新增 `Score` 字段（0-10 的整数评分）
- `models list` 命令表格新增 Score 列，用 ⭐️（2 分）和 ✨（1 分）可视化展示评分

## Capabilities

### Modified Capabilities
- `model-config`: 更新 `ModelConfig` 结构体规格，新增 Score 字段

## Impact

- 受影响的代码：`pkg/models/providers.go`、`pkg/commands/models.go`、`pkg/commands/i18n.go`
- 向后兼容：新字段 `json:"score,omitempty"`，未配置时默认为 0（不展示）
- 纯展示属性，不影响模型选择、排序或其他逻辑
