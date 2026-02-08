## ADDED Requirements

### Requirement: Skill tool registration
The system SHALL register a tool named `Skill` that the agent can invoke.

#### Scenario: Skill tool available
- **WHEN** the agent is initialized
- **THEN** the `Skill` tool SHALL be available in the agent's tool registry

### Requirement: Skill tool input schema
The `Skill` tool SHALL accept a JSON parameter with a `name` field.

#### Scenario: Valid input with skill name
- **WHEN** the agent calls the Skill tool with `{"name": "get-price"}`
- **THEN** the tool SHALL process the request

#### Scenario: Missing name parameter
- **WHEN** the agent calls the Skill tool without the `name` parameter
- **THEN** the tool SHALL return an error indicating the missing parameter

### Requirement: Skill tool success response
The `Skill` tool SHALL return a JSON response containing the skill content.

#### Scenario: Return skill content
- **WHEN** the Skill tool is called with a valid skill name
- **THEN** the tool SHALL return `{"content": "<full SKILL.md content>"}`

#### Scenario: Content includes frontmatter
- **WHEN** the skill has YAML frontmatter
- **THEN** the returned content SHALL include the frontmatter section

### Requirement: Skill tool error response
The `Skill` tool SHALL return an error response when the skill is not found.

#### Scenario: Skill not found
- **WHEN** the Skill tool is called with a non-existent skill name
- **THEN** the tool SHALL return `{"error": "Skill '<skill-name>' not found"}`

### Requirement: Skill tool description
The `Skill` tool SHALL have a description that explains its purpose to the agent.

#### Scenario: Tool description available
- **WHEN** the agent queries available tools
- **THEN** the Skill tool SHALL have a description explaining it retrieves skill content by name
