## ADDED Requirements

### Requirement: Model score field
`ModelConfig` 结构体 SHALL 包含 `Score` 字段，用于表示模型效果评分（0-10 的整数）。

#### Scenario: Score field in JSON
- **WHEN** 模型配置包含 `"score": 8`
- **THEN** 系统 SHALL 解析为 `Score = 8`

#### Scenario: Score field omitted
- **WHEN** 模型配置不包含 `score` 字段
- **THEN** 系统 SHALL 使用零值 `Score = 0`

#### Scenario: Score display mapping
- **WHEN** Score 值在 `models list` 中展示
- **THEN** 系统 SHALL 使用 ⭐️（2 分）和 ✨（1 分）符号展示
- **AND** Score <= 0 时不展示任何符号
- **AND** Score > 10 时按 10 处理
