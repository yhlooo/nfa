# Skill System

## Purpose

System for discovering, loading, and managing custom skills stored in `~/.nfa/skills/` directories.

## Requirements

### Requirement: Skill directory structure
The system SHALL recognize skills stored in `~/.nfa/skills/<skill-name>/` directories.

#### Scenario: Valid skill directory exists
- **WHEN** a directory `~/.nfa/skills/get-price/` exists
- **THEN** system SHALL identify `get-price` as an available skill

#### Scenario: Non-skill directory ignored
- **WHEN** a file `~/.nfa/skills/test.txt` exists (not a directory)
- **THEN** system SHALL NOT identify it as a skill

### Requirement: SKILL.md file format
Each skill SHALL contain a `SKILL.md` file with YAML frontmatter containing `name` and `description` fields.

#### Scenario: Valid SKILL.md with frontmatter
- **WHEN** `SKILL.md` contains YAML frontmatter with `name: get-price` and `description: Get asset prices`
- **THEN** system SHALL parse both fields successfully

#### Scenario: SKILL.md missing frontmatter
- **WHEN** `SKILL.md` exists without YAML frontmatter
- **THEN** system SHALL log a warning and skip the skill

#### Scenario: SKILL.md missing required fields
- **WHEN** `SKILL.md` frontmatter is missing the `name` field
- **THEN** system SHALL log a warning and skip the skill

### Requirement: Skill content parsing
The system SHALL extract skill content from `SKILL.md`, including both frontmatter and markdown content.

#### Scenario: Parse skill with content
- **WHEN** parsing a skill with frontmatter and markdown content
- **THEN** system SHALL return the full file content

#### Scenario: Parse skill without content
- **WHEN** parsing a skill with only frontmatter and no markdown content
- **THEN** system SHALL return frontmatter content

### Requirement: Skill discovery
The system SHALL scan `~/.nfa/skills/` to discover all available skills at agent initialization.

#### Scenario: Multiple skills available
- **WHEN** `~/.nfa/skills/` contains directories `get-price`, `analyze-trend`, and `send-report`
- **THEN** system SHALL discover all three skills

#### Scenario: Empty skills directory
- **WHEN** `~/.nfa/skills/` exists but contains no valid skill directories
- **THEN** system SHALL return an empty skill list

#### Scenario: Skills directory does not exist
- **WHEN** `~/.nfa/skills/` directory does not exist
- **THEN** system SHALL create the directory and return an empty skill list

### Requirement: Skill metadata retrieval
The system SHALL provide access to skill metadata (name and description) for listing available skills.

#### Scenario: Get skill metadata
- **WHEN** retrieving metadata for `get-price` skill
- **THEN** system SHALL return the name and description from the skill's frontmatter

#### Scenario: Get all skill metadata
- **WHEN** retrieving all skill metadata
- **THEN** system SHALL return a list of all skill names and descriptions

### Requirement: Skill content retrieval
The system SHALL return the full content of a specified skill by name.

#### Scenario: Retrieve existing skill content
- **WHEN** requesting content for skill `get-price`
- **THEN** system SHALL return the full `SKILL.md` content for that skill

#### Scenario: Retrieve non-existent skill
- **WHEN** requesting content for skill `non-existent`
- **THEN** system SHALL return an error indicating the skill was not found

### Requirement: Error handling
The system SHALL handle errors gracefully without crashing the agent.

#### Scenario: Invalid YAML frontmatter
- **WHEN** `SKILL.md` contains malformed YAML frontmatter
- **THEN** system SHALL log a warning and skip the skill

#### Scenario: File permission error
- **WHEN** the system cannot read `SKILL.md` due to permissions
- **THEN** system SHALL log an error and skip the skill
