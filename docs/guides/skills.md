# 自定义技能 (Custom Skills)

NFA 支持通过自定义技能扩展能力。技能是用户定义的功能，存储在 `~/.nfa/skills/` 目录中。

## 技能目录结构

每个技能是一个包含 `SKILL.md` 文件的目录：

```
~/.nfa/skills/
├── get-price/
│   └── SKILL.md
├── analyze-trend/
│   └── SKILL.md
└── send-report/
    └── SKILL.md
```

## SKILL.md 格式

`SKILL.md` 文件使用 YAML frontmatter 定义元数据，后面跟着具体内容：

```markdown
---
name: get-price
description: 获取资产价格
---

1. 首先确认资产正确代码，必要时可通过搜索引擎搜索
2. 通过代码查询资产近 5 个交易日的价格
3. 返回价格数据，包括日期和收盘价
```

## 必填字段

- `name`: 技能名称（在 `Skill` 工具中使用的名称）
- `description`: 技能描述（向 agent 说明技能的用途）

## 使用技能

Agent 会自动加载 `~/.nfa/skills/` 中的所有技能，并在系统提示中列出可用技能。当需要使用某个技能时，agent 会调用 `Skill` 工具并传入技能名称。

### 使用示例

假设你创建了以下技能：

```
~/.nfa/skills/
└── get-price/
    └── SKILL.md
```

当用户问"查询 AAPL 的股价"时，agent 会：

1. 识别需要使用 `get-price` 技能
2. 调用 `Skill` 工具，传入 `{"name": "get-price"}`
3. 获取技能内容并按照其中的步骤执行

### Skill 工具调用示例

```json
{
  "name": "get-price"
}
```

返回示例：

```json
{
  "content": "---\nname: get-price\ndescription: 获取资产价格\n---\n\n1. 首先确认资产正确代码，必要时可通过搜索引擎搜索\n2. 通过代码查询资产近 5 个交易日的价格\n3. 返回价格数据，包括日期、开盘价、收盘价等关键信息\n"
}
```

### 通过斜杠命令调用技能

在交互式对话中，可以通过 `/` 前缀直接调用技能：

```
/get-price AAPL
```

这会直接使用 `get-price` 技能处理请求。

### 查看已加载技能

使用 `/skills` 命令可以查看当前已加载的所有技能列表：

```
/skills
```

## 内置技能

NFA 包含一些内置技能，它们随程序发布，无需用户手动创建。当前内置技能包括：

- **short-term-trend-forecast** - 短期趋势预测

内置技能的描述会根据当前语言设置自动翻译。

### 覆盖内置技能

如果你在 `~/.nfa/skills/` 目录中创建了与内置技能同名的技能，你的自定义技能会覆盖内置技能。这允许你自定义或改进内置技能的行为。

## 示例

参考 [examples/skills/](../../examples/skills/) 目录中的示例技能。
