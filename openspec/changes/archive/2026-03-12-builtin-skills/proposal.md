## Why

当前技能系统要求用户在 `~/.nfa/skills/` 目录手动创建技能文件。新用户首次使用时没有任何可用技能，需要手动创建或复制技能文件，降低了开箱即用体验。同时，官方希望提供一些核心技能（如 `short-term-trend-forecast`），让所有用户都能直接使用。

## What Changes

- **新增内置技能支持**：通过 Go `embed.FS` 将技能文件嵌入到二进制中，提供开箱即用的默认技能
- **混合加载机制**：优先加载内置技能（embed），然后加载用户技能（`~/.nfa/skills/`），同名用户技能覆盖内置技能
- **技能来源标记**：在技能元数据中添加 `Source` 字段（`builtin` 或 `local`），用于标识技能来源，但对用户透明
- **统一访问接口**：Agent 调用 `Skill` 工具时不区分技能来源，一视同仁

## Capabilities

### New Capabilities
- `builtin-skills`: 支持通过 embed 方式将技能嵌入到二进制文件中，作为内置技能提供给用户

### Modified Capabilities
- `skill-system`: 扩展技能发现机制，支持从 embed.FS 加载内置技能，并实现内置技能和用户技能的合并与覆盖逻辑
- `skill-tool`: 更新工具描述，说明技能可来自内置或用户定义位置

## Impact

- **pkg/skills/skill_loader.go**：新增 `loadBuiltinSkills()` 方法，修改 `LoadMeta()` 和 `Get()` 方法支持混合加载
- **pkg/skills/skill_parser.go**：扩展 `SkillMeta` 结构体，添加 `Source` 字段
- **pkg/skills/builtin/**：新增目录，用于存放内置技能文件
- **pkg/skills/builtin.go**：新增文件，声明 embed.FS 并实现内置技能加载逻辑
- **pkg/skills/skill_tool.go**：更新工具描述文本
- **单元测试**：新增内置技能加载、用户覆盖内置技能等测试用例
- **二进制大小**：内置技能会增加编译产物大小
