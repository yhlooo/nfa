## MODIFIED Requirements

### Requirement: Direct model setting via command

用户 MUST 能够通过命令直接指定模型，无需进入选择菜单。

#### Scenario: Set vision model directly
- **WHEN** 用户输入 `/model :vision qwen/qwen3-vl-plus` 并按回车
- **THEN** 系统设置视觉模型为 "qwen/qwen3-vl-plus"
- **AND** 保存配置
- **AND** 显示成功消息
