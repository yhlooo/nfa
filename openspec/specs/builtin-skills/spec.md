# Builtin Skills

## Purpose

系统支持将技能文件通过 Go `embed.FS` 嵌入到二进制中，作为内置技能提供给所有用户，实现开箱即用体验。

## Requirements

### Requirement: 内置技能目录结构
系统 SHALL 将内置技能存储在 `pkg/skills/builtin/<skill-name>/SKILL.md`。

#### Scenario: 有效内置技能目录
- **WHEN** `pkg/skills/builtin/short-term-trend-forecast/SKILL.md` 存在
- **THEN** 系统 SHALL 识别 `short-term-trend-forecast` 为可用内置技能

#### Scenario: 内置技能使用标准 SKILL.md 格式
- **WHEN** 内置技能 SKILL.md 包含 YAML frontmatter
- **THEN** 系统 SHALL 按照 `skill-system` 规范解析 name 和 description 字段

### Requirement: 内置技能嵌入
系统 SHALL 使用 `//go:embed builtin` 将技能目录嵌入到二进制文件中。

#### Scenario: 编译时嵌入技能文件
- **WHEN** 编译程序
- **THEN** 系统 SHALL 将 `builtin/` 目录下的所有技能文件嵌入到二进制中

#### Scenario: 嵌入文件系统可访问
- **WHEN** 程序运行时
- **THEN** 系统 SHALL 通过 `embed.FS` 访问嵌入的技能文件

### Requirement: 技能来源标记
系统 SHALL 在 `SkillMeta` 中添加 `Source` 字段标识技能来源。

#### Scenario: 内置技能来源标记
- **WHEN** 加载内置技能
- **THEN** 系统 SHALL 设置 `SkillMeta.Source` 为 `"builtin"`

#### Scenario: 用户技能来源标记
- **WHEN** 加载用户技能
- **THEN** 系统 SHALL 设置 `SkillMeta.Source` 为 `"local"`

#### Scenario: Source 字段不影响 YAML 解析
- **WHEN** 解析 SKILL.md 文件
- **THEN** 系统 SHALL 不从 YAML frontmatter 读取 `Source` 字段

### Requirement: 两阶段技能加载
系统 SHALL 先加载内置技能，后加载用户技能，同名用户技能覆盖内置技能。

#### Scenario: 仅内置技能
- **WHEN** 仅存在内置技能 `short-term-trend-forecast`
- **THEN** 系统 SHALL 在技能列表中返回该技能，来源标记为 `"builtin"`

#### Scenario: 仅用户技能
- **WHEN** 仅存在用户技能 `my-custom-skill`
- **THEN** 系统 SHALL 在技能列表中返回该技能，来源标记为 `"local"`

#### Scenario: 内置和用户技能共存
- **WHEN** 内置技能 `short-term-trend-forecast` 和用户技能 `my-custom-skill` 都存在
- **THEN** 系统 SHALL 在技能列表中返回两个技能

#### Scenario: 用户技能覆盖同名内置技能
- **WHEN** 内置技能 `short-term-trend-forecast` 和同名用户技能都存在
- **THEN** 系统 SHALL 返回用户技能，来源标记为 `"local"`

### Requirement: 内置技能内容读取
系统 SHALL 能够从 `embed.FS` 读取内置技能内容。

#### Scenario: 读取内置技能
- **WHEN** 请求获取内置技能 `short-term-trend-forecast` 内容
- **THEN** 系统 SHALL 从嵌入文件系统返回完整的 SKILL.md 内容

#### Scenario: 读取用户技能
- **WHEN** 请求获取用户技能 `my-custom-skill` 内容
- **THEN** 系统 SHALL 从文件系统 `~/.nfa/skills/my-custom-skill/SKILL.md` 返回内容

#### Scenario: 统一读取接口
- **WHEN** Agent 调用 `Skill` 工具
- **THEN** 系统 SHALL 不区分技能来源，统一返回技能内容

### Requirement: 内置技能错误处理
系统 SHALL 优雅处理内置技能加载错误。

#### Scenario: 内置技能 YAML 格式错误
- **WHEN** 内置技能 SKILL.md 包含无效 YAML frontmatter
- **THEN** 系统 SHALL 记录警告日志并跳过该技能

#### Scenario: 内置技能缺少必填字段
- **WHEN** 内置技能 SKILL.md 缺少 `name` 或 `description` 字段
- **THEN** 系统 SHALL 记录警告日志并跳过该技能

#### Scenario: 嵌入目录不存在
- **WHEN** 编译时 `builtin/` 目录不存在
- **THEN** 编译 SHALL 失败并提示缺少嵌入目录