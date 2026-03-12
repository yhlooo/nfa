# Skill Tool Delta

## MODIFIED Requirements

### Requirement: Skill tool 描述
`Skill` 工具 SHALL 更新描述以说明技能可来自内置或用户定义位置。

#### Scenario: 工具描述包含内置技能信息
- **WHEN** Agent 查询可用工具
- **THEN** `Skill` 工具描述 SHALL 提及技能来自内置（通过 embed）或用户定义（`~/.nfa/skills/`）位置

#### Scenario: 工具描述不影响输入输出格式
- **WHEN** Agent 调用 `Skill` 工具
- **THEN** 输入输出格式 SHALL 保持不变，仅描述文本更新

### Requirement: Skill tool 统一访问
`Skill` 工具 SHALL 不区分技能来源，对所有技能一视同仁。

#### Scenario: 调用内置技能
- **WHEN** Agent 调用 `Skill` 工具请求内置技能名称
- **THEN** 工具 SHALL 返回该技能内容，无需特殊处理

#### Scenario: 调用用户技能
- **WHEN** Agent 调用 `Skill` 工具请求用户技能名称
- **THEN** 工具 SHALL 返回该技能内容，无需特殊处理

#### Scenario: 工具响应不包含来源信息
- **WHEN** `Skill` 工具返回技能内容
- **THEN** 响应 SHALL 不包含 `Source` 字段，仅返回 `name`、`description` 和 `content`
