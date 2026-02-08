## ADDED Requirements

### Requirement: System prompt includes available skills
The agent's system prompt SHALL include a list of available skills with their names and descriptions.

#### Scenario: System prompt with skills
- **WHEN** the agent is initialized with skills `get-price` and `analyze-trend`
- **THEN** the system prompt SHALL contain a section listing these skills with their names and descriptions

#### Scenario: System prompt without skills
- **WHEN** the agent is initialized with no skills available
- **THEN** the system prompt SHALL indicate that no custom skills are available

### Requirement: System prompt skill format
The system prompt SHALL format available skills in a clear, readable format for the agent.

#### Scenario: Single skill format
- **WHEN** there is one skill `get-price` with description "Get asset prices"
- **THEN** the system prompt SHALL include "Available skills: get-price - Get asset prices"

#### Scenario: Multiple skills format
- **WHEN** there are multiple skills
- **THEN** the system prompt SHALL list each skill on its own line or in a comma-separated format

### Requirement: Skill tool registration
The agent SHALL register the `Skill` tool during initialization.

#### Scenario: Skill tool registered
- **WHEN** the agent is initialized
- **THEN** the `Skill` tool SHALL be available for the agent to call

### Requirement: Skill loading during initialization
The agent SHALL load all available skills from `~/.nfa/skills/` during initialization.

#### Scenario: Load skills on startup
- **WHEN** the agent starts
- **THEN** the system SHALL scan `~/.nfa/skills/` and load all valid skills

#### Scenario: Handle load errors gracefully
- **WHEN** there are errors loading some skills
- **THEN** the agent SHALL log warnings and continue with the successfully loaded skills
