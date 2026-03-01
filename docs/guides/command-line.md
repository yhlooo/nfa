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

### 轻量模型选择 (`--light-model`)

指定当前会话使用的轻量模型。

**语法**:
```bash
nfa --light-model MODEL_NAME [PROMPT]
```

**使用示例**:

```bash
# 指定轻量模型
nfa --light-model "ollama/mistral" --model "deepseek-chat" "简单介绍一下"

# 只指定轻量模型（未指定主模型时使用配置文件）
nfa --light-model "ollama/mistral" "这是什么？"
```

**注意**: 如果未配置轻量模型，系统会自动使用主模型处理简单任务。

## 交互式命令

在交互式对话模式中，NFA 提供了强大的命令系统，允许你动态控制程序行为。

### `/model` - 模型切换命令

`/model` 命令允许你在运行时动态切换主模型、轻量模型和视觉模型，无需重启程序。

#### 交互式选择模式

**语法**:
```bash
/model              # 打开主模型选择菜单
/model :primary        # 打开主模型选择菜单
/model :light         # 打开轻量模型选择菜单
/model :vision      # 打开视觉模型选择菜单
```

**使用示例**:

```
用户: /model
# 系统显示交互式菜单：
# Select primary model
#
#  ❯ 1. ollama/llama3.2
#    2. ollama/qwen3:14b
#    3. deepseek/deepseek-chat
#
# 使用 ↑↓ 键导航，回车确认，ESC 取消
```

**菜单操作**:
- `↑` / `↓` 或 `Tab` / `Shift+Tab` - 上下移动光标
- `Enter` - 确认选择，保存配置
- `Esc` - 取消选择，返回输入状态

**模型描述**:
如果配置了模型描述信息，菜单会显示每个模型的描述（限制 80 字符，超出使用 "..." 替代）。

#### 直接设置模式

**语法**:
```bash
/model <provider>/<name>              # 直接设置主模型
/model :light <provider>/<name>        # 直接设置轻量模型
/model :vision <provider>/<name>      # 直接设置视觉模型
```

**使用示例**:

```bash
# 设置主模型
/model ollama/llama3.2

# 设置快速模型
/model :light ollama/mistral

# 设置视觉模型
/model :vision deepseek/deepseek-chat

# 同时设置（需要分步执行）
/model :main deepseek/deepseek-chat
/model :light ollama/mistral
```

**模型名称格式**:
模型名称必须使用 `<provider>/<name>` 格式：
- `ollama/llama3.2`
- `deepseek/deepseek-chat`
- `zhipu/glm-4`
- `aliyun/qwen-max`

#### 配置持久化

使用 `/model` 命令切换的模型会立即保存到配置文件 `~/.nfa/nfa.json`：

```json
{
  "defaultModels": {
    "main": "ollama/llama3.2",
    "fast": "ollama/mistral",
    "vision": "deepseek/deepseek-chat"
  }
}
```

**验证配置**:
切换模型后，下次对话会自动使用新模型。你可以通过以下方式验证：
- 查看欢迎屏幕显示的当前模型
- 观察模型响应速度和质量差异
- 使用 `-v` 日志级别查看使用的模型

#### 使用场景

**场景 1：对比不同模型效果**
```bash
# 使用主模型分析
/model ollama/llama3.2
分析一下特斯拉的财报

# 切换到快速模型
/model :light ollama/mistral
分析一下特斯拉的财报

# 对比两次回答的质量和速度
```

**场景 2：根据任务类型切换**
```bash
# 复杂分析使用强大模型
/model deepseek/deepseek-chat
详细分析当前市场趋势

# 简单查询使用快速模型
/model :light ollama/mistral
什么是市盈率？
```

**场景 3：成本控制**
```bash
# 使用本地模型
/model ollama/llama3.2

# 需要更高质量时切换到云端模型
/model deepseek/deepseek-chat
```

#### 注意事项

1. **模型可用性**:
   - 确保模型已在配置文件中定义
   - Ollama 模型需要提前下载
   - 云端模型需要有效的 API 密钥

2. **Agent 处理过程中切换**:
   - 可以在 Agent 思考或输出过程中打开选择菜单
   - 切换会在下次对话时生效
   - 不影响当前正在进行的对话

3. **配置文件验证**:
   - 直接设置模型时不验证模型是否存在
   - 如果模型不存在，下次对话时会报错
   - 使用交互式菜单可以查看所有可用模型

4. **与命令行参数的区别**:
   - `/model`: 持久化到配置文件，影响所有会话
   - `--model`: 仅影响当前会话，不修改配置文件

### 其他内置命令

#### `/clear` - 清空对话上下文

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

#### `/exit` - 退出程序

退出交互式对话模式。

**语法**:
```bash
/exit
```

**快捷键**:
- `Ctrl+C` - 也可以退出程序（如果在 Agent 处理过程中会先取消当前操作）

#### `/summarize` - 生成对话摘要

生成当前对话历史的结构化摘要。

**语法**:
```bash
/summary
# 或
/summarize
```

**输出内容**:
- 对话标题
- 对话描述
- 过程概述
- 方法论总结（如有）

**使用场景**:
- 长对话后快速回顾
- 保存对话要点
- 整理讨论内容

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
nfa --light-model "ollama/mistral" --model "deepseek-chat" "分析问题"
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
alias nfa-quick='nfa --light-model ollama/mistral'
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
| `--light-model` | - | string | 配置文件 | 轻量模型名称 |
| `--print` | `-p` | bool | false | 打印后退出 |
| `--help` | `-h` | - | - | 显示帮助信息 |
| `--version` | `-V` | - | - | 显示版本信息 |
