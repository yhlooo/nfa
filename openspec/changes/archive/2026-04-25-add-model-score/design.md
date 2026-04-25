# 模型效果评分展示 - 设计文档

## Context

当前 `models list` 命令使用 `tablewriter` 渲染 4 列表格：Name | Reasoning | Vision | Context。需要在最后一列增加 Score 评分展示。

`ModelConfig` 结构体当前有 5 个字段：`Name`、`Provider`、`Reasoning`、`Vision`、`ContextWindow`、`Prices`。需要新增 `Score` 字段。

## Goals / Non-Goals

**Goals:**
- `ModelConfig` 新增 `Score` 字段，支持 JSON 序列化/反序列化
- `models list` 表格用 ⭐️✨ 符号展示评分
- 支持 i18n 表头

**Non-Goals:**
- 不参与模型排序或选择逻辑
- 不影响模型注册或 Genkit 集成
- 不在配置文件中为建议模型预设 score 值

## Decisions

### 1. Score 字段设计

**决策：** 使用 `int` 类型，范围 0-10，JSON tag `"score,omitempty"`

```go
// 效果评分，0-10
Score int `json:"score,omitempty"`
```

**理由：**
- `int` 简单直接，无需浮点精度
- `omitempty` 确保零值不序列化，保持向后兼容
- 0-10 是常用的评分范围，易于理解

### 2. 分值映射规则

**决策：** 用 ⭐️ 表示 2 分，✨ 表示 1 分

| 分值 | 展示 |
|------|------|
| 10 | ⭐️⭐️⭐️⭐️⭐️ |
| 9 | ⭐️⭐️⭐️⭐️✨ |
| 8 | ⭐️⭐️⭐️⭐️ |
| 7 | ⭐️⭐️⭐️✨ |
| 6 | ⭐️⭐️⭐️ |
| 5 | ⭐️⭐️✨ |
| 4 | ⭐️⭐️ |
| 3 | ⭐️✨ |
| 2 | ⭐️ |
| 1 | ✨ |
| 0 | （不展示）|

实现：
```go
func scoreToStars(score int) string {
    // clamp to 0-10
    if score < 0 {
        score = 0
    }
    if score > 10 {
        score = 10
    }
    if score == 0 {
        return ""
    }
    return strings.Repeat("⭐️", score/2) + strings.Repeat("✨", score%2)
}
```

**理由：**
- ⭐️ 和 ✨ 是广泛支持的 Unicode 字符
- 用 2 分/1 分拆分可以表示 0-10 所有整数值
- 视觉直观：实心星 > 闪光星，对应 2 > 1

### 3. 表格列布局

**决策：** Score 列放在最后一列，居中对齐

```
| Name | Reasoning | Vision | Context | Score |
|------|-----------|--------|---------|-------|
| ...  | ✅        | ❌     | 128K    | ⭐️⭐️⭐️ |
```

新增对齐数组：`[Left, Center, Center, Right, Center]`

**理由：**
- 评分是附加信息，自然排在最后
- 居中对齐让星标排列更美观

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| 某些终端不支持 ⭐️✨ 字符 | 这两个是 Unicode 标准字符（U+2B50、U+2728），绝大多数现代终端都支持 |
| 零值和新字段混淆 | 零分不展示，与未设置行为一致 |

## Open Questions

无。
