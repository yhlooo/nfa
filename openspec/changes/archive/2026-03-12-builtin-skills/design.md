## Context

当前技能系统 (`pkg/skills/`) 仅支持从用户目录 `~/.nfa/skills/` 加载技能。用户首次使用时需要手动创建技能目录和文件，降低了开箱即用体验。

项目已有类似的 embed 实现参考：`pkg/i18n/locales.go` 使用 `embed.FS` 内置多语言文件。

## Goals / Non-Goals

**Goals:**
- 提供开箱即用的默认技能，无需用户手动创建
- 保持用户扩展能力，用户技能可覆盖内置技能
- Agent 调用技能时无需区分来源，统一接口访问
- 内置技能跟随程序版本发布，无需额外版本管理

**Non-Goals:**
- 不实现技能继承机制（用户技能扩展内置技能）
- 不提供技能管理 CLI 命令
- 不在首次运行时自动复制内置技能到用户目录
- 不实现技能版本管理或更新通知

## Decisions

### 1. 两阶段加载顺序

**决策：** 先加载内置技能（embed），后加载用户技能（`~/.nfa/skills/`）

**理由：**
- 确保用户技能优先级更高，同名用户技能可覆盖内置技能
- 符合"本地配置优先"的惯例
- 简化实现逻辑，避免冲突解决复杂性

**替代方案：** 先加载用户技能，后加载内置技能（仅补充不存在的技能）
- 问题：无法实现覆盖行为，用户修改内置技能后需要改名

### 2. embed 目录结构

**决策：** 在 `pkg/skills/builtin/` 目录下存放内置技能，使用 `//go:embed builtin` 声明

```
pkg/skills/
├── builtin/
│   └── short-term-trend-forecast/
│       └── SKILL.md
└── builtin.go  // 声明 var builtinSkillFS embed.FS
```

**理由：**
- 集中管理内置技能，便于维护
- 与现有 `pkg/i18n/locales.go` 模式一致
- 编译时 `//go:embed` 确保文件存在

**替代方案：** 在项目根目录创建 `skills/` 目录
- 问题：与代码目录分离，降低可维护性

### 3. 虚拟路径存储

**决策：** 内置技能在 `SkillRef` 中存储虚拟路径（如 `builtin/short-term-trend-forecast`）

**理由：**
- 统一接口，`Get()` 方法根据路径判断读取方式
- 便于调试和日志记录

### 4. SkillMeta 扩展

**决策：** 在 `SkillMeta` 结构体中添加 `Source` 字段（`yaml:"-"` 表示不从 YAML 解析）

```go
type SkillMeta struct {
    Name        string `yaml:"name"`
    Description string `yaml:"description"`
    Source      string `yaml:"-"` // "builtin" 或 "local"
}
```

**理由：**
- 标记技能来源，便于调试和未来扩展
- `yaml:"-"` 避免影响现有技能文件格式
- 对用户透明，Agent 不需要感知此字段

### 5. 常量定义

**决策：** 定义常量表示技能来源

```go
const (
    SkillSourceBuiltin = "builtin"
    SkillSourceLocal   = "local"
)
```

**理由：**
- 避免硬编码字符串，提高类型安全
- 便于统一修改和测试

### 6. 错误处理策略

**决策：** 内置技能解析失败时记录警告并跳过，不中断程序启动

**理由：**
- 与现有用户技能加载行为一致
- 避免因单个内置技能错误导致整个程序无法启动
- 提供更好的容错性

## Risks / Trade-offs

### 1. 二进制大小增加

**风险：** 内置技能会增加到编译产物大小

**缓解：**
- 当前仅内置一个技能（`short-term-trend-forecast`），影响有限
- 未来可通过配置控制是否包含内置技能

### 2. 内置技能更新需要重新编译

**风险：** 修复内置技能 bug 或添加新技能需要重新发布二进制

**缓解：**
- 内置技能作为"官方推荐"，用户可复制到本地修改
- 未来可考虑支持从远程 URL 动态加载技能（超出当前范围）

### 3. 技能覆盖可能引起混淆

**风险：** 用户创建同名技能后，可能不知道内置技能被覆盖

**缓解：**
- 当前对用户透明，不区分来源
- 未来可在日志中记录技能加载状态和覆盖行为

### 4. 测试复杂性

**风险：** 需要同时测试内置和用户技能加载逻辑

**缓解：**
- 复用现有测试框架和模式
- 新增测试用例覆盖混合加载场景

## Migration Plan

### 部署步骤

1. 创建 `pkg/skills/builtin/` 目录
2. 将现有 `~/.nfa/skills/short-term-trend-forecast/SKILL.md` 复制到 `builtin/` 目录
3. 创建 `pkg/skills/builtin.go` 文件
4. 修改 `pkg/skills/skill_loader.go` 和 `pkg/skills/skill_parser.go`
5. 运行测试验证功能
6. 编译并发布新版本

### 回滚策略

- 如出现严重问题，可回退到仅支持用户技能的版本
- 用户技能不受影响，继续正常工作

## Open Questions

1. **技能列表顺序：** `ListMeta()` 返回的技能列表应该如何排序？
   - **建议：** 先内置后用户，保持加载顺序

2. **是否需要验证内置技能：** 是否需要确保内置技能名称不以特定前缀开头（如 `builtin:`）？
   - **建议：** 不需要，保持与用户技能一致的命名规则

3. **嵌入文件编码：** 是否需要处理 BOM 或特殊编码？
   - **建议：** 当前 `short-term-trend-forecast` 为 UTF-8，暂时不需要
