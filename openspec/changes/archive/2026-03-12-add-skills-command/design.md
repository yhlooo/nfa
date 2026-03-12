## Context

**当前状态**
NFA 系统通过 ACP (Agent Client Protocol) 实现 UI 和 Agent 的通信。技能系统已实现：
1. Agent 在 `Initialize` 时加载 builtin 和 local 技能
2. 技能列表存储在 `skillLoader.ListMeta()` 中
3. UI 无法访问技能列表

**架构约束**
- UI 和 Agent 通过 ACP 协议通信，使用 `Meta` 字段传递元数据
- 现有模式：模型列表通过 `InitializeResponse.Meta` 传递
- UI 层使用 Bubble Tea 实现交互式终端界面

## Goals / Non-Goals

**Goals:**
- 在 UI 中展示已加载的技能列表
- 按 builtin/local 分类显示
- 显示技能名称和描述

**Non-Goals:**
- 不添加技能管理功能（添加/删除/编辑）
- 不实现技能详情查看
- 不新增外部依赖

## Decisions

### 1. 技能列表传递：通过 Initialize Meta

**决策**: 在 `InitializeResponse.Meta` 中添加 `skills` 键，传递技能列表。

**理由**:
- **一致性**: 与现有 `availableModels` 模式完全一致
- **简单**: 只需修改一处代码即可传递数据
- **初始化时机**: 技能在 Agent 初始化时已加载完毕

### 2. 显示格式：简洁分组列表

**决策**: 使用简洁的分组列表格式，不带边框。

```
Skills
24 skills

Builtin skills
short-term-trend-forecast - 短期趋势预测技能，分析股票的短期走势

Local skills
get-price - 获取股票实时价格
analyze-trend - 分析股票趋势
```

**理由**:
- **简洁清晰**: 符合终端界面的美学
- **信息充足**: 名称 + 描述足以识别技能
- **分组显示**: 区分 builtin 和 local 技能来源

### 3. 数据类型：直接使用 SkillMeta 切片

**决策**: 直接传递 `[]SkillMeta`，不定义新的传输结构体。

**理由**:
- `SkillMeta` 已包含所有需要的字段（Name, Description, Source）
- JSON 序列化/反序列化天然支持
- 无需额外的类型转换

## Architecture

### 数据流

```
Agent.Initialize()
    │
    ├── skillLoader.LoadMeta() // 已有：加载技能
    │
    └── 返回 Meta:
        ├── availableModels: [...]
        ├── currentModels: {...}
        └── skills: []SkillMeta  // 新增

ChatUI.initAgent()
    │
    └── 接收 Meta:
        ├── ui.curModels = ...
        ├── ui.modelSelector.SetAvailableModels(...)
        └── ui.skills = GetMetaSkillsValue(resp.Meta)  // 新增
```

### 命令处理流程

```
用户输入 /skills
    │
    ▼
ChatUI.Update() 检测到 "/skills"
    │
    ▼
调用 printSkillsList()
    │
    ├── 格式化输出:
    │   ├── 标题: Skills
    │   ├── 总数: N skills
    │   ├── Builtin skills 分组
    │   └── Local skills 分组
    │
    ▼
tea.Printf() 输出到终端
```

## Risks / Trade-offs

### Risk 1: 技能列表过长影响显示

**风险**: 如果技能数量很多，输出可能占据大量屏幕空间

**缓解**:
- 当前技能数量预期较少（<50）
- 用户可以使用 `/clear` 清除输出
- 未来可以考虑分页显示

### Trade-off: 不保存技能列表

**决策**: 不在 ChatUI 初始化后将技能列表保存到配置文件

**理由**: 技能列表是运行时动态加载的，无需持久化

## Open Questions

### Q1: 是否需要显示技能的更多信息（如路径）？

**当前设计**: 仅显示名称和描述

**决策**: 暂不显示，保持简洁

### Q2: 是否支持过滤/搜索技能？

**当前设计**: 不支持

**决策**: 暂不实现，观察实际使用情况
