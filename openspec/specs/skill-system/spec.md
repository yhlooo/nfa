# Skill System

## Purpose

System for discovering, loading, and managing skills from both built-in (embedded) and user-defined (`~/.nfa/skills/`) locations.

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

### Requirement: 技能来源标记
系统 SHALL 在 `SkillMeta` 中添加 `Source` 字段标识技能来源（`builtin` 或 `local`）。

#### Scenario: 内置技能来源标记
- **WHEN** 加载内置技能
- **THEN** 系统 SHALL 设置 `SkillMeta.Source` 为 `"builtin"`

#### Scenario: 用户技能来源标记
- **WHEN** 加载用户技能
- **THEN** 系统 SHALL 设置 `SkillMeta.Source` 为 `"local"`

### Requirement: Skill discovery
The system SHALL scan multiple sources to discover all available skills: first loading built-in skills from `embed.FS`, then loading user skills from `~/.nfa/skills/`.

#### Scenario: Multiple skills available
- **WHEN** `~/.nfa/skills/` contains directories `get-price`, `analyze-trend`, and `send-report`
- **THEN** system SHALL discover all three skills (plus any built-in skills)

#### Scenario: 仅内置技能可用
- **WHEN** 程序首次运行，无用户技能目录
- **THEN** 系统 SHALL 发现并加载所有内置技能

#### Scenario: 内置和用户技能共存
- **WHEN** 内置技能和用户技能都存在
- **THEN** 系统 SHALL 先加载所有内置技能，再加载所有用户技能

#### Scenario: 用户技能覆盖同名内置技能
- **WHEN** 用户创建与内置技能同名的技能
- **THEN** 系统 SHALL 在合并时用用户技能覆盖内置技能

#### Scenario: 技能列表合并
- **WHEN** 检索所有可用技能
- **THEN** 系统 SHALL 返回内置技能和用户技能的合并列表

#### Scenario: Empty skills directory
- **WHEN** `~/.nfa/skills/` exists but contains no valid skill directories
- **THEN** system SHALL return only built-in skills

#### Scenario: Skills directory does not exist
- **WHEN** `~/.nfa/skills/` directory does not exist
- **THEN** system SHALL create the directory and return only built-in skills

### Requirement: Skill metadata retrieval
The system SHALL provide access to skill metadata (name, description, and source) for listing available skills.

#### Scenario: Get skill metadata
- **WHEN** retrieving metadata for `get-price` skill
- **THEN** system SHALL return the name, description, and source from the skill's frontmatter

#### Scenario: 获取内置技能元数据
- **WHEN** 检索内置技能元数据
- **THEN** 系统 SHALL 返回包含 `name`、`description` 和 `Source="builtin"` 的元数据

#### Scenario: 获取用户技能元数据
- **WHEN** 检索用户技能元数据
- **THEN** 系统 SHALL 返回包含 `name`、`description` 和 `Source="local"` 的元数据

#### Scenario: 获取所有技能元数据
- **WHEN** 检索所有技能元数据
- **THEN** 系统 SHALL 返回所有技能的元数据列表，每个包含正确的 `Source` 字段

#### Scenario: Get all skill metadata
- **WHEN** retrieving all skill metadata
- **THEN** system SHALL return a list of all skill names, descriptions, and sources

### Requirement: Skill content retrieval
The system SHALL return the full content of a specified skill by name, supporting retrieval from different sources (filesystem or embed.FS).

#### Scenario: Retrieve existing skill content
- **WHEN** requesting content for skill `get-price`
- **THEN** system SHALL return the full `SKILL.md` content for that skill

#### Scenario: 检索内置技能内容
- **WHEN** 请求获取内置技能内容
- **THEN** 系统 SHALL 从 `embed.FS` 读取并返回完整内容

#### Scenario: 检索用户技能内容
- **WHEN** 请求获取用户技能内容
- **THEN** 系统 SHALL 从 `~/.nfa/skills/<skill-name>/SKILL.md` 读取并返回完整内容

#### Scenario: Retrieve non-existent skill
- **WHEN** requesting content for skill `non-existent`
- **THEN** system SHALL return an error indicating the skill was not found

### Requirement: 空技能目录处理
系统 SHALL 在用户技能目录不存在时创建目录，但仅加载内置技能。

#### Scenario: 首次运行无用户目录
- **WHEN** `~/.nfa/skills/` 目录不存在
- **THEN** 系统 SHALL 创建该目录，并仅加载内置技能

#### Scenario: 用户目录为空
- **WHEN** `~/.nfa/skills/` 目录存在但为空
- **THEN** 系统 SHALL 仅加载内置技能

### Requirement: Error handling
The system SHALL handle errors gracefully without crashing the agent.

#### Scenario: Invalid YAML frontmatter
- **WHEN** `SKILL.md` contains malformed YAML frontmatter
- **THEN** system SHALL log a warning and skip the skill

#### Scenario: File permission error
- **WHEN** the system cannot read `SKILL.md` due to permissions
- **THEN** system SHALL log an error and skip the skill