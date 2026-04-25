# 命令行参数 (Command Line)

NFA 提供丰富的命令行参数，方便用户控制程序行为、调试问题和临时切换配置。

## 概述

命令行参数分为两类：

- **全局参数** - 影响整个程序运行的行为（如日志级别、数据目录）
- **运行时参数** - 控制当前会话的配置（如模型选择）

命令行参数优先级高于配置文件，适合临时调整配置而不修改文件。

## 全局参数

### 日志级别控制 (`-v, --verbose`)

控制日志输出的详细程度，帮助调试问题。

**语法**:
```bash
nfa -v [PROMPT]
nfa -vv [PROMPT]
```

**参数说明**:

| 参数 | 日志级别 | 说明 |
|------|----------|------|
| 无参数（默认） | Info | 基础信息输出 |
| `-v` 或 `--verbose=1` | Debug | 详细调试信息 |
| `-vv` 或 `--verbose=2` | Trace | 最详细的追踪信息 |

**使用示例**:

```bash
# 使用默认日志级别
nfa "你好"

# 使用 Debug 级别日志
nfa -v "帮我分析一下股市"

# 使用 Trace 级别日志
nfa -vv "详细分析特斯拉"
```

**日志文件位置**:
所有日志都会输出到 `~/.nfa/nfa.log` 文件中，包括：
- 时间戳
- 日志级别
- 消息内容
- 错误堆栈（如有）

**查看日志**:
```bash
# 实时查看日志
tail -f ~/.nfa/nfa.log

# 查看最近 100 行
tail -n 100 ~/.nfa/nfa.log

# 搜索错误信息
grep ERROR ~/.nfa/nfa.log
```

### 日志配置详情

日志使用 `lumberjack` 进行自动轮转：

- **单个文件最大**: 500 MB
- **保留备份数**: 3 个
- **保留时长**: 28 天

这意味着日志不会无限增长，旧日志会自动清理。

### 数据目录 (`--data-root`)

指定 NFA 的数据存储根目录。

**语法**:
```bash
nfa --data-root /path/to/data [PROMPT]
```

**默认值**: `~/.nfa`

数据目录中包含配置文件、日志文件和技能目录等。自定义数据目录适用于：
- 多实例运行
- 测试环境隔离
- 指定存储位置

### 语言设置 (`--lang`)

设置界面语言。

**语法**:
```bash
nfa --lang zh [PROMPT]
nfa --lang en [PROMPT]
```

**可选值**: `zh`（中文）、`en`（英文）

不设置时自动检测系统语言。也可以在配置文件中通过 `language` 字段永久设置。

## 运行时参数

### 模型选择 (`--model`)

指定当前会话使用的主模型。

**语法**:
```bash
nfa --model MODEL_NAME [PROMPT]
```

**使用示例**:

```bash
# 使用 Deepseek 模型
nfa --model "deepseek/deepseek-v4-pro" "分析一下当前市场"

# 使用 Ollama 模型
nfa --model "ollama/llama2" "介绍一下 ETF"

# 使用智谱模型
nfa --model "z-ai/glm-5.1" "写一个投资策略"
```

**优先级**: 命令行参数 > 配置文件 > 默认值

这个参数非常适合临时测试不同模型，无需修改配置文件。

### 视觉模型选择 (`--vision-model`)

指定当前会话使用的视觉模型。

**语法**:
```bash
nfa --vision-model MODEL_NAME [PROMPT]
```

**使用示例**:

```bash
# 指定视觉模型
nfa --vision-model "qwen/qwen-vl-plus" "分析这张图片"

# 同时指定主模型和视觉模型
nfa --model "deepseek-v4-pro" --vision-model "qwen/qwen-vl-plus" "分析问题"
```

**注意**: 如果未配置视觉模型，系统会自动使用主模型处理视觉任务。

### 恢复会话 (`--resume`)

通过会话 ID 恢复之前的对话。

**语法**:
```bash
nfa --resume SESSION_ID
```

**使用示例**:

```bash
# 恢复之前的会话
nfa --resume "ffffffff-ffff-ffff-ffffffffffff"
```

## 交互式命令

在交互式对话模式中，NFA 提供了命令系统，允许你动态控制程序行为。

### `/clear` - 清空对话上下文

清空当前会话的对话历史，让 Agent "忘记"之前的对话内容。

**语法**:
```bash
/clear
```

**使用场景**:
- 开始新的话题
- 清除错误的上下文
- 重新提问同样的问题

**示例**:
```
用户: 什么是股票？
Agent: [回答]

用户: /clear
# 对话历史被清空

用户: 什么是股票？
Agent: [重新回答，不引用上次的回答]
```

### `/skills` - 列出可用技能

显示当前已加载的所有技能列表。

**语法**:
```bash
/skills
```

**使用场景**:
- 查看有哪些可用的自定义技能
- 确认技能是否正确加载

### `/exit` - 退出程序

退出交互式对话模式。

**语法**:
```bash
/exit
```

**快捷键**:
- `Ctrl+C` - 也可以退出程序（如果在 Agent 处理过程中会先取消当前操作）

### 非交互模式 (`-p, --print`)

打印答案后自动退出，不进入交互式对话模式。

**语法**:
```bash
nfa -p "PROMPT"
```

**使用示例**:

```bash
# 单次问答
nfa -p "什么是市盈率？"

# 结合模型选择
nfa --model "deepseek/deepseek-v4-pro" -p "解释一下量化交易"

# 在脚本中使用
#!/bin/bash
nfa -p "分析 $STOCK 的技术面" > analysis.txt
```

**适用场景**:
- 脚本自动化
- 批量处理
- 单次查询
- 集成到其他系统

## 参数组合使用

命令行参数可以灵活组合：

```bash
# 同时指定主模型和视觉模型，使用调试日志
nfa -v --model "deepseek/deepseek-v4-pro" --vision-model "qwen/qwen3.6-plus" "分析问题"

# 非交互模式 + 指定模型
nfa -p --model "deepseek/deepseek-v4-pro" "单次问题"

# Trace 日志 + 非交互模式
nfa -vv -p "详细调试"
```

## 参数与配置文件的优先级

当同一项配置既在配置文件中定义，又通过命令行参数指定时，优先级如下：

```
命令行参数 > 配置文件 > 默认值
```

**示例**:

配置文件 `~/.nfa/nfa.json`:
```json
{
  "defaultModels": {
    "primary": "ollama/llama2",
    "vision": ""
  }
}
```

运行命令:
```bash
nfa --model "deepseek/deepseek-v4-pro" "问题"
```

结果：当前会话使用 `deepseek/deepseek-v4-pro` 作为主模型，但配置文件未被修改。

这种设计非常适合：
- 临时测试不同模型
- 调试时开启详细日志
- 脚本中使用特定配置

## 帮助信息

NFA 提供完善的帮助系统：

```bash
# 查看主命令帮助
nfa --help

# 查看子命令帮助
nfa models --help
nfa models list --help
```

帮助信息包含：
- 参数说明
- 使用示例
- 默认值
- 相关命令

## 子命令

### `models list` - 列出可用模型

```bash
nfa models list
```

列出当前配置中所有可用的模型。

### `version` - 查看版本信息

```bash
nfa version
nfa version -f json
```

显示当前 NFA 的版本信息。`-f` 参数支持 `json` 格式输出。

## 常见使用场景

### 场景 1：调试问题

当遇到问题时，使用详细日志帮助诊断：

```bash
nfa -vv "复现问题"
```

然后查看日志：
```bash
tail -f ~/.nfa/nfa.log
```

### 场景 2：测试新模型

想测试一个新模型，但不想修改配置文件：

```bash
nfa --model "new-model" "测试一下效果"
```

### 场景 3：脚本自动化

在脚本中使用非交互模式：

```bash
#!/bin/bash
STOCK="AAPL"
nfa -p "分析 $STOCK 的技术指标" > "$STOCK-analysis.txt"
```

### 场景 4：使用视觉模型

需要分析图表或图片时指定视觉模型：

```bash
nfa --model "deepseek/deepseek-v4-pro" --vision-model "qwen/qwen-vl-plus" "分析图表"
```

## 参数验证

NFA 会对命令行参数进行验证：

```bash
# 无效的日志级别
$ nfa -vvv "测试"
Error: invalid log verbosity: 3 (expected: 0, 1 or 2)
```

如果参数验证失败，程序会输出清晰的错误信息并退出。

## 参数参考速查表

### 全局参数

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--verbose` | `-v` | uint32 | 0 | 日志级别 (0/1/2) |
| `--data-root` | - | string | `~/.nfa` | 数据存储根目录 |
| `--lang` | - | string | 自动检测 | 界面语言 (en/zh) |
| `--help` | `-h` | - | - | 显示帮助信息 |
| `--version` | `-V` | - | - | 显示版本信息 |

### 运行时参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--model` | string | 配置文件 | 主模型名称 |
| `--vision-model` | string | 配置文件 | 视觉模型名称 |
| `--print` | `-p` | bool | false | 打印后退出 |
| `--resume` | string | - | 恢复会话 ID |
