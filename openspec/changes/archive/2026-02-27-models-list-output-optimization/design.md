# models list 命令输出优化 - 设计文档

## Context

当前 `models list` 命令的实现非常简单，直接使用 `fmt.Println` 打印 `ModelConfig` 结构体，输出的是 Go 的默认格式化字符串，人类不可读：

```
{deepseek/deepseek-reasoner   true false {0.002 0.003 0} 128000 64000}
```

用户无法：
- 识别哪些是默认配置的 main/fast/vision 模型
- 快速了解模型的能力（推理、视觉）
- 了解模型的上下文窗口大小
- 获取模型的描述信息

项目已经通过 `lipgloss` 引入了 `github.com/muesli/termenv` 依赖，可以复用其实现终端颜色输出。

## Goals / Non-Goals

**Goals:**

- 提供人性化的模型列表输出，使用 emoji 和颜色增强可读性
- 支持通过命令行参数过滤模型（提供商、能力）
- 支持 JSON 输出格式，便于脚本解析
- 显示模型的关键信息：名称、能力、上下文、描述
- 标识默认配置的 main/fast/vision 模型

**Non-Goals:**

- 不实现交互式模型浏览（已有 `/model` 命令）
- 不修改 `ModelConfig` 结构体
- 不改变模型配置文件格式
- 不支持复杂的过滤表达式（如正则表达式）

## Decisions

### 1. 输出格式选择

**决策：** 使用简洁列表格式，而非表格格式

**理由：**
- 简洁列表更适合终端显示，不需要复杂的列对齐
- 更易于复制粘贴模型名称
- 表格在终端宽度不足时换行效果差
- 括号和弱化显示提供足够的视觉区分

**示例：**
```
deepseek/deepseek-reasoner (Reasoning, 128K)         DeepSeek 推理模型
deepseek/deepseek-chat     (128K)                   DeepSeek 对话模型
zhipu/glm-4.7-flashx       (Reasoning, 200K)         智谱 GLM-4.7 Flash
zhipu/glm-4.6v             (Reasoning, Visual, 128K) 智谱视觉多模态
```

### 2. 能力标签方案

**决策：** 使用文字标签表示能力，不使用 emoji

| 能力  | 文字标签   | 显示位置 |
|-------|-----------|----------|
| 推理  | Reasoning | 括号内   |
| 视觉  | Visual    | 括号内   |
| 上下文 | 128K 等   | 括号内   |

**括号内格式：**
- 有能力：`(Reasoning, 128K)` 或 `(Reasoning, Visual, 128K)`
- 无能力：`(128K)`
- 多个能力用 `, ` 分隔

**理由：**
- 文字标签更正式、清晰
- 避免终端 emoji 兼容性问题
- 括号将辅助信息（能力+上下文）与主要内容（名称+描述）分离

### 3. JSON 输出结构

**决策：** 返回包含模型完整信息的 JSON 数组

```json
{
  "models": [
    {
      "name": "deepseek/deepseek-reasoner",
      "provider": "deepseek",
      "description": "DeepSeek 推理模型",
      "capabilities": {
        "reasoning": true,
        "vision": false
      },
      "contextWindow": 128000,
      "maxOutputTokens": 64000,
      "tags": ["main"],
      "cost": {
        "input": 0.002,
        "output": 0.003
      }
    }
  ]
}
```

**理由：**
- 结构化数据便于脚本解析
- 包含所有可用的模型元数据
- `tags` 字段标识模型的用途（main/fast/vision）
- 便于未来扩展其他工具集成

### 4. 过滤参数设计

**决策：** 支持简单的等值过滤，可组合使用

```bash
--provider <name>      # 按提供商过滤
--capability <type>    # 按能力过滤（reasoning|vision）
```

**理由：**
- 简单直观，易于理解
- 满足常见使用场景
- 组合使用提供足够的灵活性
- 避免过度设计（不需要正则或复杂表达式）

### 5. 文本输出布局

**决策：** 使用固定宽度的列布局（3 列）

```
列 1: 模型名称（左对齐，35 字符）
列 2: 括号内容（左对齐，30 字符）
列 3: 描述（左对齐，剩余空间，最大 50 字符）
```

**括号内容格式：**
- 包含能力列表和上下文大小
- 示例：`(Reasoning, 128K)`、`(Reasoning, Visual, 128K)`、`(128K)`
- 使用 `, ` 分隔多个元素

**描述截断规则：**
- 超过 50 字符时截断并添加 `...`
- 如果模型名称超过 35 字符，括号内容换行到下一行
- 没有描述时显示 `-`

**理由：**
- 固定宽度确保对齐整齐
- 截断避免单行过长
- 括号内整合能力+上下文，布局更紧凑

**列对齐实现方案：**

使用手动格式化配合 `github.com/mattn/go-runewidth` 库实现精确的列对齐。

**解决方案：** 使用 `runewidth` 计算实际显示宽度

```go
import "github.com/mattn/go-runewidth"

const (
    col1Width = 35  // 模型名称
    col2Width = 30  // 括号内容
    col3MaxWidth = 50  // 描述最大宽度
)

func renderModelLine(model ModelConfig, styles *outputStyles) string {
    // 列 1: 模型名称（左对齐）
    name := model.Name
    nameWidth := runewidth.StringWidth(name)
    if nameWidth > col1Width {
        name = truncateByWidth(name, col1Width)
        nameWidth = col1Width
    }
    padding := col1Width - nameWidth
    if padding < 0 {
        padding = 0
    }
    col1 := name + strings.Repeat(" ", padding)

    // 列 2: 括号内容（左对齐）
    tag := buildCapabilityTag(model)
    tagWidth := runewidth.StringWidth(tag)
    padding = col2Width - tagWidth
    if padding < 0 {
        padding = 0
    }
    col2 := tag + strings.Repeat(" ", padding)

    // 列 3: 描述（左对齐，截断）
    desc := model.Description
    if desc == "" {
        desc = "-"
    }
    desc = truncateByWidth(desc, col3MaxWidth)

    return col1 + col2 + desc
}

// 按显示宽度截断字符串
func truncateByWidth(s string, max int) string {
    currentWidth := 0
    runes := []rune(s)
    for i, r := range runes {
        w := runewidth.RuneWidth(r)
        if currentWidth + w > max {
            // 添加 "..." 时需要预留 3 个宽度
            if currentWidth + 3 > max {
                return string(runes[:i])
            }
            return string(runes[:i]) + "..."
        }
        currentWidth += w
    }
    return s
}
```

**选择理由：**
- `runewidth` 已在项目依赖中（indirect，通过 bubbletea 引入）
- 专门处理 Unicode 字符的正确显示宽度
- 支持东亚字符（CJK）的正确宽度计算
- 不依赖 `text/tabwriter`，更灵活可控

### 6. 颜色实现方案

**决策：** 仅对括号内容使用弱化显示，其他部分使用默认颜色

```go
import "github.com/muesli/termenv"

// 创建颜色输出
output := termenv.NewOutput(os.Stdout)
profile := output.ColorProfile()

// 定义样式
capabilityTagStyle = output.String().Faint()  // 括号及内容弱化显示
```

**应用：**
- 括号及内容 `(Reasoning, 128K)` 使用弱化/淡色
- 模型名称、描述使用终端默认颜色
- 不使用彩色，保持简洁

**理由：**
- 项目已有依赖（通过 lipgloss 引入）
- 自动检测终端颜色支持（TrueColor、ANSI 256、ASCII）
- 自动处理 `NO_COLOR` 环境变量
- 跨平台兼容（包括 Windows）
- 弱化显示将辅助信息与主要内容区分开来

### 7. 终端能力检测

**决策：** 自动检测终端能力并降级

**检测逻辑：**
1. 检测 `NO_COLOR` 环境变量 → 纯文本输出
2. 检测输出是否为终端（`isatty`） → 非终端使用纯文本
3. 检测 `TERM` 环境变量 → 判断颜色支持级别
4. 自动降级到合适的能力级别

**理由：**
- 确保在所有环境下都能正常工作
- 遵循 `NO_COLOR` 标准
- 避免在文件重定向时包含 ANSI 转义序列

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Emoji 在某些终端显示为方块 | termenv 自动检测，不支持时使用文本标记 |
| 长描述破坏布局 | 实现智能截断和换行逻辑 |
| 颜色在浅色终端不可读 | 使用终端配置的调色板，而非硬编码 RGB |
| JSON 格式未来不稳定 | 保持向后兼容，新字段可选 |
| 过滤参数与未来功能冲突 | 使用简短但语义清晰的参数名 |

### 已知限制

- Windows CMD 不支持 emoji（但 Windows Terminal 和 PowerShell 7+ 支持）
- 非常长的模型名称（> 35 字符）会导致描述换行
- JSON 输出不包含颜色标记（这是设计如此）

## 实现概要

### 核心函数

```go
// 过滤模型列表
func filterModels(models []models.ModelConfig, opts ModelsListOptions) []models.ModelConfig

// 输出人性化文本格式
func outputTextModels(models []models.ModelConfig, w io.Writer) error

// 输出 JSON 格式
func outputJSONModels(models []models.ModelConfig, w io.Writer) error

// 渲染单个模型的文本行
func renderModelLine(model models.ModelConfig, styles *outputStyles) string

// 构建能力标签括号内容
func buildCapabilityTag(model models.ModelConfig) string

// 从模型名称提取提供商
func extractProvider(modelName string) string

// 格式化上下文大小
func formatContextWindow(bytes int64) string

// 创建输出样式（根据终端能力）
func newOutputStyles(output *termenv.Output) *outputStyles
```

### 样式定义

```go
type outputStyles struct {
    capabilityTag termenv.Style  // 括号及内容弱化样式
}
```

### 命令行选项

```go
type ModelsListOptions struct {
    Provider    string   // --provider
    Capability  string   // --capability (reasoning|vision)
    Format      string   // --format (text|json)
}
```

## Open Questions

- 是否需要 `--sort` 参数支持排序？（暂不需要，已按提供商排序）
- 是否需要支持显示价格信息？（暂不需要，JSON 中已包含）