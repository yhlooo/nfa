# Skill System Delta

## ADDED Requirements

### Requirement: 技能来源标记
系统 SHALL 在 `SkillMeta` 中添加 `Source` 字段标识技能来源（`builtin` 或 `local`）。

#### Scenario: 内置技能来源标记
- **WHEN** 加载内置技能
- **THEN** 系统 SHALL 设置 `SkillMeta.Source` 为 `"builtin"`

#### Scenario: 用户技能来源标记
- **WHEN** 加载用户技能
- **THEN** 系统 SHALL 设置 `SkillMeta.Source` 为 `"local"`

## MODIFIED Requirements

### Requirement: 技能发现机制
系统 SHALL 扫描多个来源发现所有可用技能：首先从 `embed.FS` 加载内置技能，然后从 `~/.nfa/skills/` 加载用户技能。

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

### Requirement: 技能内容检索
系统 SHALL 支持从不同来源（文件系统或 embed.FS）返回技能内容。

#### Scenario: 检索内置技能内容
- **WHEN** 请求获取内置技能内容
- **THEN** 系统 SHALL 从 `embed.FS` 读取并返回完整内容

#### Scenario: 检索用户技能内容
- **WHEN** 请求获取用户技能内容
- **THEN** 系统 SHALL 从 `~/.nfa/skills/<skill-name>/SKILL.md` 读取并返回完整内容

#### Scenario: 检索不存在的技能
- **WHEN** 请求获取不存在的技能名称
- **THEN** 系统 SHALL 返回错误，指示技能未找到

### Requirement: 技能元数据检索
系统 SHALL 在返回技能元数据时包含 `Source` 字段。

#### Scenario: 获取内置技能元数据
- **WHEN** 检索内置技能元数据
- **THEN** 系统 SHALL 返回包含 `name`、`description` 和 `Source="builtin"` 的元数据

#### Scenario: 获取用户技能元数据
- **WHEN** 检索用户技能元数据
- **THEN** 系统 SHALL 返回包含 `name`、`description` 和 `Source="local"` 的元数据

#### Scenario: 获取所有技能元数据
- **WHEN** 检索所有技能元数据
- **THEN** 系统 SHALL 返回所有技能的元数据列表，每个包含正确的 `Source` 字段

### Requirement: 空技能目录处理
系统 SHALL 在用户技能目录不存在时创建目录，但仅加载内置技能。

#### Scenario: 首次运行无用户目录
- **WHEN** `~/.nfa/skills/` 目录不存在
- **THEN** 系统 SHALL 创建该目录，并仅加载内置技能

#### Scenario: 用户目录为空
- **WHEN** `~/.nfa/skills/` 目录存在但为空
- **THEN** 系统 SHALL 仅加载内置技能
