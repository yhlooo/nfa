# 命令行参数 (Command Line)

NFA 提供丰富的命令行参数，方便用户控制程序行为、调试问题和临时切换配置。

## 概述

命令行参数分为两类：

- **全局参数** - 影响整个程序运行的行为（如日志级别）
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
nfa --model "deepseek-chat" "分析一下当前市场"

# 使用 Ollama 模型
nfa --model "ollama/llama2" "介绍一下 ETF"

# 使用自定义提供商的模型
nfa --model "aliyun/qwen-max" "写一个投资策略"
```

**优先级**: 命令行参数 > 配置文件 > 默认值

这个参数非常适合临时测试不同模型，无需修改配置文件。

### 快速模型选择 (`--fast-model`)

指定当前会话使用的快速模型。

**语法**:
```bash
nfa --fast-model MODEL_NAME [PROMPT]
```

**使用示例**:

```bash
# 指定快速模型
nfa --fast-model "ollama/mistral" --model "deepseek-chat" "简单介绍一下"

# 只指定快速模型（未指定主模型时使用配置文件）
nfa --fast-model "ollama/mistral" "这是什么？"
```

**注意**: 如果未配置快速模型，系统会自动使用主模型处理快速任务。

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
nfa --model "deepseek-chat" -p "解释一下量化交易"

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
# 同时指定主模型和快速模型，使用调试日志
nfa -v --model "deepseek-chat" --fast-model "ollama/mistral" "分析问题"

# 非交互模式 + 指定模型
nfa -p --model "deepseek-chat" "单次问题"

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
    "main": "ollama/llama2",
    "fast": "ollama/llama2"
  }
}
```

运行命令:
```bash
nfa --model "deepseek-chat" "问题"
```

结果：当前会话使用 `deepseek-chat` 作为主模型，但配置文件未被修改。

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

### 场景 4：成本控制

使用本地模型作为快速模型，云端模型作为主模型：

```bash
nfa --fast-model "ollama/mistral" --model "deepseek-chat" "分析问题"
```

这样可以降低 API 调用成本。

## 参数验证

NFA 会对命令行参数进行验证：

```bash
# 无效的日志级别
$ nfa -vvv "测试"
Error: invalid log verbosity: 3 (expected: 0, 1 or 2)

# 无效的模型名称
$ nfa --model "invalid-model" "测试"
Error: model "invalid-model" not found
```

如果参数验证失败，程序会输出清晰的错误信息并退出。

## 配置文件位置

虽然命令行参数优先级更高，但了解配置文件位置也很重要：

- **配置文件**: `~/.nfa/nfa.json`
- **日志文件**: `~/.nfa/nfa.log`
- **技能目录**: `~/.nfa/skills/`
- **数据目录**: `~/.nfa/`

配置文件会在首次运行时自动创建，如果不存在。

## 高级技巧

### 1. 使用别名简化常用命令

在 `.bashrc` 或 `.zshrc` 中添加别名：

```bash
alias nfa-debug='nfa -vv'
alias nfa-quick='nfa --fast-model ollama/mistral'
alias nfa-single='nfa -p'
```

使用：
```bash
nfa-debug "调试问题"
nfa-quick "快速查询"
nfa-single "单次问题"
```

### 2. 环境变量存储密钥

可以使用环境变量存储 API 密钥（需配合配置管理）：

```bash
export NFA_OPENAI_KEY="your-api-key"
```

然后在配置文件中使用占位符替换。

### 3. Shell 脚本批量处理

```bash
#!/bin/bash
for stock in AAPL GOOGL MSFT; do
  nfa -p "分析 $stock 的走势" >> analysis.txt
  echo "---" >> analysis.txt
done
```

## 故障排查

### 日志级别不生效

确保使用正确的参数：
- `-v` 或 `-vv` 是简写
- `--verbose=1` 或 `--verbose=2` 是完整形式

### 模型未找到

1. 检查模型名称是否正确
2. 确认模型已下载（Ollama）
3. 确认 API 密钥有效（云端模型）

### 参数冲突

不同参数之间不会冲突，但如果配置不合理可能导致意外行为：
- 使用 `-p` 时不应期望进入交互模式
- 未配置 vision 模型时使用 WebBrowse 可能受限

## 参数参考速查表

| 参数 | 简写 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--verbose` | `-v` | uint32 | 0 | 日志级别 (0/1/2) |
| `--model` | - | string | 配置文件 | 主模型名称 |
| `--fast-model` | - | string | 配置文件 | 快速模型名称 |
| `--print` | `-p` | bool | false | 打印后退出 |
| `--help` | `-h` | - | - | 显示帮助信息 |
| `--version` | `-V` | - | - | 显示版本信息 |
