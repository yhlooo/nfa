# Skill Tool

## Purpose

Tool interface that allows the agent to retrieve and execute skills from both built-in and user-defined locations.

## Requirements

### Requirement: Skill tool registration
The system SHALL register a tool named `Skill` that the agent can invoke.

#### Scenario: Skill tool available
- **WHEN** the agent is initialized
- **THEN** the `Skill` tool SHALL be available in the agent's tool registry

### Requirement: Skill tool input schema
The `Skill` tool SHALL accept a JSON parameter with a `name` field.

#### Scenario: Valid input with skill name
- **WHEN** the agent calls the Skill tool with `{"name": "get-price"}`
- **THEN** tool SHALL process the request

#### Scenario: Missing name parameter
- **WHEN** the agent calls the Skill tool without the `name` parameter
- **THEN** tool SHALL return an error indicating the missing parameter

### Requirement: Skill tool success response
The `Skill` tool SHALL return a JSON response containing skill content.

#### Scenario: Return skill content
- **WHEN** the Skill tool is called with a valid skill name
- **THEN** tool SHALL return `{"content": "<full SKILL.md content>"}`

#### Scenario: Content includes frontmatter
- **WHEN** the skill has YAML frontmatter
- **THEN** the returned content SHALL include the frontmatter section

### Requirement: Skill tool error response
The `Skill` tool SHALL return an error response when the skill is not found.

#### Scenario: Skill not found
- **WHEN** the Skill tool is called with a non-existent skill name
- **THEN** tool SHALL return `{"error": "Skill '<skill-name>' not found"}`

### Requirement: Skill tool 描述
`Skill` 工具 SHALL 更新描述以说明技能可来自内置或用户定义位置。

#### Scenario: Tool description available
- **WHEN** the agent queries available tools
- **THEN** the Skill tool SHALL have a description explaining it retrieves skill content by name

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